package models

import (
	"gorm.io/gorm"
)

type TrendLike struct {
	gorm.Model
	ID      uint  `json:"id"`
	TrendID uint  `gorm:"index"  json:"trend_id"`
	UserID  uint  `gorm:"index" json:"user_id"`
	User    Users `gorm:"foreignKey:UserID" json:"user"`
}

func (TrendLike) TableName() string {
	return "trend_like"
}
