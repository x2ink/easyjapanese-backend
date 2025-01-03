package models

import (
	"gorm.io/gorm"
)

type Image struct {
	gorm.Model
	ID       uint   `json:"id"`
	Url      string `gorm:"type:varchar(255);default:null" json:"url"`
	Target   string `gorm:"type:varchar(25);default:null" json:"target"`
	TargetID uint   `gorm:"index" json:"target_id"`
}

func (Image) TableName() string {
	return "image"
}
