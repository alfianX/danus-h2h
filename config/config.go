package config

import (
	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	Mode              string `envconfig:"MODE" default:"debug"`
	ListenPort        int    `envconfig:"LISTEN" default:"88"`
	HostAddress       string `envconfig:"HOST_ADDRESS" required:"true"`
	Database          string `envconfig:"MYSQL_DSN" required:"true"`
	HsmAddress        string `envconfig:"HSM_ADDRESS" required:"true"`
	TimeoutTrx        int    `envconfig:"TIMEOUT_TRX" default:"60"`
	TimeoutInactivity string `envconfig:"TIMEOUT_INACTIVITY" default:"60"`
	Debug             int    `envconfig:"DEBUG_LOG" default:"0"`
	LicenseKey        string `envconfig:"LICENSE_KEY"`
}

func NewParsedConfig() (Config, error) {
	_ = godotenv.Load(".env")
	cnf := Config{}
	err := envconfig.Process("", &cnf)
	return cnf, err
}
