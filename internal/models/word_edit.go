package models

import (
	"gorm.io/gorm"
)

type WordEdit struct {
	gorm.Model
	Words       []string `json:"words" gorm:"serializer:json"`
	Kana        string   `json:"kana"`
	Tone        string   `json:"tone"`
	Rome        string   `json:"rome"`
	Detail      []Detail `json:"detail" gorm:"serializer:json"`
	Description string   `json:"description"`
	UserID      uint     `gorm:"type:bigint" json:"user_id"`
	WordID      uint     `gorm:"type:int" json:"word_id"`
	Status      uint     `gorm:"column:status;default:0" json:"status"`
	Comment     string   `gorm:"type:text" json:"comment"`
	User        Users    `gorm:"foreignKey:UserID;references:ID" json:"user"`
}

func (WordEdit) TableName() string {
	return "word_edit"
}
