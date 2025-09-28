package models

import (
	"gorm.io/gorm"
)

type GrammarExample struct {
	Zh   string `json:"zh"`
	Ruby []Ruby `json:"ruby"`
}

type Ruby struct {
	Base string `json:"base"`
	Ruby string `json:"ruby"`
}
type Grammar struct {
	gorm.Model
	Grammar     string           `gorm:"type:varchar(255)"`
	Level       string           `gorm:"type:varchar(50)"`
	Connect     string           `gorm:"type:text"`
	Meanings    []string         `gorm:"serializer:json"`
	Explanation []string         `gorm:"serializer:json"`
	Examples    []GrammarExample `gorm:"serializer:json"`
}

func (Grammar) TableName() string {
	return "grammar"
}
