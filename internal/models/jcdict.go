package models

import "gorm.io/gorm"

type Jcdict struct {
	gorm.Model
	Word     string          `json:"word" gorm:"index"`
	Tone     string          `json:"tone"`
	Rome     string          `json:"rome"`
	Browse   int             `json:"browse"`
	Voice    string          `json:"voice"`
	Kana     string          `json:"kana" gorm:"index"`
	WordType string          `json:"wordtype"`
	Meaning  []JcdictMeaning `gorm:"foreignKey:WordID;references:ID" json:"meaning"`
	Example  []JcdictExample `gorm:"foreignKey:WordID;references:ID" json:"example"`
}

func (Jcdict) TableName() string {
	return "jcdict"
}
