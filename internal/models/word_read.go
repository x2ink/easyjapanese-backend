package models

import (
	"gorm.io/gorm"
)

type WordRead struct {
	gorm.Model
	UserID uint   `gorm:"column:user_id"`
	Voice  string `gorm:"column:voice;size:255"`
	Like   uint   `gorm:"column:like;default:0"`
	Status uint   `gorm:"column:status;default:0"`
	WordID uint   `gorm:"column:word_id" json:"word_id"`
	User   Users  `gorm:"foreignKey:UserID" json:"user"`
}

func (WordRead) TableName() string {
	return "word_read"
}
