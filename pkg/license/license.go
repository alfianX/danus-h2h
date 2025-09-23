package license

import (
	"bytes"
	"crypto/des"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// var (
// 	kernel32                  = syscall.NewLazyDLL("kernel32.dll")
// 	procGetVolumeInformationA = kernel32.NewProc("GetVolumeInformationA")
// )

func GetCurrentDir() string {
	ex, err := os.Executable()
	if err != nil {
		panic(err)
	}

	// Print the executable file path
	// fmt.Println("Executable path:", ex)

	// Get the directory of the executable
	exPath := filepath.Dir(ex)
	// fmt.Println("Executable directory:", exPath)
	return exPath
}

func CheckLicense(key string, volumesn string) string {
	src, _ := hex.DecodeString(key)
	keys, _ := hex.DecodeString(strings.ToUpper(volumesn) + "00000000")
	decrypted, err := DesDecrypt(src, keys)
	if err != nil {
		fmt.Println("Error:", err)
		return err.Error()
	}

	result := string(decrypted)

	lic := strings.Split(result, ";")
	if len(lic) > 2 {
		if lic[0] == "TMS2" {
			if lic[2] == "00" {
				// loc, _ := time.LoadLocation("Asia/Jakarta")
				currentTime := time.Now()
				gmtFormat := "20060102"
				dateString := currentTime.Format(gmtFormat)
				licenseData := lic[1]
				if lic[1] < dateString {
					return "License Expired " + licenseData[:4] + "-" + licenseData[4:6] + "-" + licenseData[6:8]
				} else {
					return "OK"
				}
			} else {
				return "License Not Valid"
			}
		} else {
			return "License Error"
		}
	} else {
		return "License Error"
	}
}

func DesDecrypt(src, key []byte) ([]byte, error) {
	block, err := des.NewCipher(key)
	if err != nil {
		return nil, err
	}
	out := make([]byte, len(src))
	dst := out
	bs := block.BlockSize()
	if len(src)%bs != 0 {
		return nil, errors.New("crypto/cipher: input not full blocks")
	}
	for len(src) > 0 {
		block.Decrypt(dst, src[:bs])
		src = src[bs:]
		dst = dst[bs:]
	}
	out = ZeroUnPadding(out)
	// out = PKCS5UnPadding(out)
	return out, nil
}

func ZeroUnPadding(origData []byte) []byte {
	return bytes.TrimFunc(origData,
		func(r rune) bool {
			return r == rune(0)
		})
}
