package repo

import "time"

type Services struct {
	ID             int64     `json:"id"`
	ServiceName    string    `json:"service_name"`
	ServicePrefix  string    `json:"service_prefix"`
	ServiceAddress string    `json:"service_address"`
	ServiceMethod  string    `json:"service_method"`
	UpdatedAt      time.Time `gorm:"autoUpdateTime:false" json:"updated_at"`
}

type TransactionHistory struct {
	ID           int64      `json:"id"`
	Mti          string     `json:"mti"`
	Procode      string     `json:"procode"`
	Tid          string     `json:"tid"`
	Mid          string     `json:"mid"`
	Pan          string     `json:"pan"`
	Amount       int64      `json:"amount"`
	TrxDate      *time.Time `gorm:"autoUpdateTime:false" json:"trx_date"`
	Stan         string     `json:"stan"`
	StanHost     string     `json:"stan_host"`
	Rrn          string     `json:"rrn"`
	RrnHost      string     `json:"rrn_host"`
	MerchantName string     `json:"merchant_name"`
	ResponseCode string     `json:"response_code"`
	IsoReq       string     `json:"iso_req"`
	IsoRes       string     `json:"iso_res"`
	CreatedAt    time.Time  `gorm:"autoUpdateTime:false" json:"created_at"`
	UpdatedAt    time.Time  `gorm:"autoUpdateTime:false" json:"updated_at"`
}

func (TransactionHistory) TableName() string {
	return "transaction_history"
}

type Key struct {
	Zmk string `json:"zmk"`
	Zpk string `json:"zpk"`
	Tmk string `json:"tmk"`
}

func (Key) TableName() string {
	return "key"
}

type TerminalKey struct {
	ID        int64     `json:"id"`
	Tid       string    `json:"tid"`
	Tpk       string    `json:"tpk"`
	CreatedAt time.Time `gorm:"autoCreateTime:false" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime:false" json:"updated_at"`
}

func (TerminalKey) TableName() string {
	return "terminal_key"
}
