package models

import (
	"gorm.io/gorm"
)

type TrendImage struct {
	gorm.Model
	ID       uint   `json:"id"`
	Url      string `gorm:"type:varchar(255);default:null" json:"url"`
	TargetID uint   `gorm:"index" json:"target_id"`
}

func (TrendImage) TableName() string {
	return "trend_image"
}
