package models

import "gorm.io/gorm"

type Feedback struct {
	gorm.Model
	Content string `gorm:"type:text" json:"content"`
	Type    string `gorm:"type:varchar(100)" json:"type"`
	UserID  uint   `gorm:"type:bigint" json:"user_id"`
}

func (Feedback) TableName() string {
	return "feedback"
}
