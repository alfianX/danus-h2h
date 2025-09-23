package transport

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/alfianX/danus-h2h/config"
	"github.com/alfianX/danus-h2h/internal/handler"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestNewTCP(t *testing.T) {
	// Skenario 1: Test jika handler berhasil dibuat
	t.Run("Success", func(t *testing.T) {
		// Buat mock handler yang akan mengembalikan handler dummy
		mockNewHandlerFunc := func(cnf config.Config, appLogger *logrus.Logger) (*handler.Handler, error) {
			return &handler.Handler{}, nil
		}

		// Konfigurasi dummy
		cnf := config.Config{ListenPort: 8080}
		appLogger := logrus.New()

		// Panggil NewTCP dengan mock handler
		tcp, err := NewTCP(appLogger, cnf, mockNewHandlerFunc)

		// Verifikasi hasil
		assert.NoError(t, err)
		assert.NotNil(t, tcp)
		assert.Equal(t, cnf, tcp.config)
		assert.Equal(t, 1000, tcp.maxClient)
		assert.NotNil(t, tcp.handler)
		assert.Equal(t, appLogger, tcp.log)
	})

	// Skenario 2: Test jika handler gagal dibuat
	t.Run("HandlerCreationFails", func(t *testing.T) {
		// Buat error yang diharapkan dari mock
		expectedErr := errors.New("failed to create handler")

		// Buat mock handler yang akan mengembalikan error
		mockNewHandlerFunc := func(cnf config.Config, appLogger *logrus.Logger) (*handler.Handler, error) {
			return nil, expectedErr
		}

		// Konfigurasi dummy
		cnf := config.Config{ListenPort: 8080}
		appLogger := logrus.New()

		// Panggil NewTCP dengan mock handler
		tcp, err := NewTCP(appLogger, cnf, mockNewHandlerFunc)

		// Verifikasi hasil
		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)
		assert.Nil(t, tcp)
	})
}

// MockHost adalah implementasi host palsu yang menerima koneksi dari server kita.
func MockHost(t *testing.T, hostPort int, messageToReceive []byte, messageToSend []byte) {
	// Buat listener untuk host palsu
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", hostPort))
	if err != nil {
		t.Fatalf("failed to start mock host: %v", err)
	}
	defer listener.Close()

	// Tunggu koneksi dari server
	conn, err := listener.Accept()
	if err != nil {
		t.Fatalf("failed to accept connection on mock host: %v", err)
	}
	defer conn.Close()

	// 1. Terima pesan dari server
	receivedHeader := make([]byte, 2)
	_, err = io.ReadFull(conn, receivedHeader)
	if err != nil {
		t.Errorf("mock host failed to read header: %v", err)
		return
	}
	// Asumsi pesan yang diterima adalah panjang header 2 byte
	length := int(receivedHeader[0])<<8 | int(receivedHeader[1])
	receivedMessage := make([]byte, length)
	_, err = io.ReadFull(conn, receivedMessage)
	if err != nil {
		t.Errorf("mock host failed to read message body: %v", err)
		return
	}

	// Verifikasi pesan yang diterima
	assert.Equal(t, messageToReceive, receivedMessage, "mock host received unexpected message")

	// 2. Kirim respons kembali
	_, err = conn.Write(messageToSend)
	if err != nil {
		t.Errorf("mock host failed to write response: %v", err)
	}
}

// MockNewHandler adalah stub untuk handler.NewHandler yang akan kita gunakan.
func MockNewHandler(cnf config.Config, appLogger *logrus.Logger) (*handler.Handler, error) {
	// Kita akan menggunakan handler sungguhan, tapi dengan mock DB.
	// Untuk test ini, kita bisa mengembalikan handler kosong untuk kesederhanaan.
	return &handler.Handler{
		Config: cnf,
		Log:    appLogger,
		// Mock dependensi lain jika perlu (db, hsm, dll.)
	}, nil
}

// TestServerClientE2E menguji alur lengkap dari koneksi klien hingga server merespons.
func TestServerClientE2E(t *testing.T) {
	// Persiapan
	clientPort := 8787
	hostPort := 9090

	// Konfigurasi server dengan host palsu
	cnf := config.Config{
		ListenPort:  clientPort,
		HostAddress: fmt.Sprintf("localhost:%d", hostPort),
	}
	appLogger := logrus.New()

	// Pesan yang akan dikirim dari klien dan diterima oleh host palsu
	clientMessage := []byte("Hello from client")

	// Pesan yang akan dikirim dari host palsu dan diterima oleh server
	hostResponse := []byte("Hello from host")

	// Persiapkan sebuah channel untuk memberi tahu bahwa test selesai
	done := make(chan struct{})

	// 1. Jalankan host palsu dalam goroutine terpisah
	go func() {
		MockHost(t, hostPort, clientMessage, hostResponse)
		close(done) // Beri sinyal bahwa host palsu telah selesai
	}()

	// 2. Jalankan server TCP dalam goroutine terpisah
	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		tcpServer, err := NewTCP(appLogger, cnf, MockNewHandler)
		assert.NoError(t, err)
		tcpServer.Run(ctx)
	}()

	// Beri server waktu untuk memulai
	time.Sleep(100 * time.Millisecond)

	// 3. Jalankan klien test
	clientConn, err := net.Dial("tcp", fmt.Sprintf("localhost:%d", clientPort))
	assert.NoError(t, err, "failed to connect to server")
	defer clientConn.Close()

	// Kirim pesan dari klien
	_, err = clientConn.Write(clientMessage)
	assert.NoError(t, err, "failed to write to server")

	// Terima respons dari server
	response := make([]byte, len(hostResponse))
	_, err = io.ReadFull(clientConn, response)
	assert.NoError(t, err, "failed to read response from server")

	// Verifikasi respons
	assert.Equal(t, hostResponse, response, "client received unexpected response")

	// Tunggu sampai host palsu selesai
	<-done

	// 4. Lakukan graceful shutdown pada server
	cancel()
	wg.Wait()
}
