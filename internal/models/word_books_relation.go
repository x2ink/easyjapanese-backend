package models

import "gorm.io/gorm"

type WordBooksRelation struct {
	gorm.Model
	WordID uint         `json:"word_id"`
	BookID uint         `json:"book_id"`
	UserID uint         `json:"user_id"`
	Word   JapaneseDict `gorm:"gorm:foreignKey:WordID;references:ID" json:"word"`
	Book   WordBook     `gorm:"gorm:foreignKey:BookID;references:ID" json:"book"`
}

func (WordBooksRelation) TableName() string {
	return "word_books_relation"
}
