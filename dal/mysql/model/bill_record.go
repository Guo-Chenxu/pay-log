package model

import "time"

type BillRecord struct {
	ID              int64     `gorm:"primaryKey"`
	UserID          int64     `gorm:"not null;index:idx_user_time;index:idx_user_dedup"`
	Channel         int8      `gorm:"not null;index:idx_user_dedup"` // see consts.Channel*
	TransactionTime time.Time `gorm:"not null;index:idx_user_time;index:idx_user_dedup"`
	Counterparty    string    `gorm:"size:128"`
	Description     string    `gorm:"size:256;index:idx_user_dedup"`
	Category        string    `gorm:"size:64"`
	BillType        int8      `gorm:"not null"`                                  // see consts.BillType*
	Amount          int64     `gorm:"type:bigint;not null;index:idx_user_dedup"` // cents
	PaymentMethod   string    `gorm:"size:64"`
	Status          string    `gorm:"size:32"`
	OrderNo         string    `gorm:"size:64"`
	MerchantOrderNo string    `gorm:"size:64"`
	Remark          string    `gorm:"size:256"`
	IsInvestment    bool      `gorm:"type:tinyint(1);not null;default:0"`
	CreateTime      time.Time `gorm:"autoCreateTime"`
	UpdateTime      time.Time `gorm:"autoUpdateTime"`
	IsDeleted       bool      `gorm:"type:tinyint(1);not null;default:0"`
}
