package models

import (
	"gorm.io/gorm"
)

type TrendLike struct {
	gorm.Model
	ID       uint  `json:"id"`
	TargetID uint  `gorm:"index"  json:"target_id"`
	UserID   uint  `gorm:"index" json:"user_id"`
	User     Users `gorm:"foreignKey:UserID" json:"user"`
}

func (TrendLike) TableName() string {
	return "trend_like"
}
