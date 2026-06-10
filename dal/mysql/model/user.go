package model

import "time"

type User struct {
	ID         int64     `gorm:"primaryKey"`
	Username   string    `gorm:"size:64;uniqueIndex:uk_username"`
	Password   string    `gorm:"size:128"`
	CreateTime time.Time `gorm:"autoCreateTime"`
	UpdateTime time.Time `gorm:"autoUpdateTime"`
	IsDeleted  bool      `gorm:"type:tinyint(1);not null;default:0;index"`
}
