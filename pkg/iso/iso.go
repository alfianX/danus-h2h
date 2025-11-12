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
	isomessage := iso8583.NewMessage(Spec87)

	now := time.Now().UTC()
	de7 := now.Format("0102150405")
	de11 := fmt.Sprintf("%012s", stan)

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
	isomessage := iso8583.NewMessage(Spec87)

	now := time.Now().UTC()
	de7 := now.Format("0102150405")
	de11 := fmt.Sprintf("%012s", stan)

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
	isomessage := iso8583.NewMessage(Spec87)

	now := time.Now().UTC()
	de7 := now.Format("0102150405")
	de11 := fmt.Sprintf("%012s", stan)

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
	isomessage := iso8583.NewMessage(Spec87)

	now := time.Now().UTC()
	de7 := now.Format("0102150405")
	de11 := fmt.Sprintf("%012s", stan)

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

func CreateIsoResLogon(msg []byte, bit48, stan string) ([]byte, error) {
	isoStr := string(msg)
	isomessage := iso8583.NewMessage(Spec87)
	err := isomessage.Unpack([]byte(isoStr))
	if err != nil {
		return nil, err
	}

	err = isomessage.Field(11, stan)
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

	// iso8583.Describe(isomessage, os.Stdout)

	return rawMessage, nil
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
	isomessage.Unpack([]byte("0210303801000E810400301000000000000000000067152710103000192030303030303230303039343132353230353030313230303436313101804944303133415050524F564544202D2030304E503232355445524D494E414C204944203A2036303036303030343B4D45524348414E54204944203A203630303630303630303630303630303B3B4341524420545950452020203A2044454249543B353337363839585858585858303039362843484950293B3B53414C453B4441544520203A2030332F30352F3230313820202020202054494D45203A2031343A30363A32353B4241544348203A2030303030303100261020360C0000000349697198871010360C000000034970719887"))

	iso8583.Describe(isomessage, os.Stdout)

	fmt.Println("OK")
}
