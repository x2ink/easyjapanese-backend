package models

import (
	"gorm.io/gorm"
)

type Message struct {
	gorm.Model
	Status   int
	ToID     uint
	FromID   uint
	Title    string
	Content  string
	Path     string
	Type     string
	Cover    string
	Tag      string
	FromUser Users `gorm:"foreignKey:FromID"`
	ToUser   Users `gorm:"foreignKey:ToID"`
}

func (Message) TableName() string {
	return "message"
}
