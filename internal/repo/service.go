package repo

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"
)

func ServiceGetAddress(ctx context.Context, db *gorm.DB, prefix string) (Services, error) {
	var service Services
	result := db.WithContext(ctx).Where("service_prefix = ?", prefix).First(&service)

	return service, result.Error
}

func TransactionHistorySave(ctx context.Context, db *gorm.DB, data *TransactionHistory) (int64, error) {
	result := db.WithContext(ctx).Select(
		"mti",
		"procode",
		"mid",
		"tid",
		"pan",
		"amount",
		"trx_date",
		"stan",
		"stan_host",
		"rrn",
		"merchant_name",
		"iso_req",
		"created_at",
	).Create(&data)

	return data.ID, result.Error
}

func TransactionHistoryUpdateStanHost(ctx context.Context, db *gorm.DB, data *TransactionHistory) error {
	result := db.WithContext(ctx).Model(&TransactionHistory{ID: data.ID}).Updates(&TransactionHistory{
		StanHost:  data.StanHost,
		UpdatedAt: data.UpdatedAt,
	})

	return result.Error
}

func TransactionHistoryUpdateResponse(ctx context.Context, db *gorm.DB, data *TransactionHistory) error {
	result := db.WithContext(ctx).Model(&TransactionHistory{ID: data.ID}).Updates(&TransactionHistory{
		ResponseCode: data.ResponseCode,
		IsoRes:       data.IsoRes,
		UpdatedAt:    data.UpdatedAt,
	})

	return result.Error
}

func TransactionHistoryGetDataWD(ctx context.Context, db *gorm.DB, data *TransactionHistory) (TransactionHistory, error) {
	var trxHistory TransactionHistory
	result := db.WithContext(ctx).
		Where(`mid = ? AND tid = ? AND amount = ? AND trx_date = ? AND stan = ? AND response_code = ?`,
			data.Mid, data.Tid, data.Amount, data.TrxDate, data.Stan, data.ResponseCode).
		First(&trxHistory)

	return trxHistory, result.Error
}

func TransactionGetStanHost(ctx context.Context, db *gorm.DB, data *TransactionHistory) (string, error) {
	var trxHistory TransactionHistory
	result := db.WithContext(ctx).Select("stan_host").Order("stan_host DESC").
		Where("mti = ? AND procode = ? AND amount = ? AND stan = ? AND tid = ? AND  mid = ?",
			data.Mti, data.Procode, data.Amount, data.Stan, data.Tid, data.Mid).
		First(&trxHistory)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return "", nil
		}
		return "", result.Error
	}

	return trxHistory.StanHost, nil
}

func KeyGetZMK(ctx context.Context, db *gorm.DB) (string, error) {
	var key Key
	result := db.WithContext(ctx).Select("zmk").First(&key)

	return key.Zmk, result.Error
}

func KeyUpdateZPK(ctx context.Context, db *gorm.DB, zpk string) error {
	result := db.WithContext(ctx).Session(&gorm.Session{AllowGlobalUpdate: true}).
		Model(&Key{}).Update("zpk", zpk)

	return result.Error
}

func KeyGetZPK(ctx context.Context, db *gorm.DB) (string, error) {
	var key Key
	result := db.WithContext(ctx).Select("zpk").First(&key)

	return key.Zpk, result.Error
}

func KeyGetTMK(ctx context.Context, db *gorm.DB) (string, error) {
	var key Key
	result := db.WithContext(ctx).Select("tmk").First(&key)

	return key.Tmk, result.Error
}

func TerminalKeySave(ctx context.Context, db *gorm.DB, data *TerminalKey) error {
	var count int64
	resultCnt := db.WithContext(ctx).Model(&TerminalKey{}).Where(`tid = ?`, data.Tid).Count(&count)

	if resultCnt.Error != nil {
		return resultCnt.Error
	}

	if count > 0 {
		resultUpdate := db.WithContext(ctx).Model(&TerminalKey{}).Where(`tid = ?`, data.Tid).Updates(&TerminalKey{
			Tpk:       data.Tpk,
			UpdatedAt: time.Now(),
		})

		if resultUpdate.Error != nil {
			return resultCnt.Error
		}
	} else {
		result := db.WithContext(ctx).Select(
			"tid",
			"tpk",
			"created_at",
		).Create(&data)

		if result.Error != nil {
			return result.Error
		}
	}

	return nil
}

func TerminalKeyUpdate(ctx context.Context, db *gorm.DB, data *TerminalKey) error {
	result := db.WithContext(ctx).Model(&TerminalKey{ID: data.ID}).Updates(&TerminalKey{
		Tpk:       data.Tpk,
		UpdatedAt: data.UpdatedAt,
	})

	return result.Error
}

func TerminalKeyGetTPK(ctx context.Context, db *gorm.DB, data *TerminalKey) (string, error) {
	var terminalKey TerminalKey
	result := db.WithContext(ctx).Select("tpk").Where("tid = ?", data.Tid).Find(&terminalKey)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return "", nil
		}
		return "", result.Error
	}

	return terminalKey.Tpk, result.Error
}
