package models

import (
	"gorm.io/gorm"
)

type Like struct {
	gorm.Model
	ID       uint   `json:"id"`
	Target   string ` json:"target"`
	TargetID uint   `gorm:"index"  json:"target_id"`
	UserID   uint   `gorm:"index" json:"user_id"`
	User     Users  `gorm:"foreignKey:UserID" json:"user"`
}

func (Like) TableName() string {
	return "like"
}
