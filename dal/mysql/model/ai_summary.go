package model

import "time"

type AISummary struct {
	ID          int64     `gorm:"primaryKey"`
	UserID      int64     `gorm:"not null;index:idx_ai_range"`
	StartYear   int       `gorm:"not null;index:idx_ai_range"`
	StartMonth  int       `gorm:"not null;index:idx_ai_range"`
	EndYear     int       `gorm:"not null;index:idx_ai_range"`
	EndMonth    int       `gorm:"not null;index:idx_ai_range"`
	AIAnalysis  string    `gorm:"type:text"`
	TriggerType int8      `gorm:"default:1"` // see consts.AITrigger*
	IsDeleted   bool      `gorm:"type:tinyint(1);not null;default:0;index:idx_ai_range"`
	CreateTime  time.Time `gorm:"autoCreateTime"`
	UpdateTime  time.Time `gorm:"autoUpdateTime"`
}
