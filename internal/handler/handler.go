package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"sync"
	"time"

	"github.com/alfianX/danus-h2h/config"
	"github.com/alfianX/danus-h2h/internal/repo"
	f "github.com/alfianX/danus-h2h/pkg/function"
	"github.com/alfianX/danus-h2h/pkg/iso"
	"github.com/alfianX/danus-h2h/pkg/license"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"

	_ "github.com/joho/godotenv/autoload"
)

const (
	HeaderLen          = 2
	MaxMessageLength   = 4096
	HexBase            = 16
	IntBitSize         = 64
	TPDUExpected       = "60"
	RCErrGeneral       = "96"
	RCErrLicense       = "15"
	RCErrInvalidTrx    = "12"
	RCErrFormatError   = "30"
	NetMgmtTypeLogon   = "101"
	NetMgmtTypeSignOn  = "001"
	NetMgmtTypeSignOff = "002"
	NetMgmtTypeNewKey  = "102"
	NetMgmtTypeEcho    = "301"
)

type Stan struct {
	Stan int64 `json:"stan"`
}

type StanManage struct {
	StanClient string
	Duration   time.Time
}

type HostResponse struct {
	Data []byte
	Err  error
}

type ReversalAdvice struct {
	Data []byte
}

type Handler struct {
	Config         config.Config
	sliceChan      chan []byte
	volumesn       string
	responseMap    sync.Map
	tpduConn       sync.Map
	mu             sync.Mutex
	stan           int64
	stanManage     map[string]StanManage
	reversalAdvice map[string]ReversalAdvice
	hostConnLock   sync.Mutex
	hostConn       net.Conn
	db             *gorm.DB
	Log            *logrus.Logger
	// lastPingSent     sync.Map
	// lastPongReceived sync.Map
}

func NewHandler(cnf config.Config, log *logrus.Logger) (*Handler, error) {
	sc := make(chan []byte)

	var volumesn string
	if f.IsRunningInContainer() {
		fmt.Println("docker nih")
		volumesn = license.GetVolumeDocker("/host-root")
	} else {
		fmt.Println("biasa nih")
		volumesn = license.GetVolume()
	}

	db, err := repo.InitDSN(cnf.Database)
	if err != nil {
		return nil, err
	}

	fileStan, err := os.Open("stan.json")
	if err != nil {
		return nil, err
	}
	defer fileStan.Close()

	byteValue, err := io.ReadAll(fileStan)
	if err != nil {
		return nil, err
	}

	var stan Stan
	err = json.Unmarshal(byteValue, &stan)
	if err != nil {
		return nil, err
	}

	h := Handler{
		Config:         cnf,
		sliceChan:      sc,
		volumesn:       volumesn,
		stan:           stan.Stan,
		stanManage:     make(map[string]StanManage),
		reversalAdvice: make(map[string]ReversalAdvice),
		db:             db,
		Log:            log,
		// lastPingSent:     sync.Map{},
		// lastPongReceived: sync.Map{},
	}

	// go h.checkConnectionStatus()

	return &h, nil
}

func (h *Handler) loadConfig() error {
	fileStan, err := os.Open("stan.json")
	if err != nil {
		return err
	}
	defer fileStan.Close()

	byteValue, err := io.ReadAll(fileStan)
	if err != nil {
		return err
	}

	var stan Stan
	if err := json.Unmarshal(byteValue, &stan); err != nil {
		return err
	}

	h.stan = stan.Stan

	// fmt.Println("Configuration reloaded!")
	return nil
}

func (h *Handler) editConfig(newStan int64) error {
	fileStan, err := os.OpenFile("stan.json", os.O_RDWR, 0644)
	if err != nil {
		return err
	}
	defer fileStan.Close()

	byteValue, err := io.ReadAll(fileStan)
	if err != nil {
		return err
	}

	var stan Stan
	err = json.Unmarshal(byteValue, &stan)
	if err != nil {
		return err
	}

	if newStan == 999999999999 {
		newStan = 1
	} else {
		newStan = newStan + 1
	}

	stan.Stan = newStan

	if err := fileStan.Truncate(0); err != nil {
		return err
	}

	if _, err := fileStan.Seek(0, 0); err != nil {
		return err
	}

	encoder := json.NewEncoder(fileStan)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(stan); err != nil {
		return err
	}

	return nil
}

func (h *Handler) handleErrorAndRespond(conn net.Conn, clientMsg, rc string, logMsg string, err error) {
	h.Log.Errorf("%s %v", logMsg, err)
	msgResponse, buildErr := iso.BuildErrorResponse(clientMsg, rc, 1)
	if buildErr != nil {
		h.Log.Errorf("failed to build error response for RC %s: %v", rc, buildErr)
		conn.Close() // Tutup koneksi jika bahkan error response tidak bisa dibuat
		return
	}
	h.sendBackHandler(msgResponse, conn)
}

func (h *Handler) CleanUpTimeoutStan() {
	ticker := time.NewTicker(1 * time.Minute) // Cek setiap 1 menit
	defer ticker.Stop()

	for range ticker.C {
		h.mu.Lock()
		for stanHost, data := range h.stanManage {
			if time.Since(data.Duration) > 2*time.Minute { // Contoh: hapus setelah 2 menit
				h.Log.Warnf("Cleaning up timed-out STAN: %s", stanHost)
				delete(h.stanManage, stanHost)
			}
		}
		h.mu.Unlock()
	}
}
