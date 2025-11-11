package models

import "gorm.io/gorm"

type EditMeaning struct {
	Type     string `json:"type"`
	Meanings string `json:"meanings"`
}
type EditExample struct {
	Jp string `json:"jp"`
	Zh string `json:"zh"`
}
type WordEdit struct {
	gorm.Model
	UserID      uint          `json:"user_id"`
	Status      int8          `gorm:"type:tinyint;default:0" json:"status"`
	Comment     string        `gorm:"type:text" json:"comment"`
	Words       string        `gorm:"type:varchar(255)" json:"words"`
	Kana        string        `gorm:"type:varchar(255)" json:"kana"`
	Tone        string        `gorm:"type:varchar(255)" json:"tone"`
	Description string        `gorm:"type:varchar(255)" json:"description"`
	Meanings    []EditMeaning `json:"meanings" gorm:"serializer:json"`
	Examples    []EditExample `json:"examples" gorm:"serializer:json"`
	WordID      uint          `json:"word_id"`
}

func (WordEdit) TableName() string {
	return "word_edit"
}
