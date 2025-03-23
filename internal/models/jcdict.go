package models

import "gorm.io/gorm"

type Jcdict struct {
	gorm.Model
	Word     string             `json:"word" gorm:"index"`
	Tone     string             `json:"tone"`
	Rome     string             `json:"rome"`
	Voice    string             `json:"voice"`
	Kana     string             `json:"kana" gorm:"index"`
	Wordtype string             `gorm:"column:'wordtype'" json:"wordtype"`
	Browse   int                `json:"browse"`
	Detail   string             `json:"detail"`
	Meaning  []JcdictMeaning    `gorm:"foreignKey:WordID;references:ID" json:"meaning"`
	Example  []JcdictExample    `gorm:"foreignKey:WordID;references:ID" json:"example"`
	Book     []WordBookRelation `gorm:"foreignKey:WordID;references:ID" json:"book"`
}

func (Jcdict) TableName() string {
	return "jcdict"
}
