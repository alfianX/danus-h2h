package repo

import (
	"database/sql"
	"fmt"

	"github.com/alfianX/danus-h2h/config"
	"github.com/sirupsen/logrus"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func InitDSN(log *logrus.Logger, cnf config.Config) (*gorm.DB, error) {
	user, pass, host, port, dbName := cnf.DbUser, cnf.DbPass, cnf.DbHost, cnf.DbPort, cnf.DbName

	// 1. KONEKSI AWAL (Tanpa Nama Database)
	// Kita connect ke MySQL server secara umum dulu
	dsnNoDB := fmt.Sprintf("%s:%s@tcp(%s:%s)/", user, pass, host, port)
	rawDB, err := sql.Open("mysql", dsnNoDB)
	if err != nil {
		log.Errorf("Failed to connect to MySQL server: %v", err)
		return nil, err
	}
	defer rawDB.Close()

	// 2. CREATE DATABASE JIKA BELUM ADA
	_, err = rawDB.Exec(fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s", dbName))
	if err != nil {
		log.Errorf("Failed to create database: %v", err)
		return nil, err
	}

	// 3. KONEKSI GORM (Dengan Nama Database)
	// Sekarang database sudah pasti ada, kita bisa pakai GORM seperti biasa
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		user, pass, host, port, dbName)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		// Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Errorf("Failed to connect to MySQL database with GORM: %v", err)
		return nil, err
	}

	err = db.AutoMigrate(
		&TransactionHistory{},
		&Key{},
		&TerminalKey{},
	)
	if err != nil {
		log.Errorf("Failed to auto-migrate database: %v", err)
		return nil, err
	}

	defaultKeys := []Key{
		{ID: 1, Zmk: "U3259E52B9A5026441005A22168804E53", Zpk: "U9624E16AAB3F7D0788E350999C75D40C", Tmk: "U7AA18609E80329C9ADBBE731368F068B"},
	}
	for _, k := range defaultKeys {
		db.Clauses(clause.OnConflict{DoNothing: true}).Create(&k)
	}

	return db, nil
}
