package model

import "time"

type MonthSummary struct {
	ID               int64     `gorm:"primaryKey"`
	UserID           int64     `gorm:"not null;uniqueIndex:uk_user_ym"`
	Year             int       `gorm:"not null;uniqueIndex:uk_user_ym"`
	Month            int       `gorm:"not null;uniqueIndex:uk_user_ym"`
	TotalIncome      int64     `gorm:"type:bigint;not null;default:0"`
	TotalExpense     int64     `gorm:"type:bigint;not null;default:0"`
	AlipayIncome     int64     `gorm:"type:bigint;not null;default:0"`
	AlipayExpense    int64     `gorm:"type:bigint;not null;default:0"`
	WechatIncome     int64     `gorm:"type:bigint;not null;default:0"`
	WechatExpense    int64     `gorm:"type:bigint;not null;default:0"`
	InvestmentAmount int64     `gorm:"type:bigint;not null;default:0"`
	CreateTime       time.Time `gorm:"autoCreateTime"`
	UpdateTime       time.Time `gorm:"autoUpdateTime"`
	IsDeleted        bool      `gorm:"type:tinyint(1);not null;default:0"`
}
