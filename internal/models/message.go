package models

import (
	"gorm.io/gorm"
)

type Message struct {
	gorm.Model
	Status    int
	ToID      uint
	FromID    uint
	CommentId uint
	Content   string
	Comment   Comment `gorm:"foreignKey:CommentId"`
	FromUser  Users   `gorm:"foreignKey:FromID"`
	ToUser    Users   `gorm:"foreignKey:ToID"`
}

func (Message) TableName() string {
	return "message"
}
