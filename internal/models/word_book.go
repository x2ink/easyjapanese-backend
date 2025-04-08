package models

import "gorm.io/gorm"

type Icon struct {
	Bg   string `json:"bg"`
	Data string `json:"data"`
	Type string `json:"type"`
}
type WordBook struct {
	gorm.Model
	Name     string             `json:"name"`
	UserID   uint               `gorm:"type:bigint" json:"user_id"`
	Category string             `json:"category"`
	Describe string             `json:"describe"`
	Status   int                `json:"status" gorm:"default:0"`
	Icon     Icon               `json:"icon" gorm:"serializer:json"`
	Tag      string             `json:"tag"`
	Words    []WordBookRelation `gorm:"foreignKey:BookID;references:ID" json:"words"`
}

func (WordBook) TableName() string {
	return "word_book"
}
