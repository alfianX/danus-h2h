package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/alfianX/danus-h2h/config"
	"github.com/alfianX/danus-h2h/internal/handler"
	"github.com/alfianX/danus-h2h/internal/transport"
	"github.com/alfianX/danus-h2h/pkg/logger"
)

var (
	version     = "1.1.2" // Application version
	showVersion = flag.Bool("version", false, "Display the application version")
)

func main() {
	flag.Parse()

	// If "--version" flag is provided, display version and exit
	if *showVersion {
		fmt.Printf("App Version: %s\n", version)
		return
	}

	// Create a root context that can be cancelled
	rootCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create a channel to listen for OS signals (e.g., Ctrl+C)
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		if err := run(rootCtx); err != nil {
			log.Fatalf("application exited with error: %v", err)
		}
		log.Println("Application stopped gracefully.")
	}()

	// Block until a signal is received
	sig := <-sigCh
	log.Printf("Received signal: %v, initiating shutdown...", sig)

	// Cancel the context to signal all goroutines to stop
	cancel()

	// Wait for the run goroutine to finish with a timeout
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// Goroutine finished within the timeout
		log.Println("All goroutines have finished. Exiting.")
	case <-time.After(5 * time.Second):
		// Timeout occurred, force exit
		log.Println("Graceful shutdown timed out, forcing exit.")
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	// iso.TestIso()

	cnf, err := config.NewParsedConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %+v", err)
	}

	logCfg := logger.LoggerConfig{
		EnableAllDebugFiles: cnf.Debug != 0,
		LogDir:              "log",
	}

	appLogger, allFileHooks, err := logger.InitLogger(logCfg)
	if err != nil {
		return fmt.Errorf("failed to initialize logger: %+v", err)
	}
	defer logger.CloseAllFileHooks(appLogger, allFileHooks)

	server, err := transport.NewTCP(appLogger, cnf, handler.NewHandler)
	if err != nil {
		return fmt.Errorf("failed to create new TCP server: %+v", err)
	}

	err = server.Run(ctx)
	if err != nil && !errors.Is(err, context.Canceled) {
		return fmt.Errorf("server exited with error: %+v", err)
	}

	return nil
}
