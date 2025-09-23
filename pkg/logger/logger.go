package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// RotatingFileHook adalah dasar untuk hook file dengan rotasi harian
type RotatingFileHook struct {
	file       *os.File
	filePath   string // Base path of the log file (e.g., "server_errors")
	fileExt    string // File extension (e.g., ".log")
	lastOpened time.Time
	mu         sync.Mutex // Mutex untuk melindungi operasi file
	formatter  logrus.Formatter
	levels     []logrus.Level // Level yang ditangani oleh hook ini
}

// rotateFile melakukan rotasi file log jika tanggalnya berubah
func (hook *RotatingFileHook) rotateFile() error {
	now := time.Now()
	if hook.file == nil || hook.lastOpened.Year() != now.Year() ||
		hook.lastOpened.Month() != now.Month() || hook.lastOpened.Day() != now.Day() {

		if hook.file != nil {
			if err := hook.file.Close(); err != nil {
				// Gunakan fmt.Errorf karena logger utama belum tentu tersedia atau sudah dikonfigurasi
				return fmt.Errorf("failed to close old log file %s: %w", hook.file.Name(), err)
			}
		}

		datedFileName := fmt.Sprintf("%s_%s%s", hook.filePath, now.Format("2006-01-02"), hook.fileExt)

		logDir := filepath.Dir(datedFileName)
		if err := os.MkdirAll(logDir, 0755); err != nil {
			return fmt.Errorf("failed to create log directory %s: %w", logDir, err)
		}

		newFile, err := os.OpenFile(datedFileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			return fmt.Errorf("failed to open new log file %s: %w", datedFileName, err)
		}
		hook.file = newFile
		hook.lastOpened = now
		// Kita tidak bisa menggunakan logrus.Infof di sini karena mungkin belum terkonfigurasi
		fmt.Printf("INFO: Rotated log file to: %s\n", datedFileName)
	}
	return nil
}

// ErrorFileHook
type ErrorFileHook struct {
	*RotatingFileHook
}

func NewErrorFileHook(filePath, fileExt string, formatter logrus.Formatter) (*ErrorFileHook, error) {
	baseHook := &RotatingFileHook{
		filePath:  filePath,
		fileExt:   fileExt,
		formatter: formatter, // Gunakan formatter yang diberikan
		levels:    []logrus.Level{logrus.ErrorLevel, logrus.FatalLevel, logrus.PanicLevel},
	}
	return &ErrorFileHook{baseHook}, nil
}

func (hook *ErrorFileHook) Fire(entry *logrus.Entry) error {
	if entry.Level <= logrus.ErrorLevel {
		hook.mu.Lock()
		defer hook.mu.Unlock()

		if err := hook.rotateFile(); err != nil {
			return err
		}

		formatted, err := hook.formatter.Format(entry)
		if err != nil {
			return fmt.Errorf("failed to format log entry for error file: %w", err)
		}
		_, err = hook.file.Write(formatted)
		return err
	}
	return nil
}

func (hook *ErrorFileHook) Levels() []logrus.Level {
	return hook.levels
}

// DebugFileHook
type DebugFileHook struct {
	*RotatingFileHook
	TargetTag string
}

func NewDebugFileHook(filePath, fileExt string, formatter logrus.Formatter, targetTag string) (*DebugFileHook, error) {
	baseHook := &RotatingFileHook{
		filePath:  filePath,
		fileExt:   fileExt,
		formatter: formatter, // Gunakan formatter yang diberikan
		levels:    []logrus.Level{logrus.DebugLevel},
	}
	return &DebugFileHook{RotatingFileHook: baseHook, TargetTag: targetTag}, nil
}

func (hook *DebugFileHook) Fire(entry *logrus.Entry) error {
	if entry.Level == logrus.DebugLevel {
		if tag, ok := entry.Data["debug_tag"].(string); ok && tag == hook.TargetTag {
			hook.mu.Lock()
			defer hook.mu.Unlock()

			if err := hook.rotateFile(); err != nil {
				return err
			}

			formatted, err := hook.formatter.Format(entry)
			if err != nil {
				return fmt.Errorf("failed to format log entry for debug file: %w", err)
			}
			_, err = hook.file.Write(formatted)
			return err
		}
	}
	return nil
}

func (hook *DebugFileHook) Levels() []logrus.Level {
	return hook.levels
}

// --- FUNGSI INISIALISASI LOGGER UTAMA ---

// LoggerConfig adalah konfigurasi untuk logger
type LoggerConfig struct {
	EnableAllDebugFiles bool
	LogDir              string
	// Tambahkan konfigurasi lain jika diperlukan, misal level default
}

// InitLogger menginisialisasi logger dengan konfigurasi yang diberikan
// Fungsi ini harus dipanggil sekali di awal aplikasi (misal di main())
func InitLogger(cfg LoggerConfig) (*logrus.Logger, []*RotatingFileHook, error) {
	loggerInstance := logrus.New()

	// 1. Atur Formatter untuk logger utama (akan digunakan oleh hooks juga)
	consoleFormatter := &logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
		ForceColors:     true, // Hanya berlaku untuk konsol
	}
	loggerInstance.SetFormatter(consoleFormatter)
	loggerInstance.SetOutput(os.Stdout)
	loggerInstance.SetLevel(logrus.DebugLevel)

	fileJSONFormatter := &logrus.JSONFormatter{
		TimestampFormat: "2006-01-02 15:04:05",
		// FieldNameMap: untuk rename field, jika perlu
	}

	// Pilih formatter yang akan digunakan untuk hooks file
	fileFormatter := fileJSONFormatter

	// Slice untuk menyimpan semua base hook agar bisa ditutup filenya
	var allHooks []*RotatingFileHook

	// Pastikan direktori log ada sebelum membuka file
	if err := os.MkdirAll(cfg.LogDir, 0755); err != nil {
		return nil, nil, fmt.Errorf("failed to create log directory %s: %w", cfg.LogDir, err)
	}

	// 4. Tambahkan ErrorFileHook (selalu aktif)
	errorHook, err := NewErrorFileHook(filepath.Join(cfg.LogDir, "server_errors"), ".log", fileFormatter)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create error file hook: %w", err)
	}
	loggerInstance.AddHook(errorHook)
	allHooks = append(allHooks, errorHook.RotatingFileHook)

	// 5. Definisi path untuk 4 file debug yang sudah ditentukan
	debugFilePaths := []struct {
		BaseName  string
		TargetTag string
	}{
		{BaseName: "dl_in_debug", TargetTag: "dl_in"},
		{BaseName: "dl_out_debug", TargetTag: "dl_out"},
		{BaseName: "ul_in_debug", TargetTag: "ul_in"},
		{BaseName: "ul_out_debug", TargetTag: "ul_out"},
	}

	// 6. Tambahkan DebugFileHooks hanya jika EnableAllDebugFiles aktif
	if cfg.EnableAllDebugFiles {
		for _, df := range debugFilePaths {
			debugHook, err := NewDebugFileHook(filepath.Join(cfg.LogDir, df.BaseName), ".log", fileFormatter, df.TargetTag)
			if err != nil {
				return nil, nil, fmt.Errorf("failed to create debug file hook for %s: %w", df.BaseName, err)
			}
			loggerInstance.AddHook(debugHook)
			allHooks = append(allHooks, debugHook.RotatingFileHook)
		}
	} else {
		loggerInstance.Warn("All debug files are disabled by configuration.")
	}

	return loggerInstance, allHooks, nil
}

// CloseAllFileHooks menutup semua file log yang dibuka oleh hooks
// Fungsi ini harus dipanggil di akhir aplikasi (misal dengan defer di main())
func CloseAllFileHooks(logger *logrus.Logger, hooks []*RotatingFileHook) {
	for _, hook := range hooks {
		hook.mu.Lock()
		if hook.file != nil {
			if err := hook.file.Close(); err != nil {
				logger.Errorf("Failed to close log file %s: %v", hook.file.Name(), err)
			} else {
				logger.Warnf("Closed log file: %s", hook.file.Name())
			}
		}
		hook.mu.Unlock()
	}
}
