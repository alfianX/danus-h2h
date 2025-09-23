package handler

import (
	"context"
	"encoding/hex"
	"fmt"
	"io"
	"net"
	"strconv"
	"time"

	"github.com/alfianX/danus-h2h/internal/repo"
	f "github.com/alfianX/danus-h2h/pkg/function"
	"github.com/alfianX/danus-h2h/pkg/iso"
	"github.com/moov-io/iso8583"
)

func (h *Handler) ConnectToHost() {
	h.Log.Infof("Try to connect to host %s...", h.Config.HostAddress)

	// Jeda awal sebelum retry
	time.Sleep(5 * time.Second)

	// Logika retry dengan backoff eksponensial
	for backoff := 1 * time.Second; backoff <= 3600*time.Second; backoff *= 2 {
		hostConn, err := net.Dial("tcp", h.Config.HostAddress)
		if err == nil {
			h.Log.Infof("Successfully connected to host %s", h.Config.HostAddress)
			// h.lastPingSent = sync.Map{}
			// h.lastPongReceived = sync.Map{}
			h.hostConnLock.Lock()
			h.hostConn = hostConn
			h.hostConnLock.Unlock()
			// go h.hostHealthCheck()
			go h.hostHandler()
			return
		}

		h.Log.Errorf("Failed to connect to host %s: %v. Retrying in %v...", h.Config.HostAddress, err, backoff)
		time.Sleep(backoff)
	}

	// Jika sampai di sini, artinya semua upaya reconnect gagal
	h.Log.Fatalf("Fatal: all reconnect attempts to host %s failed. Exiting.", h.Config.HostAddress)
}

func (h *Handler) hostHandler() {
	// Defer ini akan membersihkan koneksi dan memicu reconnect
	// hanya saat hostHandler berhenti karena error.
	defer func() {
		h.Log.Warnf("Host handler for %s is stopping. Initiating reconnect...", h.Config.HostAddress)
		h.hostConnLock.Lock()
		if h.hostConn != nil {
			h.hostConn.Close()
			h.hostConn = nil // Set nil agar goroutine lain tahu koneksi sudah putus
		}
		h.hostConnLock.Unlock()
		go h.ConnectToHost()
		// go h.checkConnectionStatus()
	}()

	for {
		header := make([]byte, HeaderLen)
		_, err := io.ReadFull(h.hostConn, header)
		if err != nil {
			if err == io.EOF {
				h.Log.Warnf("host handler -> Host connection closed gracefully.")
			} else {
				h.Log.Errorf("host handler -> Error reading from host: %v", err)
			}
			return
		}

		headerStr := hex.EncodeToString(header)
		msgLength, err := strconv.ParseInt(headerStr, 16, 64)
		if err != nil {
			h.Log.Errorf("host handler -> failed to parse message length from host : %v", err)
			continue
		}

		messageBytes := make([]byte, msgLength)
		_, err = io.ReadFull(h.hostConn, messageBytes)
		if err != nil {
			if err == io.EOF {
				h.Log.Warnf("host handler -> Host connection closed gracefully during message body read.")
			} else {
				h.Log.Errorf("host handler -> Error reading message body from host: %v", err)
			}
			return
		}

		fullMessage := append(header, messageBytes...)
		isoStr := string(fullMessage[2:])

		isomessage := iso8583.NewMessage(iso.Spec87)
		err = isomessage.Unpack([]byte(isoStr))
		if err != nil {
			h.Log.Errorf("host handler -> failed to unpack ISO message: %v", err)
			continue
		}

		mti, err := isomessage.GetMTI()
		if err != nil {
			h.Log.Errorf("host handler -> failed to get MTI: %v", err)
			continue // Kembali
		}

		if mti == "0800" {
			go h.networkManagementHandler(fullMessage[2:])
		} else {
			// if mti == "0810" {
			// 	bit70, err := isomessage.GetString(70)
			// 	if err != nil {
			// 		h.Log.Errorf("host handler -> failed to get bit70: %v", err)
			// 		continue
			// 	}

			// 	if bit70 == "301" {
			// 		// h.Log.Info("Received echo test response (0810) from host.")
			// 		// Perbarui timestamp respons terakhir yang diterima
			// 		h.lastPongReceived.Store(true, time.Now())
			// 		continue
			// 	}
			// }
			stan, err := isomessage.GetString(11)
			if err != nil {
				h.Log.Errorf("host handler -> failed to get STAN: %v", err)
				continue
			}
			stan = fmt.Sprintf("%06s", stan)
			value, ok := h.responseMap.Load(stan)
			if !ok {
				h.Log.Warnf("host handler -> Received unexpected host response for stan: %s. Ignoring.", stan)
				continue
			}
			responseChan, ok := value.(chan HostResponse)
			if !ok {
				h.Log.Errorf("host handler -> Invalid channel type for stan: %s.", stan)
				continue
			}

			responseChan <- HostResponse{Data: fullMessage[2:], Err: nil}
		}

	}
}

func (h *Handler) sendSingleHostHandler(conn net.Conn, msg []byte, idTrx int64) {
	h.hostConnLock.Lock()
	hostConn := h.hostConn
	h.hostConnLock.Unlock()

	if hostConn == nil {
		h.handleErrorAndRespond(conn, "", "96", "send single host handler -> host is not connected", fmt.Errorf("host not connected"))
		return
	}

	isomessage := iso8583.NewMessage(iso.Spec87)
	if err := isomessage.Unpack(msg); err != nil {
		h.handleErrorAndRespond(conn, "", "96", "send single host handler - unpack iso:", err)
		return
	}
	mti, err := isomessage.GetMTI()
	if err != nil {
		h.handleErrorAndRespond(conn, "", "96", "send single host handler - unpack mti:", err)
		return
	}
	stan, err := isomessage.GetString(11)
	if err != nil {
		h.handleErrorAndRespond(conn, "", "96", "send single host handler - unpack bit 11:", err)
		return
	}
	stan = fmt.Sprintf("%06s", stan)

	// 1. Buat channel respons unik untuk transaksi ini
	responseChan := make(chan HostResponse, 1)

	// 2. Simpan channel ini ke map
	h.responseMap.Store(stan, responseChan)
	defer h.responseMap.Delete(stan) // Penting: Pastikan channel dihapus dari map

	isoString := string(msg)
	hostMsg := hex.EncodeToString(msg)
	i := len(msg)
	msgWlen := fmt.Sprintf("%04X%s", i, hostMsg)
	msgSend, err := hex.DecodeString(msgWlen)
	if err != nil {
		h.handleErrorAndRespond(conn, "", "96", "send single host handler - fail decode iso:", err)
		return
	}

	h.Log.WithField("debug_tag", "ul_out").Debugf("message to host : %s", isoString)

	_, err = hostConn.Write(msgSend)
	if err != nil {
		h.handleErrorAndRespond(conn, "", "96", "send single host handler - fail write to host:", err)
		return
	}

	select {
	case response := <-responseChan:
		if response.Err != nil {
			h.handleErrorAndRespond(conn, "", "96", "response from host had an error", response.Err)
			return
		} else {
			isoResponse := response.Data
			isoResponseString := string(isoResponse)

			h.Log.WithField("debug_tag", "ul_in").Debugf("message from host : %s", isoResponseString)

			isomessageRes := iso8583.NewMessage(iso.Spec87)
			err = isomessageRes.Unpack([]byte(isoResponseString))
			if err != nil {
				h.handleErrorAndRespond(conn, "", "96", "send single host handler - unpack iso:", err)
				return
			}

			tx := h.db.Begin()
			defer tx.Commit()

			bit39, err := isomessageRes.GetString(39)
			if err != nil {
				h.handleErrorAndRespond(conn, "", "96", "send single host handler - unpack bit 39:", err)
				return
			}

			isoResponse, err = iso.IsoConvertToHex([]byte(isoResponseString))
			if err != nil {
				h.handleErrorAndRespond(conn, "", "96", "send single host handler - ", err)
				return
			}

			isoResponse, err = h.changeStanFromHost(isoResponse)
			if err != nil {
				h.handleErrorAndRespond(conn, "", "96", "send single host handler - change stan:", err)
				return
			}
			isoResponseString = hex.EncodeToString(isoResponse)

			if idTrx != 0 {
				err = repo.TransactionHistoryUpdateResponse(context.Background(), tx, &repo.TransactionHistory{
					ID:           idTrx,
					ResponseCode: bit39,
					IsoRes:       isoResponseString,
					UpdatedAt:    time.Now(),
				})
				if err != nil {
					h.handleErrorAndRespond(conn, "", "96", "send single host handler - update response trx:", err)
					return
				}
			}

			if mti == "0421" && bit39 == "00" {
				isomessageRes := iso8583.NewMessage(iso.Spec87Hex)
				err = isomessageRes.Unpack([]byte(isoResponseString))
				if err != nil {
					tx.Rollback()
					h.handleErrorAndRespond(conn, "", "96", "send single host handler - unpack iso:", err)
					return
				}
				tid, err := isomessageRes.GetString(41)
				if err != nil {
					tx.Rollback()
					h.handleErrorAndRespond(conn, "", "96", "send single host handler - unpack bit 39:", err)
					return
				}
				stanClient, err := isomessageRes.GetString(11)
				if err != nil {
					tx.Rollback()
					h.handleErrorAndRespond(conn, "", "96", "send single host handler - unpack bit 39:", err)
					return
				}
				stanClient = fmt.Sprintf("%06s", stanClient)

				delete(h.reversalAdvice, tid+stanClient)
			}

			tx.Commit()

			go h.sendBackHandler(isoResponse, conn)
		}
	case <-time.After(60 * time.Second):
		switch mti {
		case "0420":
			stanData, ok := h.stanManage[stan]
			if !ok {
				h.handleErrorAndRespond(conn, "", "96", "send single host handler - stan "+stan+" not found in map", err)
				return
			}
			stanClient := stanData.StanClient
			tid, err := isomessage.GetString(41)
			if err != nil {
				h.handleErrorAndRespond(conn, "", "96", "send single host handler - unpack bit 41:", err)
				return
			}

			reversalAvc := ReversalAdvice{Data: msg}
			h.reversalAdvice[tid+stanClient] = reversalAvc
			return
		case "0421":
			return
		default:
			return
		}
		// h.handleErrorAndRespond(conn, "", "96", "send single host handler - timeout", fmt.Errorf("timeout waiting for host response"))
	}
}

func (h *Handler) changeStanFromHost(msg []byte) ([]byte, error) {
	h.mu.Lock()
	defer h.mu.Unlock()

	isoStr := hex.EncodeToString(msg)
	isomessage := iso8583.NewMessage(iso.Spec87Hex)
	err := isomessage.Unpack([]byte(isoStr))
	if err != nil {
		return nil, err
	}

	stanHost, err := isomessage.GetString(11)
	if err != nil {
		return nil, err
	}

	stanData, ok := h.stanManage[stanHost]
	if !ok {
		return nil, fmt.Errorf("stan %s not found in map", stanHost)
	}

	stanClient := stanData.StanClient
	delete(h.stanManage, stanHost)

	err = isomessage.Field(11, stanClient)
	if err != nil {
		return nil, err
	}

	newMsg, err := isomessage.Pack()
	if err != nil {
		return nil, err
	}

	msgSend, err := hex.DecodeString(string(newMsg))
	if err != nil {
		return nil, err
	}

	return msgSend, nil
}

func (h *Handler) networkManagementHandler(msg []byte) {
	isoStr := string(msg)
	isomessage := iso8583.NewMessage(iso.Spec87)
	err := isomessage.Unpack([]byte(isoStr))
	if err != nil {
		h.Log.Errorf("network management handler -> unpack iso: %v", err)
		return
	}

	nmiCode, err := isomessage.GetString(70)
	if err != nil {
		h.Log.Errorf("network management handler -> unpack bit 70: %v", err)
		return
	}

	var isoResponse []byte
	if nmiCode == "301" || nmiCode == "001" || nmiCode == "002" {
		isoResponse, err = iso.CreateIsoResNman(msg)
		if err != nil {
			h.Log.Errorf("network management handler -> create iso response net management: %v", err)
			return
		}
	} else if nmiCode == "102" {
		zmk, err := repo.KeyGetZMK(context.Background(), h.db)
		if err != nil {
			h.Log.Errorf("network management handler -> get zmk: %v", err)
			return
		}

		de48, err := isomessage.GetString(48)
		if err != nil {
			h.Log.Errorf("network management handler -> unpack bist 48: %v", err)
			return
		}
		zpk := de48[:32]

		zpkEnc, err := f.HSMSaveZPK(h.Config.HsmAddress, zmk, zpk)
		if err != nil {
			h.Log.Errorf("network management handler -> save zpk to hsm: %v", err)
			return
		}

		err = repo.KeyUpdateZPK(context.Background(), h.db, zpkEnc)
		if err != nil {
			h.Log.Errorf("network management handler -> update zpk to db: %v", err)
			return
		}

		isoResponse, err = iso.CreateIsoResKeyChange(msg)
		if err != nil {
			h.Log.Errorf("network management handler -> create iso res key change: %v", err)
			return
		}
	}

	hostResMsg := hex.EncodeToString(isoResponse)
	i := len(msg)
	msgWlen := fmt.Sprintf("%04X%s", i, hostResMsg)
	msgSend, err := hex.DecodeString(msgWlen)
	if err != nil {
		h.Log.Errorf("network management handler -> decode msg: %v", err)
		return
	}

	_, err = h.hostConn.Write(msgSend)
	if err != nil {
		h.Log.Errorf("network management handler -> write response nm to host: %v", err)
		return
	}
}

// func (h *Handler) hostHealthCheck() {
// 	h.Log.Info("Starting host health check goroutine.")
// 	ticker := time.NewTicker(30 * time.Second)
// 	defer ticker.Stop()

// 	for range ticker.C {
// 		h.hostConnLock.Lock()
// 		conn := h.hostConn
// 		h.hostConnLock.Unlock()

// 		// Jika koneksi sudah putus, tidak perlu melakukan health check
// 		if conn == nil {
// 			h.Log.Warn("Health check skipped: host connection is nil.")
// 			continue
// 		}

// 		// Kirim pesan Network Management Request (contoh 0800)
// 		// Pesan ini tidak membutuhkan respons dari responseMap
// 		isoString := "0800822000010000000004000000000000000729093815000041042000301"
// 		hostMsg := hex.EncodeToString([]byte(isoString))
// 		i := len(isoString)
// 		msgWlen := fmt.Sprintf("%04X%s", i, hostMsg)
// 		msgSend, err := hex.DecodeString(msgWlen)
// 		if err != nil {
// 			h.Log.Errorf("Failed to create 0800 message: %v", err)
// 			continue
// 		}

// 		// h.Log.Debug("Sending health check message to host.")

// 		// Hanya kirim, jangan baca respons di sini.
// 		conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
// 		_, err = conn.Write(msgSend)
// 		conn.SetWriteDeadline(time.Time{})

// 		if err != nil {
// 			h.Log.Errorf("Error writing health check message: %v", err)
// 			h.forceDisconnect()
// 			continue
// 		}

// 		// Perbarui timestamp untuk melacak kapan ping terakhir dikirim
// 		h.lastPingSent.Store(true, time.Now())
// 	}
// }

// // Fungsi helper untuk memutuskan koneksi secara paksa dan aman
// func (h *Handler) forceDisconnect() {
// 	h.hostConnLock.Lock()
// 	if h.hostConn != nil {
// 		h.Log.Warn("Forcing host connection to close.")
// 		h.hostConn.Close()
// 		h.hostConn = nil

// 		h.lastPingSent = sync.Map{}
// 		h.lastPongReceived = sync.Map{}
// 	}
// 	h.hostConnLock.Unlock()
// }

// func (h *Handler) checkConnectionStatus() {
// 	// h.Log.Info("Starting connection status checker goroutine.")

// 	// Timeout ping, harus lebih lama dari interval pengiriman (30 detik)
// 	pingTimeout := 45 * time.Second

// 	for {
// 		time.Sleep(10 * time.Second) // Cek setiap 10 detik

// 		// Cek kapan respons terakhir diterima
// 		lastPong, ok := h.lastPongReceived.Load(true)
// 		if !ok {
// 			// Jika belum pernah menerima pong, cek kapan ping terakhir dikirim
// 			lastPing, pingOk := h.lastPingSent.Load(true)
// 			if pingOk && time.Since(lastPing.(time.Time)) > pingTimeout {
// 				h.Log.Errorf("Connection status check failed: No pong received after initial ping.")
// 				h.forceDisconnect()
// 			}
// 			continue
// 		}

// 		// Jika respons terakhir terlalu lama, anggap koneksi mati
// 		if time.Since(lastPong.(time.Time)) > pingTimeout {
// 			h.Log.Errorf("Connection status check failed: Host is not responding.")
// 			h.forceDisconnect()
// 		}
// 	}
// }
