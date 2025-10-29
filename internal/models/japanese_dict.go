package models

import (
	"gorm.io/gorm"
)

type DictExample struct {
	Jp string `json:"jp"`
	Zh string `json:"zh"`
}
type Meaning struct {
	Zh       string        `json:"zh"`
	Jp       string        `json:"jp"`
	Examples []DictExample `json:"examples"`
}
type Detail struct {
	Type     string    `json:"type"`
	Meanings []Meaning `json:"meanings"`
}
type JapaneseDict struct {
	gorm.Model
	Words       []string `json:"words" gorm:"serializer:json"`
	Kana        string   `json:"kana"`
	Tone        string   `json:"tone"`
	Browse      uint     `json:"browse"`
	SearchText  string   `json:"search_text"`
	Rome        string   `json:"rome"`
	Detail      []Detail `json:"detail" gorm:"serializer:json"`
	Meanings    string   `json:"-" gorm:"column:meanings_text"`
	Description string   `json:"description"`
}

func (JapaneseDict) TableName() string {
	return "japanese_dict"
}
