package repo

import (
	"fmt"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func InitDSN(dsn string) (*gorm.DB, error) {
	fullDsn := fmt.Sprintf(
		"%s?charset=utf8mb4&parseTime=True&loc=Local",
		dsn,
	)

	db, err := gorm.Open(mysql.Open(fullDsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		fmt.Printf("failed to get underlying sql.DB: %v\n", err)
		return nil, err
	}

	// Atur jumlah koneksi maksimum yang bisa dibuka
	sqlDB.SetMaxOpenConns(100)
	// Atur jumlah koneksi idle maksimum
	sqlDB.SetMaxIdleConns(10)
	// Atur masa hidup maksimum koneksi
	sqlDB.SetConnMaxLifetime(5 * time.Minute)

	return db, nil
}
