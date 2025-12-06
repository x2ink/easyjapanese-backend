package models

import "gorm.io/gorm"

type Notes struct {
	gorm.Model
	UserID     uint   `gorm:"type:bigint" json:"user_id"`
	User       Users  `gorm:"foreignKey:UserID;references:ID" json:"user"`
	Content    string `gorm:"type:text" json:"content"`
	TargetID   uint   `json:"target_id"`
	TargetType string `json:"target_type"`
}

func (Notes) TableName() string {
	return "notes"
}
