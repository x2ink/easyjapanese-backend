package models

import "gorm.io/gorm"

type WordBookRelation struct {
	gorm.Model
	WordID uint `json:"word_id"`
	BookID uint `json:"book_id"`
	UserID uint `json:"user_id"`
	// Word   Jcdict   `gorm:"gorm:foreignKey:WordID;references:ID" json:"word"`
	Book WordBook `gorm:"gorm:foreignKey:BookID;references:ID" json:"book"`
}

func (WordBookRelation) TableName() string {
	return "word_book_relation"
}
