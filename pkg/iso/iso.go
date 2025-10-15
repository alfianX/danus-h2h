package iso

import (
	"encoding/hex"
	"fmt"
	"os"
	"time"

	"github.com/moov-io/iso8583"
)

func IsoConvertToAscii(msg []byte) ([]byte, error) {
	isoStr := string(msg)
	isomessageHex := iso8583.NewMessage(Spec87Hex)
	err := isomessageHex.Unpack([]byte(isoStr))
	if err != nil {
		return nil, fmt.Errorf("ISO convert to ASCII -> fail unpack ISO!, err: %v", err)
	}

	mti, err := isomessageHex.GetMTI()
	if err != nil {
		return nil, fmt.Errorf("ISO convert to ASCII -> fail parsing MTI!, err: %v", err)
	}

	isomessage := iso8583.NewMessage(Spec87)
	isomessage.MTI(mti)

	fields := isomessageHex.GetFields()
	for r := range fields {
		bit, err := isomessageHex.GetString(r)
		if err != nil {
			return nil, fmt.Errorf("ISO convert to ASCII -> fail parsing bit %d!, err: %v", r, err)
		}

		err = isomessage.Field(r, bit)
		if err != nil {
			return nil, fmt.Errorf("ISO convert to ASCII -> fail set bit %d!, err: %v", r, err)
		}
	}

	rawMessage, err := isomessage.Pack()
	if err != nil {
		return nil, fmt.Errorf("ISO convert to ASCII -> ISO pack fail, err: %v", err)
	}

	return rawMessage, nil
}

func IsoConvertToHex(msg []byte) ([]byte, error) {
	isoStr := string(msg)
	isomessage := iso8583.NewMessage(Spec87)
	err := isomessage.Unpack([]byte(isoStr))
	if err != nil {
		return nil, fmt.Errorf("ISO convert to HEX -> fail unpack ISO!, err: %v", err)
	}

	mti, err := isomessage.GetMTI()
	if err != nil {
		return nil, fmt.Errorf("ISO convert to HEX -> fail parsing MTI!, err: %v", err)
	}
	if mti == "0430" {
		mti = "0410"
	}

	isomessageHex := iso8583.NewMessage(Spec87Hex)
	isomessage.MTI(mti)

	fields := isomessage.GetFields()
	for r := range fields {
		bit, err := isomessage.GetString(r)
		if err != nil {
			return nil, fmt.Errorf("ISO convert to HEX -> fail parsing bit %d!, err: %v", r, err)
		}

		err = isomessageHex.Field(r, bit)
		if err != nil {
			return nil, fmt.Errorf("ISO convert to HEX -> fail set bit %d!, err: %v", r, err)
		}
	}

	rawMessage, err := isomessageHex.Pack()
	if err != nil {
		return nil, fmt.Errorf("ISO convert to HEX -> ISO pack fail, err: %v", err)
	}

	msgSend, err := hex.DecodeString(string(rawMessage))
	if err != nil {
		return nil, err
	}

	return msgSend, nil
}

func CreateIsoEchoTest(stan, bit32 string) ([]byte, error) {
	isomessage := iso8583.NewMessage(Spec87Hex)

	now := time.Now().UTC()
	de7 := now.Format("0102150405")
	de11 := fmt.Sprintf("%06s", stan)

	err := isomessage.Field(7, de7)
	if err != nil {
		return nil, err
	}
	err = isomessage.Field(11, de11)
	if err != nil {
		return nil, err
	}
	err = isomessage.Field(32, bit32)
	if err != nil {
		return nil, err
	}
	err = isomessage.Field(70, "301")
	if err != nil {
		return nil, err
	}
	isomessage.MTI("0800")

	iso, err := isomessage.Pack()
	if err != nil {
		return nil, err
	}

	return iso, nil

	// sign on 001
	// sign off 002
	// logon edc 101
	// req 102
	// cut off 201
	// echo test 301

}

func CreateIsoSignOn(stan, bit32 string) ([]byte, error) {
	isomessage := iso8583.NewMessage(Spec87Hex)

	now := time.Now().UTC()
	de7 := now.Format("0102150405")
	de11 := fmt.Sprintf("%06s", stan)

	err := isomessage.Field(7, de7)
	if err != nil {
		return nil, err
	}
	err = isomessage.Field(11, de11)
	if err != nil {
		return nil, err
	}
	err = isomessage.Field(32, bit32)
	if err != nil {
		return nil, err
	}
	err = isomessage.Field(70, "001")
	if err != nil {
		return nil, err
	}
	isomessage.MTI("0800")

	iso, err := isomessage.Pack()
	if err != nil {
		return nil, err
	}

	return iso, nil

	// sign on 001
	// sign off 002
	// req 102
	// cut off 201
}

func CreateIsoSignOff(stan, bit32 string) ([]byte, error) {
	isomessage := iso8583.NewMessage(Spec87Hex)

	now := time.Now().UTC()
	de7 := now.Format("0102150405")
	de11 := fmt.Sprintf("%06s", stan)

	err := isomessage.Field(7, de7)
	if err != nil {
		return nil, err
	}
	err = isomessage.Field(11, de11)
	if err != nil {
		return nil, err
	}
	err = isomessage.Field(32, bit32)
	if err != nil {
		return nil, err
	}
	err = isomessage.Field(70, "002")
	if err != nil {
		return nil, err
	}
	isomessage.MTI("0800")

	iso, err := isomessage.Pack()
	if err != nil {
		return nil, err
	}

	return iso, nil

	// sign on 001
	// sign off 002
	// req 102
	// cut off 201
}

func CreateIsoNewKey(stan, bit32 string) ([]byte, error) {
	isomessage := iso8583.NewMessage(Spec87Hex)

	now := time.Now().UTC()
	de7 := now.Format("0102150405")
	de11 := fmt.Sprintf("%06s", stan)

	err := isomessage.Field(7, de7)
	if err != nil {
		return nil, err
	}
	err = isomessage.Field(11, de11)
	if err != nil {
		return nil, err
	}
	err = isomessage.Field(32, bit32)
	if err != nil {
		return nil, err
	}
	err = isomessage.Field(70, "102")
	if err != nil {
		return nil, err
	}
	isomessage.MTI("0800")

	iso, err := isomessage.Pack()
	if err != nil {
		return nil, err
	}

	return iso, nil

	// sign on 001
	// sign off 002
	// req 102
	// cut off 201
}

func CreateIsoResNman(msg []byte) ([]byte, error) {
	isoStr := string(msg)
	isomessage := iso8583.NewMessage(Spec87)
	err := isomessage.Unpack([]byte(isoStr))
	if err != nil {
		return nil, err
	}

	err = isomessage.Field(39, "00")
	if err != nil {
		return nil, err
	}

	isomessage.MTI("0810")

	rawMessage, err := isomessage.Pack()
	if err != nil {
		return nil, err
	}

	return rawMessage, nil
}

func CreateIsoResKeyChange(msg []byte) ([]byte, error) {
	isoStr := hex.EncodeToString(msg)
	isomessage := iso8583.NewMessage(Spec87)
	err := isomessage.Unpack([]byte(isoStr))
	if err != nil {
		return nil, err
	}

	isomessage.UnsetField(48)
	err = isomessage.Field(39, "00")
	if err != nil {
		return nil, err
	}
	isomessage.MTI("0810")

	rawMessage, err := isomessage.Pack()
	if err != nil {
		return nil, err
	}

	msgSend, err := hex.DecodeString(string(rawMessage))
	if err != nil {
		return nil, err
	}
	return msgSend, nil
}

func CreateIsoResLogon(msg []byte, bit48 string) ([]byte, error) {
	isoStr := hex.EncodeToString(msg)
	isomessage := iso8583.NewMessage(Spec87Hex)
	err := isomessage.Unpack([]byte(isoStr))
	if err != nil {
		return nil, err
	}

	err = isomessage.Field(39, "00")
	if err != nil {
		return nil, err
	}
	err = isomessage.Field(48, bit48)
	if err != nil {
		return nil, err
	}

	isomessage.MTI("0810")

	rawMessage, err := isomessage.Pack()
	if err != nil {
		return nil, err
	}

	iso8583.Describe(isomessage, os.Stdout)

	msgSend, err := hex.DecodeString(string(rawMessage))
	if err != nil {
		return nil, err
	}
	return msgSend, nil
}

func CreateIsoResReversal(msg []byte) ([]byte, error) {
	isoStr := hex.EncodeToString(msg)
	isomessage := iso8583.NewMessage(Spec87Hex)
	err := isomessage.Unpack([]byte(isoStr))
	if err != nil {
		return nil, err
	}

	err = isomessage.Field(39, "00")
	if err != nil {
		return nil, err
	}

	isomessage.MTI("0410")

	rawMessage, err := isomessage.Pack()
	if err != nil {
		return nil, err
	}

	iso8583.Describe(isomessage, os.Stdout)

	msgSend, err := hex.DecodeString(string(rawMessage))
	if err != nil {
		return nil, err
	}
	return msgSend, nil
}

// buildErrorResponse adalah fungsi helper untuk membuat respons error ISO 8583
func BuildErrorResponse(msg, responseCode string, isoType int) ([]byte, error) {
	var specIso *iso8583.MessageSpec
	if isoType == 1 {
		specIso = Spec87Hex
	} else if isoType == 2 {
		specIso = Spec87
	}

	var originalMTI string
	var originalSTAN string
	isomessageRes := iso8583.NewMessage(specIso)
	now := time.Now().UTC()
	bit7 := now.Format("0102150405")

	isomessage := iso8583.NewMessage(specIso)
	if msg != "" {
		err := isomessage.Unpack([]byte(msg))
		if err != nil {
			return nil, fmt.Errorf("build error response -> failed to unpack iso: %w", err)
		}

		originalMTI, err = isomessage.GetMTI()
		if err != nil {
			return nil, fmt.Errorf("build error response -> failed to unpack mti: %w", err)
		}

		originalSTAN, err = isomessage.GetString(11)
		if err != nil {
			return nil, fmt.Errorf("build error response -> failed to unpack stan: %w", err)
		}
	}

	responseMTI := ""
	if len(originalMTI) == 4 {
		mtiType := string(originalMTI[0])
		mtiVersion := string(originalMTI[1])
		mtiOriginator := string(originalMTI[2])
		if mtiOriginator == "0" {
			responseMTI = fmt.Sprintf("%s%s1%s", mtiType, mtiVersion, string(originalMTI[3]))
		} else {
			responseMTI = originalMTI
		}
	} else {
		responseMTI = "0000" // Default MTI untuk pesan yang rusak total
	}
	isomessageRes.MTI(responseMTI)

	if err := isomessageRes.Field(7, bit7); err != nil {
		return nil, fmt.Errorf("build error response -> failed to set datetime (7): %w", err)
	}

	if originalSTAN != "" {
		if err := isomessageRes.Field(11, originalSTAN); err != nil {
			return nil, fmt.Errorf("build error response -> failed to set stan (11): %w", err)
		}
	}

	if err := isomessageRes.Field(39, responseCode); err != nil {
		return nil, fmt.Errorf("build error response -> failed to set response code (39): %w", err)
	}

	packedResponse, err := isomessageRes.Pack()
	if err != nil {
		return nil, fmt.Errorf("build error response -> failed to pack error response: %w", err)
	}

	var msgResponse []byte
	if isoType == 1 {
		msgResponse, err = hex.DecodeString(string(packedResponse))
		if err != nil {
			return nil, err
		}
	} else if isoType == 2 {
		msgResponse = packedResponse
	}

	return msgResponse, nil
}

func TestIso() {
	isomessage := iso8583.NewMessage(Spec87Hex)

	// isomessage.MTI("0200")
	// isomessage.Field(11, "000002")
	// isomessage.Field(35, "5894505000165999D25122260000067300000")
	// isomessage.Field(41, "99980001")

	// iso, err := isomessage.Pack()
	// if err != nil {
	// 	log.Fatalln(err)
	// }

	// strIso := string(iso)
	// i := len(strIso)
	// TPDU := "6000000001"
	// i = (i + 10) / 2
	// msgWlen := fmt.Sprintf("%04X%s%s", i, TPDU, strIso)

	// fmt.Println(msgWlen)
	isomessage.Unpack([]byte("02007238C40128B1920016589450500016599931000001065560000008081531150000341531150808080860120021042000365894505000165999D251222600000673000030323630373131313239333039393938303030314D414C414E47202020202020202020202020202020202020202020202020202020202020202020200389878300163030303132303135303036343632393833303685D5DF61BE6500A201669B02E8005A0852641401117319685F24034410319F0702AB809F3303E0F8C89F34030200009F3501229F41040000006782025400950542800480005F2A0203605F3401009A032412029C01009F02060000050000009F03060000000000009F101C0101A0001100003F5E55A800000000000000000000000000000000009F1A0203609F260805850FC91B2C285B9F360203AB9F2701809F37047842C68A8407A0000006021010"))

	iso8583.Describe(isomessage, os.Stdout)

	fmt.Println("OK")
}
