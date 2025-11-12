package handler

import (
	"context"
	"encoding/hex"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/alfianX/danus-h2h/internal/repo"
	f "github.com/alfianX/danus-h2h/pkg/function"
	"github.com/alfianX/danus-h2h/pkg/iso"
	"github.com/alfianX/danus-h2h/pkg/license"
	"github.com/moov-io/iso8583"
)

type errorMessage struct {
	RC  string
	Err error
}

func (h *Handler) ClientHandler(conn net.Conn, sem chan struct{}, wg *sync.WaitGroup) {
	defer func() {
		wg.Done() // Memberi tahu WaitGroup bahwa goroutine selesai
		<-sem     // Melepaskan semaphore
	}()

	timeoutTime, _ := strconv.Atoi(h.Config.TimeoutInactivity)
	conn.SetReadDeadline(time.Now().Add(time.Duration(timeoutTime) * time.Second))

	for {
		header := make([]byte, HeaderLen)
		_, err := io.ReadFull(conn, header)
		if err != nil {
			if err == io.EOF {
				return // Keluar dari loop dan defer akan menutup koneksi jika belum
			}
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				// Mungkin perlu menutup koneksi di sini jika timeout artinya koneksi tidak bisa dipakai lagi
				conn.Close()
				return
			}
			if opErr, ok := err.(*net.OpError); ok {
				// Penanganan error khusus net.OpError, misalnya jika koneksi tertutup secara paksa
				if strings.Contains(opErr.Err.Error(), "connection reset by peer") ||
					strings.Contains(opErr.Err.Error(), "forcibly closed by the remote host") {
					// Koneksi sudah tidak valid, cukup keluar
					return
				}
			}
			// Error lainnya yang tidak diharapkan
			h.Log.Errorf("client handler -> unhandled read error from client %s: %v", conn.RemoteAddr(), err)
			conn.Close() // Pastikan ditutup jika error tidak di-handle secara spesifik
			return
		}

		var messageBytes []byte
		var message []byte

		headerStr := hex.EncodeToString(header)
		msgLength, err := strconv.ParseInt(headerStr, HexBase, IntBitSize)
		if err != nil {
			h.handleErrorAndRespond(conn, "", RCErrFormatError, "client handler - parse int msg length:", err)
			return
		}

		if msgLength <= 0 || msgLength > MaxMessageLength {
			h.handleErrorAndRespond(conn, "", RCErrFormatError, "client handler - invalid message length", nil)
			return
		}

		messageBytes = make([]byte, msgLength)
		_, err = io.ReadFull(conn, messageBytes)
		if err != nil {
			if err == io.EOF {
				return // Keluar dari loop dan defer akan menutup koneksi jika belum
			}
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				// Mungkin perlu menutup koneksi di sini jika timeout artinya koneksi tidak bisa dipakai lagi
				conn.Close()
				return
			}
			if opErr, ok := err.(*net.OpError); ok {
				// Penanganan error khusus net.OpError, misalnya jika koneksi tertutup secara paksa
				if strings.Contains(opErr.Err.Error(), "connection reset by peer") ||
					strings.Contains(opErr.Err.Error(), "forcibly closed by the remote host") {
					// Koneksi sudah tidak valid, cukup keluar
					return
				}
			}

			h.handleErrorAndRespond(conn, "", RCErrGeneral, "client handler - read from client:", err)
			conn.Close()
			return
		}

		message = append(header, messageBytes...)
		isoRequestString := strings.ToUpper(hex.EncodeToString(message))

		conn.SetReadDeadline(time.Now().Add(time.Duration(timeoutTime) * time.Second))

		license := license.CheckLicense(os.Getenv("LICENSE_KEY"), h.volumesn)
		if license != "OK" {
			h.handleErrorAndRespond(conn, "", RCErrLicense, "client handler - license:", err)
			return
		}

		h.Log.WithField("debug_tag", "dl_in").Debugf("message request [%s]: %s", conn.RemoteAddr().String(), isoRequestString)

		TP := isoRequestString[4:6]

		if TP == "60" {
			h.tpduConn.Store(conn, isoRequestString[4:14])
			isoSend, idTrx, direction, err := h.clientPrepare(message[7:])
			if err.Err != nil {
				h.handleErrorAndRespond(conn, isoRequestString[14:], err.RC, "client handler - ", err.Err)
				return
			}

			if direction == 0 {
				go h.sendSingleHostHandler(conn, isoSend, idTrx)
			} else {
				h.sendBackHandler(isoSend, conn)
				return
			}
		} else {
			msgResponse, err := iso.BuildErrorResponse("", RCErrInvalidTrx, 1)
			if err != nil {
				conn.Close()
				h.Log.Errorf("client handler -> %v", err)
				return
			}
			h.sendBackHandler(msgResponse, conn)
			return
		}
	}
}

// balikan dari fungsi ini 1. message iso, 2. type 0=diteruskan ke host, 1=dibalikan ke client, 3. error
func (h *Handler) clientPrepare(msg []byte) ([]byte, int64, int, errorMessage) {
	isoReqString := strings.ToUpper(hex.EncodeToString(msg))
	isoSend, err := iso.IsoConvertToAscii([]byte(isoReqString))
	if err != nil {
		return nil, 0, 1, errorMessage{Err: fmt.Errorf("client prepare -> %s", err), RC: RCErrGeneral}
	}

	isomessage := iso8583.NewMessage(iso.Spec87)

	if err := isomessage.Unpack(isoSend); err != nil {
		return nil, 0, 1, errorMessage{Err: fmt.Errorf("client prepare -> unpack iso: %s", err), RC: RCErrGeneral}
	}
	// iso8583.Describe(isomessage, os.Stdout)

	mti, err := isomessage.GetMTI()
	if err != nil {
		return nil, 0, 1, errorMessage{Err: fmt.Errorf("client prepare -> get mti: %s", err), RC: RCErrGeneral}
	}

	stan, err := isomessage.GetString(11)
	if err != nil {
		return nil, 0, 1, errorMessage{Err: fmt.Errorf("client prepare -> unpack stan: %s", err), RC: RCErrGeneral}
	}

	// var isoSend []byte
	var stanHost string
	if stan != "" {
		isoSend, stanHost, err = h.changeStanFromClient(isoSend)
		if err != nil {
			return nil, 0, 1, errorMessage{Err: fmt.Errorf("client prepare -> change stan: %s", err), RC: RCErrGeneral}
		}
	} else {
		return nil, 0, 1, errorMessage{Err: fmt.Errorf("client prepare -> stan empty %s", ""), RC: RCErrFormatError}
	}

	var idTrx int64
	if mti == "0800" {
		bit70, err := isomessage.GetString(70)
		if err != nil {
			return nil, 0, 1, errorMessage{Err: fmt.Errorf("client prepare -> unpack bit 70: %s", err), RC: RCErrGeneral}
		}

		isoSend, err = h.networkManagementCore(isomessage, isoSend, stanHost)
		if err != nil {
			return nil, 0, 1, errorMessage{Err: fmt.Errorf("client prepare -> %s", err), RC: RCErrGeneral}
		}

		if bit70 == NetMgmtTypeLogon {
			return isoSend, 0, 1, errorMessage{}
		}
	} else if mti == "0200" {
		t := time.Now().UTC()
		jdn := f.JulianDayNumber(t)
		stanHostInt, err := strconv.Atoi(stanHost)
		if err != nil {
			return nil, 0, 1, errorMessage{Err: fmt.Errorf("client prepare -> %s", err), RC: RCErrGeneral}
		}

		err = isomessage.Field(37, fmt.Sprintf("%06d%06d", jdn%1000000, stanHostInt))
		if err != nil {
			return nil, 0, 1, errorMessage{Err: fmt.Errorf("client prepare -> %s", err), RC: RCErrGeneral}
		}

		idTrx, err = h.transactionCore(isomessage, isoReqString, stanHost)
		if err != nil {
			return nil, 0, 1, errorMessage{Err: fmt.Errorf("client prepare -> %s", err), RC: RCErrGeneral}
		}

		pinBlock, err := isomessage.GetString(52)
		if err != nil {
			return nil, 0, 1, errorMessage{Err: fmt.Errorf("client prepare -> unpack pinblock: %s", err), RC: RCErrGeneral}
		}
		if pinBlock != "" {
			pan, err := isomessage.GetString(2)
			if err != nil {
				return nil, 0, 1, errorMessage{Err: fmt.Errorf("client prepare -> unpack pan: %s", err), RC: RCErrGeneral}
			}
			tid, err := isomessage.GetString(41)
			if err != nil {
				return nil, 0, 1, errorMessage{Err: fmt.Errorf("client prepare -> unpack tid: %s", err), RC: RCErrGeneral}
			}

			zpk, err := repo.KeyGetZPK(context.Background(), h.db)
			if err != nil {
				return nil, 0, 1, errorMessage{Err: fmt.Errorf("client prepare -> get zpk from db: %s", err), RC: RCErrGeneral}
			}

			tpk, err := repo.TerminalKeyGetTPK(context.Background(), h.db, &repo.TerminalKey{Tid: tid})
			if err != nil {
				return nil, 0, 1, errorMessage{Err: fmt.Errorf("client prepare -> get tpk from db: %s", err), RC: RCErrGeneral}
			}

			panStart := len(pan) - 13
			panEnd := len(pan) - 1
			panParsed := pan[panStart:panEnd]

			newPinBlock, err := f.HSMTranslatePin(h.Config.HsmAddress, tpk, zpk, pinBlock, panParsed)
			if err != nil {
				return nil, 0, 1, errorMessage{Err: fmt.Errorf("client prepare -> hsm translate pin: %s", err), RC: "55"}
			}

			err = isomessage.Field(11, stanHost)
			if err != nil {
				return nil, 0, 1, errorMessage{Err: fmt.Errorf("client prepare -> add stan to iso: %s", err), RC: RCErrGeneral}
			}

			err = isomessage.Field(52, newPinBlock)
			if err != nil {
				return nil, 0, 1, errorMessage{Err: fmt.Errorf("client prepare -> add pinblock to iso: %s", err), RC: RCErrGeneral}
			}

			isoSend, err = isomessage.Pack()
			if err != nil {
				return nil, 0, 1, errorMessage{Err: fmt.Errorf("client prepare -> pack iso trx : %s", err), RC: RCErrGeneral}
			}
		}
	} else if mti == "0400" {
		procode, err := isomessage.GetString(3)
		if err != nil {
			return nil, 0, 1, errorMessage{Err: fmt.Errorf("client prepare -> unpack procode: %s", err), RC: RCErrGeneral}
		}

		amountStr, err := isomessage.GetString(4)
		if err != nil {
			return nil, 0, 1, errorMessage{Err: fmt.Errorf("client prepare -> unpack amount: %s", err), RC: RCErrGeneral}
		}
		var amount int64
		if amountStr != "" {
			amount, err = strconv.ParseInt(amountStr, 10, 64)
			if err != nil {
				return nil, 0, 1, errorMessage{Err: fmt.Errorf("client prepare -> convert amount: %s", err), RC: RCErrGeneral}
			}
		}

		bit12, err := isomessage.GetString(12)
		if err != nil {
			return nil, 0, 1, errorMessage{Err: fmt.Errorf("client prepare -> unpack bit 12: %s", err), RC: RCErrGeneral}
		}
		bit13, err := isomessage.GetString(13)
		if err != nil {
			return nil, 0, 1, errorMessage{Err: fmt.Errorf("client prepare -> unpack bit 13: %s", err), RC: RCErrGeneral}
		}
		loc, _ := time.LoadLocation("Asia/Jakarta")
		var trxDate *time.Time
		if bit12 != "" && bit13 != "" {
			trxDateStr := strconv.Itoa(time.Now().Year()) + "-" + bit13[:2] + "-" + bit13[2:4] + " " + bit12[:2] + ":" + bit12[2:4] + ":" + bit12[4:6]
			trxDateOri, err := time.ParseInLocation("2006-01-02 15:04:05", trxDateStr, loc)
			if err != nil {
				return nil, 0, 1, errorMessage{Err: fmt.Errorf("client prepare -> conver datetime: %s", err), RC: RCErrGeneral}
			}
			trxDate = &trxDateOri
		}

		tid, err := isomessage.GetString(41)
		if err != nil {
			return nil, 0, 1, errorMessage{Err: fmt.Errorf("client prepare -> unpack tid: %s", err), RC: RCErrGeneral}
		}
		mid, err := isomessage.GetString(42)
		if err != nil {
			return nil, 0, 1, errorMessage{Err: fmt.Errorf("client prepare -> unpack mid: %s", err), RC: RCErrGeneral}
		}

		stanClient := fmt.Sprintf("%06s", stan)
		stanHostDB, err := repo.TransactionGetStanHost(context.Background(), h.db, &repo.TransactionHistory{
			Mti:     "0200",
			Procode: procode,
			Amount:  amount,
			Stan:    stanClient,
			Tid:     tid,
			Mid:     mid,
			TrxDate: trxDate,
		})
		if err != nil {
			return nil, 0, 1, errorMessage{Err: fmt.Errorf("client prepare -> get stan host from db: %s", err), RC: RCErrGeneral}
		}

		if stanHostDB == "" {
			delete(h.stanManage, stanHost)

			isoSend, err := iso.CreateIsoResReversal(msg)
			if err != nil {
				return nil, 0, 1, errorMessage{Err: fmt.Errorf("client prepare -> %s", err), RC: RCErrGeneral}
			}

			return isoSend, 0, 1, errorMessage{}
		}

		stanManage := StanManage{StanClient: stan, Duration: time.Now()}
		h.stanManage[stanHostDB] = stanManage

		idTrx, err = h.transactionCore(isomessage, isoReqString, stanHostDB)
		if err != nil {
			return nil, 0, 1, errorMessage{Err: fmt.Errorf("client prepare -> %s", err), RC: RCErrGeneral}
		}

		err = isomessage.Field(11, stanHostDB)
		if err != nil {
			return nil, 0, 1, errorMessage{Err: fmt.Errorf("client prepare -> set stan to iso: %s", err), RC: RCErrGeneral}
		}

		_, ok := h.reversalAdvice[tid+stanHostDB]
		if !ok {
			isomessage.MTI("0420")
		} else {
			isomessage.MTI("0421")
		}

		isoSend, err = isomessage.Pack()
		if err != nil {
			return nil, 0, 1, errorMessage{Err: fmt.Errorf("client prepare -> pack iso reversal : %s", err), RC: RCErrGeneral}
		}
	} else {
		return nil, 0, 1, errorMessage{Err: fmt.Errorf("client prepare -> invalid mti: %s", err), RC: RCErrInvalidTrx}
	}

	return isoSend, idTrx, 0, errorMessage{}
}

func (h *Handler) sendBackHandler(msg []byte, conn net.Conn) {
	if conn != nil {
		TPDU := "6000000000"
		clientMsg := strings.ToUpper(hex.EncodeToString(msg))
		i := len(clientMsg)
		value, ok := h.tpduConn.Load(conn)
		if ok {
			TPDU = value.(string)
		}
		h.tpduConn.Delete(conn)

		TPDU = TPDU[:2] + TPDU[6:10] + TPDU[2:6]
		i = (i + 10) / 2
		msgWlen := fmt.Sprintf("%04X%s%s", i, TPDU, clientMsg)
		msgSend, err := hex.DecodeString(msgWlen)
		if err != nil {
			h.handleErrorAndRespond(conn, "", RCErrGeneral, "send back handler - decode iso:", err)
			return
		}
		isoString := strings.ToUpper(hex.EncodeToString(msgSend))

		_, err = conn.Write(msgSend)
		if err != nil {
			if opErr, ok := err.(*net.OpError); ok && opErr.Timeout() {
				h.Log.Errorf("send back handler -> write data client timeout: %v", err)
				return
			}
			h.Log.Errorf("send back handler -> write data client: %v", err)
			return
		}

		h.Log.WithField("debug_tag", "dl_out").Debugf("message response [%s]: %s", conn.RemoteAddr().String(), isoString)
	}
}

func (h *Handler) changeStanFromClient(msg []byte) ([]byte, string, error) {
	h.mu.Lock()
	defer h.mu.Unlock()

	isoStr := string(msg)
	isomessage := iso8583.NewMessage(iso.Spec87)
	err := isomessage.Unpack([]byte(isoStr))
	if err != nil {
		return nil, "", err
	}

	stanClient, err := isomessage.GetString(11)
	if err != nil {
		return nil, "", err
	}

	stanClient = fmt.Sprintf("%06s", stanClient)
	stanHostInt := h.stan
	stanHost := fmt.Sprintf("%012d", stanHostInt)

	err = isomessage.Field(11, stanHost)
	if err != nil {
		return nil, "", err
	}

	err = h.editConfig(stanHostInt)
	if err != nil {
		return nil, "", err
	}

	err = h.loadConfig()
	if err != nil {
		return nil, "", err
	}

	newMsg, err := isomessage.Pack()
	if err != nil {
		return nil, "", err
	}

	stanManage := StanManage{StanClient: stanClient, Duration: time.Now()}
	h.stanManage[stanHost] = stanManage

	return newMsg, stanHost, nil
}

func (h *Handler) transactionCore(isomessage *iso8583.Message, msg, stanHost string) (int64, error) {
	mti, err := isomessage.GetMTI()
	if err != nil {
		return 0, fmt.Errorf("transaction core -> get mti: %w", err)
	}

	procode, err := isomessage.GetString(3)
	if err != nil {
		return 0, fmt.Errorf("transaction core -> unpack procode: %w", err)
	}
	tid, err := isomessage.GetString(41)
	if err != nil {
		return 0, fmt.Errorf("transaction core -> unpack tid: %w", err)
	}
	mid, err := isomessage.GetString(42)
	if err != nil {
		return 0, fmt.Errorf("transaction core -> unpack mid: %w", err)
	}
	pan, err := isomessage.GetString(2)
	if err != nil {
		return 0, fmt.Errorf("transaction core -> unpack pan: %w", err)
	}
	if pan != "" {
		pan = f.MaskPan(pan)
	}
	amountStr, err := isomessage.GetString(4)
	if err != nil {
		return 0, fmt.Errorf("transaction core -> unpack amount: %w", err)
	}
	var amount int64
	if amountStr != "" {
		amount, err = strconv.ParseInt(amountStr, 10, 64)
		if err != nil {
			return 0, fmt.Errorf("transaction core -> convert amount: %w", err)
		}
	}
	bit12, err := isomessage.GetString(12)
	if err != nil {
		return 0, fmt.Errorf("transaction core -> unpack bit 12: %w", err)
	}
	bit13, err := isomessage.GetString(13)
	if err != nil {
		return 0, fmt.Errorf("transaction core -> unpack bit 13: %w", err)
	}
	loc, _ := time.LoadLocation("Asia/Jakarta")
	var trxDate *time.Time
	if bit12 != "" && bit13 != "" {
		trxDateStr := strconv.Itoa(time.Now().Year()) + "-" + bit13[:2] + "-" + bit13[2:4] + " " + bit12[:2] + ":" + bit12[2:4] + ":" + bit12[4:6]
		trxDateOri, err := time.ParseInLocation("2006-01-02 15:04:05", trxDateStr, loc)
		if err != nil {
			return 0, fmt.Errorf("transaction core -> convert datetime: %w", err)
		}
		trxDate = &trxDateOri
	}
	stan, err := isomessage.GetString(11)
	if err != nil {
		return 0, fmt.Errorf("transaction core -> unpack stan: %w", err)
	}
	stan = fmt.Sprintf("%06s", stan)

	merhcantName, err := isomessage.GetString(43)
	if err != nil {
		return 0, fmt.Errorf("transaction core -> unpack merchant name: %w", err)
	}
	rrn, err := isomessage.GetString(37)
	if err != nil {
		return 0, fmt.Errorf("transaction core -> unpack rrn: %w", err)
	}

	idTrx, err := repo.TransactionHistorySave(context.Background(), h.db, &repo.TransactionHistory{
		Mti:          mti,
		Procode:      procode,
		Tid:          tid,
		Mid:          mid,
		Pan:          pan,
		Amount:       amount,
		TrxDate:      trxDate,
		Stan:         stan,
		StanHost:     stanHost,
		Rrn:          rrn,
		MerchantName: merhcantName,
		IsoReq:       msg,
		CreatedAt:    time.Now(),
	})
	if err != nil {
		return 0, fmt.Errorf("transaction core -> save trx: %w", err)
	}

	return idTrx, nil
}

func (h *Handler) networkManagementCore(isomessage *iso8583.Message, msg []byte, stanHost string) ([]byte, error) {
	tid, err := isomessage.GetString(41)
	if err != nil {
		return nil, fmt.Errorf("network management -> unpack tid: %w", err)
	}

	bit32, err := isomessage.GetString(32)
	if err != nil {
		return nil, fmt.Errorf("network management -> unpack bit 23: %w", err)
	}

	bit70, err := isomessage.GetString(70)
	if err != nil {
		return nil, fmt.Errorf("network management -> unpack bit 70: %w", err)
	}

	var isoSend []byte

	if bit70 == NetMgmtTypeLogon {
		tmk, err := repo.KeyGetTMK(context.Background(), h.db)
		if err != nil {
			return nil, fmt.Errorf("network management -> get tmk: %w", err)
		}

		twk, tpk, err := f.HSMGenerateKey(h.Config.HsmAddress, tmk)
		if err != nil {
			return nil, fmt.Errorf("generate key hsm: %w", err)
		}
		// twk := "60C49773967F03579F9E28CA7AA30DDD"
		// tpk := tmk

		err = repo.TerminalKeySave(context.Background(), h.db, &repo.TerminalKey{
			Tid:       tid,
			Tpk:       tpk,
			CreatedAt: time.Now(),
		})
		if err != nil {
			return nil, fmt.Errorf("network management -> save tpk: %w", err)
		}

		stan, err := isomessage.GetString(11)
		if err != nil {
			return nil, fmt.Errorf("network management -> unpack stan: %w", err)
		}
		delete(h.stanManage, stanHost)

		isoSend, err = iso.CreateIsoResLogon(msg, twk, stan)
		if err != nil {
			return nil, fmt.Errorf("network management -> create res iso logon: %w", err)
		}

		isoSend, err = iso.IsoConvertToHex([]byte(isoSend))
		if err != nil {
			return nil, fmt.Errorf("network management -> convert res iso logon to hex: %w", err)
		}
	} else if bit70 == NetMgmtTypeSignOn {
		isoSend, err = iso.CreateIsoSignOn(stanHost, bit32)
		if err != nil {
			return nil, fmt.Errorf("network management -> create iso sign on: %w", err)
		}
	} else if bit70 == NetMgmtTypeSignOff {
		isoSend, err = iso.CreateIsoSignOff(stanHost, bit32)
		if err != nil {
			return nil, fmt.Errorf("network management -> create iso sign off: %w", err)
		}
	} else if bit70 == NetMgmtTypeNewKey {
		isoSend, err = iso.CreateIsoNewKey(stanHost, bit32)
		if err != nil {
			return nil, fmt.Errorf("network management -> create iso new key: %w", err)
		}
	} else if bit70 == NetMgmtTypeEcho {
		isoSend, err = iso.CreateIsoEchoTest(stanHost, bit32)
		if err != nil {
			return nil, fmt.Errorf("network management -> create iso echo test: %w", err)
		}
	} else {
		return nil, fmt.Errorf("network management -> invalid %s", "bit70")
	}

	return isoSend, nil
}
