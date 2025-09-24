package models

import "gorm.io/gorm"

type WordBook struct {
	gorm.Model
	Name     string              `json:"name"`
	UserID   uint                `gorm:"type:bigint" json:"user_id"`
	Category string              `json:"category"`
	Describe string              `json:"describe"`
	Status   int                 `json:"status" gorm:"default:0"`
	Icon     string              `json:"icon"`
	Tag      string              `json:"tag"`
	Words    []WordBooksRelation `gorm:"foreignKey:BookID;references:ID" json:"words"`
}

func (WordBook) TableName() string {
	return "word_books"
}
