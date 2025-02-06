package models

import "gorm.io/gorm"

type MyBooks struct {
	gorm.Model
	UserID   uint   `gorm:"type:bigint" json:"user_id"`
	Name     string `gorm:"type:varchar(255)" json:"name"`
	Describe string `gorm:"type:text" json:"describe"`
}

func (MyBooks) TableName() string {
	return "mybooks"
}
