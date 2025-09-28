package models

import (
	"strings"

	"gorm.io/gorm"
)

type DictExample struct {
	Jp string `json:"jp"`
	Zh string `json:"zh"`
}
type Meaning struct {
	Zh      string        `json:"zh"`
	Jp      string        `json:"jp"`
	Anti    string        `json:"anti"`
	Note    string        `json:"note"`
	Syns    string        `json:"syns"`
	Example []DictExample `json:"example"`
}
type Detail struct {
	Type     string    `json:"type"`
	Meanings []Meaning `json:"meanings"`
}
type JapaneseDict struct {
	gorm.Model
	Words      []string `json:"words" gorm:"serializer:json"`
	Kana       string   `json:"kana"`
	Tone       string   `json:"tone"`
	Browse     uint     `json:"browse"`
	SearchText string   `json:"search_text"`
	Rome       string   `json:"rome"`
	Detail     []Detail `json:"icon" gorm:"serializer:json"`
	Meanings   string   `json:"-" gorm:"column:meanings_text"`
}

func (JapaneseDict) TableName() string {
	return "japanese_dict"
}
func (j *JapaneseDict) AfterFind(tx *gorm.DB) error {
	var meanings []string
	for _, d := range j.Detail {
		for _, m := range d.Meanings {
			if m.Zh != "" {
				meanings = append(meanings, m.Zh)
			}
		}
	}
	j.Meanings = strings.Join(meanings, "\n")
	return nil
}
