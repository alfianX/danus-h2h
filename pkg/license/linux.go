//go:build linux
// +build linux

package license

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"

	"golang.org/x/sys/unix"
)

func GetVolume() string {
	var stat unix.Statfs_t
	err := unix.Statfs("/", &stat)
	if err != nil {
		fmt.Println("Error:", err)
		return ""
	}

	var volumeID uint64
	// Debug and adapt based on your platform's Fsid structure
	if len(stat.Fsid.Val) >= 2 {
		volumeID = uint64(stat.Fsid.Val[0])<<32 | uint64(stat.Fsid.Val[1])
	} else {
		fmt.Println("Fsid structure incompatible.")
		return ""
	}

	// Convert the volume ID to a hexadecimal string
	dongleKey := make([]byte, 8)
	binary.LittleEndian.PutUint64(dongleKey, volumeID)

	return hex.EncodeToString(dongleKey)[:8]
}

func GetVolumeDocker(rootPath string) string {
	var stat unix.Statfs_t
	err := unix.Statfs(rootPath, &stat)
	if err != nil {
		fmt.Println("Error:", err)
		return ""
	}

	var volumeID uint64
	// Debug and adapt based on your platform's Fsid structure
	if len(stat.Fsid.Val) >= 2 {
		volumeID = uint64(stat.Fsid.Val[0])<<32 | uint64(stat.Fsid.Val[1])
	} else {
		fmt.Println("Fsid structure incompatible.")
		return ""
	}

	// Convert the volume ID to a hexadecimal string
	dongleKey := make([]byte, 8)
	binary.LittleEndian.PutUint64(dongleKey, volumeID)

	return hex.EncodeToString(dongleKey)[:8]
}
