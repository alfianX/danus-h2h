package function

import (
	"os"
	"strings"
)

// IsRunningInContainer memeriksa apakah proses berjalan di dalam container Docker/Linux.
func IsRunningInContainer() bool {
	// Cek keberadaan file penanda /.dockerenv di root filesystem container.
	// File ini dibuat secara otomatis oleh Docker.
	if _, err := os.Stat("/.dockerenv"); err == nil {
		return true
	}

	// Membaca file cgroup yang mendeskripsikan grup kontrol dari proses saat ini
	content, err := os.ReadFile("/proc/self/cgroup")
	if err != nil {
		// Jika file tidak dapat dibaca, mungkin bukan environment Linux atau ada masalah
		return false
	}

	// Di dalam container, content biasanya berisi path/ID yang merujuk pada container.
	// Kita bisa mencari string "docker" atau "container"
	return strings.Contains(string(content), "docker") ||
		strings.Contains(string(content), "container") ||
		strings.Contains(string(content), "kubepods")
}
