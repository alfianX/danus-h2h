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
	ID           int64      `gorm:"primaryKey;autoIncrement" json:"id"`
	Mti          string     `gorm:"type:varchar(4);index" json:"mti"`
	Procode      string     `gorm:"type:varchar(6)" json:"procode"`
	Tid          string     `gorm:"type:varchar(16);index" json:"tid"`
	Mid          string     `gorm:"type:varchar(20)" json:"mid"`
	Pan          string     `gorm:"type:varchar(20)" json:"pan"`
	Amount       int64      `gorm:"type:bigint" json:"amount"`
	TrxDate      *time.Time `json:"trx_date"`
	Stan         string     `gorm:"type:varchar(12)" json:"stan"`
	StanHost     string     `gorm:"type:varchar(12)" json:"stan_host"`
	Rrn          string     `gorm:"type:varchar(20);index" json:"rrn"`
	RrnHost      string     `gorm:"type:varchar(20)" json:"rrn_host"`
	MerchantName string     `gorm:"type:varchar(255)" json:"merchant_name"`
	ResponseCode string     `gorm:"type:varchar(4)" json:"response_code"`
	IsoReq       string     `gorm:"type:text" json:"iso_req"`
	IsoRes       string     `gorm:"type:text" json:"iso_res"`
	CreatedAt    time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time  `gorm:"autoUpdateTime" json:"updated_at"`
}

// Opsi: Menentukan nama tabel secara manual
func (TransactionHistory) TableName() string {
	return "transaction_history"
}

type Key struct {
	ID  int64  `gorm:"primaryKey;autoIncrement" json:"id"`
	Zmk string `gorm:"type:varchar(64);uniqueIndex:idx_module_zmk" json:"zmk"`
	Zpk string `gorm:"type:varchar(64);uniqueIndex:idx_module_zpk" json:"zpk"`
	Tmk string `gorm:"type:varchar(64);uniqueIndex:idx_module_tmk" json:"tmk"`
}

func (Key) TableName() string {
	return "key"
}

type TerminalKey struct {
	ID        int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	Tid       string    `gorm:"type:varchar(16);uniqueIndex" json:"tid"`
	Tpk       string    `gorm:"type:varchar(64)" json:"tpk"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

func (TerminalKey) TableName() string {
	return "terminal_key"
}
