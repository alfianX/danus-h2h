package transport

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/alfianX/danus-h2h/config"
	"github.com/alfianX/danus-h2h/internal/handler"
	"github.com/sirupsen/logrus"
)

type TCP struct {
	config    config.Config
	maxClient int
	handler   *handler.Handler
	log       *logrus.Logger
}

func NewTCP(appLogger *logrus.Logger, cnf config.Config, newHandlerFunc func(config.Config, *logrus.Logger) (*handler.Handler, error)) (*TCP, error) {
	h, err := newHandlerFunc(cnf, appLogger)
	if err != nil {
		appLogger.Errorf("Failed to create handler: %v", err)
		return nil, err
	}

	go h.CleanUpTimeoutStan()

	s := TCP{
		config:    cnf,
		maxClient: 1000,
		handler:   h,
		log:       appLogger,
	}

	return &s, nil
}

func (s *TCP) Run(ctx context.Context) error {
	go s.handler.ConnectToHost()

	s.log.Infof("Server listen on port: %d", s.config.ListenPort)
	serverAddress := fmt.Sprintf("0.0.0.0:%d", s.config.ListenPort)
	listener, err := net.Listen("tcp", serverAddress)
	if err != nil {
		s.log.Errorf("Failed to listen on %s: %v", serverAddress, err)
		return err
	}
	defer listener.Close()

	sem := make(chan struct{}, s.maxClient)
	waitingQueue := make(chan net.Conn, s.maxClient)
	var wg sync.WaitGroup

	go func() {
		for conn := range waitingQueue {
			select {
			case sem <- struct{}{}:
				wg.Add(1)
				go s.handler.ClientHandler(conn, sem, &wg)
			case <-ctx.Done():
				fmt.Printf("Server shutting down, dropping queued client: %v\n", conn.RemoteAddr())
				conn.Close()
				return
			}
		}
		wg.Wait()
	}()

	for {
		select {
		case <-ctx.Done():
			s.log.Infof("Server shutting down due to context cancellation.")
			listener.Close()
			close(waitingQueue)
			wg.Wait()
			s.log.Infof("All client handlers finished. Server stopped.")
			return ctx.Err()
		default:
			conn, err := listener.Accept()
			if err != nil {
				if opErr, ok := err.(*net.OpError); ok && opErr.Op == "accept" && opErr.Net == "tcp" && opErr.Err.Error() == "use of closed network connection" {
					s.log.Infof("Listener closed, stopping accept loop for graceful shutdown.")
					return nil
				}

				if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
					s.log.Warnf("Temporary error accepting connection: %v. Retrying...", err)
					time.Sleep(100 * time.Millisecond)
					continue
				}

				s.log.Errorf("Fatal error accepting connection: %v, stopping server!", err)
				return err
			}

			select {
			case waitingQueue <- conn:

			case <-ctx.Done():
				s.log.Warnf("Server shutting down, dropping new connection: %v\n", conn.RemoteAddr())
				conn.Close()
				return ctx.Err()
			}
		}
	}
}
