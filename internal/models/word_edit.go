package models

import "gorm.io/gorm"

type WordEdit struct {
	gorm.Model
	WordID  uint   `gorm:"column:word_id"`
	Detail  string `gorm:"type:text"`
	Meaning string `gorm:"type:text"`
	Example string `gorm:"type:text"`
	Comment string `gorm:"type:text" json:"comment"`
	UserID  uint   `gorm:"column:user_id"`
	Status  int    `gorm:"default:0"`
	User    Users  `gorm:"foreignKey:UserID" json:"user"`
}

func (WordEdit) TableName() string {
	return "word_edit"
}
