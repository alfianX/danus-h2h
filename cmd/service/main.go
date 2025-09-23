package main

import (
	"context"
	"errors"
	"flag" // Import package flag
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/kardianos/service" // Impor library service

	"github.com/alfianX/danus-h2h/config"
	"github.com/alfianX/danus-h2h/internal/handler"
	"github.com/alfianX/danus-h2h/internal/transport"
	"github.com/alfianX/danus-h2h/pkg/logger"
	// Hapus import iso jika iso.TestIso() tidak lagi digunakan
	// "github.com/alfianX/danus-h2h/pkg/iso"
)

var (
	version     = "1.0.0" // Application version
	showVersion = flag.Bool("version", false, "Display the application version")
)

// Program adalah struct yang mengimplementasikan interface service.Service
type program struct {
	exit      chan struct{}      // Channel untuk sinyal keluar
	logger    service.Logger     // Logger untuk service
	appCtx    context.Context    // Context untuk aplikasi utama
	appCancel context.CancelFunc // Fungsi untuk membatalkan context aplikasi
}

// Start dipanggil saat service dimulai
func (p *program) Start(s service.Service) error {
	p.logger.Info("Service starting...")
	p.exit = make(chan struct{})
	p.appCtx, p.appCancel = context.WithCancel(context.Background()) // Inisialisasi context aplikasi

	// Jalankan logika utama service di goroutine terpisah
	go p.run()
	return nil
}

// run adalah logika utama dari service Anda
func (p *program) run() {
	p.logger.Info("Service started. Running application logic...")

	// --- Logika dari fungsi 'run' main.go Anda dipindahkan ke sini ---
	cnf, err := config.NewParsedConfig()
	if err != nil {
		p.logger.Errorf("Failed to load config: %v", err)
		// Jika ada error fatal saat startup, sinyal untuk stop service
		p.appCancel()
		return
	}

	logCfg := logger.LoggerConfig{
		EnableAllDebugFiles: cnf.Debug != 0,
		LogDir:              "log",
	}

	appLogger, allFileHooks, err := logger.InitLogger(logCfg)
	if err != nil {
		p.logger.Errorf("Failed to initialize logger: %v", err)
		p.appCancel()
		return
	}
	// Pastikan file hooks ditutup saat service berhenti
	defer logger.CloseAllFileHooks(appLogger, allFileHooks)

	server, err := transport.NewTCP(appLogger, cnf, handler.NewHandler)
	if err != nil {
		p.logger.Errorf("Failed to create new TCP server: %v", err)
		p.appCancel()
		return
	}

	// Gunakan context aplikasi untuk menjalankan server
	err = server.Run(p.appCtx)
	if err != nil && !errors.Is(err, context.Canceled) {
		p.logger.Errorf("Server exited with error: %v", err)
	}
	// --- Akhir dari logika 'run' main.go Anda ---

	p.logger.Info("Application logic finished. Waiting for exit signal...")
	// Tunggu sinyal dari Stop() atau pembatalan context dari dalam run()
	select {
	case <-p.exit:
		// Service dihentikan oleh OS
	case <-p.appCtx.Done():
		// Aplikasi berhenti karena error internal atau pembatalan context
		p.logger.Info("Application context cancelled, initiating service stop.")
	}
	p.logger.Info("Service run goroutine exiting.")
}

// Stop dipanggil saat service dihentikan
func (p *program) Stop(s service.Service) error {
	p.logger.Info("Service received stop signal. Stopping application...")
	close(p.exit) // Kirim sinyal keluar ke goroutine run
	p.appCancel() // Batalkan context aplikasi untuk menghentikan server

	// Beri waktu sebentar untuk goroutine run berhenti
	// Dalam aplikasi nyata, Anda mungkin ingin menunggu goroutine run selesai dengan sync.WaitGroup
	time.Sleep(1 * time.Second)
	p.logger.Info("Service stopped.")
	return nil
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Printf("Warning: .env file not loaded: %v", err)
		// Jangan fatal, karena di produksi mungkin tidak ada .env
	}

	flag.Parse() // Parse command-line flags

	// If "--version" flag is provided, display version and exit
	if *showVersion {
		fmt.Printf("App Version: %s\n", version)
		return
	}

	svcConfig := &service.Config{
		Name:        "DanusH2HService",                         // Nama service Anda
		DisplayName: "Danus H2H Application Service",           // Nama tampilan service di OS
		Description: "H2H application for payment processing.", // Deskripsi service
	}

	prg := &program{}
	s, err := service.New(prg, svcConfig)
	if err != nil {
		log.Fatal(err)
	}

	errs := make(chan error, 5)
	prg.logger, err = s.Logger(errs) // Gunakan prg.logger
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		for {
			err := <-errs
			if err != nil {
				log.Print(err)
			}
		}
	}()

	// Tangani perintah dari command line (install, uninstall, start, stop, run)
	if len(os.Args) > 1 {
		verb := os.Args[1]
		switch verb {
		case "install":
			err = s.Install()
			if err != nil {
				log.Fatalf("Failed to install service: %v", err)
			}
			fmt.Println("Service installed successfully.")
		case "uninstall":
			err = s.Uninstall()
			if err != nil {
				log.Fatalf("Failed to uninstall service: %v", err)
			}
			fmt.Println("Service uninstalled successfully.")
		case "start":
			err = s.Start()
			if err != nil {
				log.Fatalf("Failed to start service: %v", err)
			}
			fmt.Println("Service started successfully.")
		case "stop":
			err = s.Stop()
			if err != nil {
				log.Fatalf("Failed to stop service: %v", err)
			}
			fmt.Println("Service stopped successfully.")
		case "status": // Tambahkan case untuk perintah status
			status, err := s.Status()
			if err != nil {
				log.Fatalf("Failed to get service status: %v", err)
			}
			switch status {
			case service.StatusRunning:
				fmt.Println("Service is running.")
			case service.StatusStopped:
				fmt.Println("Service is stopped.")
			case service.StatusUnknown:
				fmt.Println("Service status is unknown.")
			default:
				fmt.Printf("Service is in an unexpected state: %v\n", status)
			}
		case "run":
			fmt.Println("Running service in interactive mode. Press Ctrl+C to stop.")

			// Buat context untuk menangani sinyal OS saat di mode interaktif
			ctx, cancel := context.WithCancel(context.Background())
			signalCh := make(chan os.Signal, 1)
			signal.Notify(signalCh, syscall.SIGINT, syscall.SIGTERM)

			go func() {
				<-signalCh
				fmt.Println("\nReceived interrupt signal. Shutting down...")
				cancel() // Batalkan context
			}()

			// Atur context aplikasi untuk mode interaktif
			prg.appCtx = ctx
			prg.appCancel = cancel

			err = s.Run() // Panggil Run untuk menjalankan service di foreground
			if err != nil {
				log.Fatalf("Failed to run service: %v", err)
			}
			fmt.Println("Service stopped.")
		default:
			fmt.Printf("Unknown command: %s\n", verb)
			fmt.Printf("Usage: %s [install|uninstall|start|stop|run|--version]\n", os.Args[0])
		}
		return
	}

	// Jika tidak ada argumen, jalankan service (ini akan dipanggil oleh OS saat service dimulai)
	err = s.Run()
	if err != nil {
		log.Fatalf("Failed to run service: %v", err)
	}
}
