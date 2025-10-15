package function

import (
	"encoding/hex"
	"errors"
	"fmt"
	"net"
	"strings"
)

func SendMessageToHsm(IPPORT, message string) (string, error) {
	iso, _ := hex.DecodeString(message)

	tcpServer, err := net.ResolveTCPAddr("tcp", IPPORT)
	if err != nil {
		return "", err
	}

	conn, err := net.DialTCP("tcp", nil, tcpServer)
	if err != nil {
		return "", err
	}
	defer conn.Close()

	_, err = conn.Write(iso)
	if err != nil {
		return "", err
	}

	received := make([]byte, 1024)
	bytesRead, err := conn.Read(received)
	if err != nil {
		return "", err
	}

	conn.Close()

	messageHost := hex.EncodeToString(received[:bytesRead])

	return messageHost, nil
}

func HSMSaveZPK(IPPORT, zmk, zpk string) (string, error) {
	command := "GIKUFA" + zmk + "U" + zpk
	len := len(command)
	lenHex := strings.ToUpper(fmt.Sprintf("%04x", len))
	message := lenHex + hex.EncodeToString([]byte(command))

	response, err := SendMessageToHsm(IPPORT, message)
	if err != nil {
		return "", err
	}

	resByte, err := hex.DecodeString(response)
	if err != nil {
		return "", err
	}

	hsmResponse := string(resByte[8:10])
	if hsmResponse != "00" {
		return "", errors.New("HSM response invalid")
	}

	return string(resByte[10:43]), nil
}

func HSMGenerateKey(IPPORT, tmk string) (string, string, error) {
	command := "0000HC" + tmk + ";XU0"
	len := len(command)
	lenHex := strings.ToUpper(fmt.Sprintf("%04x", len))
	message := lenHex + hex.EncodeToString([]byte(command))

	response, err := SendMessageToHsm(IPPORT, message)
	if err != nil {
		return "", "", err
	}

	resByte, err := hex.DecodeString(response)
	if err != nil {
		return "", "", err
	}

	hsmResponse := string(resByte[8:10])
	if hsmResponse != "00" {
		return "", "", errors.New("hsm response invalid")
	}

	twk := string(resByte[11:43])
	tpk := string(resByte[43:76])

	return twk, tpk, nil
}

func HSMTranslatePin(IPPORT, tpk, zpk, pinBlock, panParsed string) (string, error) {
	command := "GIKUCA" + tpk + zpk + "12" + pinBlock + "0101" + panParsed
	len := len(command)
	lenHex := strings.ToUpper(fmt.Sprintf("%04x", len))
	message := lenHex + hex.EncodeToString([]byte(command))
	fmt.Println("translate pin : " + message)

	response, err := SendMessageToHsm(IPPORT, message)
	if err != nil {
		return "", err
	}

	resByte, err := hex.DecodeString(response)
	if err != nil {
		return "", err
	}

	hsmResponse := string(resByte[8:10])
	if hsmResponse != "00" {
		return "", errors.New("HSM response invalid")
	}

	return string(resByte[12:28]), nil
}
