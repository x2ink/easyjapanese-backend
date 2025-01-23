package models

import (
	"gorm.io/gorm"
)

type ArticleContent struct {
	Explain string `json:"explain"`
	Ja      string `json:"ja"`
	Ch      string `json:"ch"`
	Read    []Read `json:"read"`
	Url     string `json:"url"`
	Type    string `json:"type"`
}
type ArticleWords struct {
	Word    string `json:"word"`
	Kana    string `json:"kana"`
	Meaning string `json:"meaning"`
}
type Article struct {
	gorm.Model
	UserID      uint             `json:"user_id"`
	User        Users            `gorm:"foreignKey:UserID" json:"user"`
	Grammer     string           `json:"grammer"`
	Content     []ArticleContent `json:"content" gorm:"serializer:json"`
	Title       string           `json:"title"`
	Browse      int              `json:"browse"`
	Audio       string           `json:"audio"`
	Words       []ArticleWords   `json:"words" gorm:"serializer:json"`
	Encapsulate string           `json:"encapsulate"`
	Icon        string           `json:"icon"`
}

func (Article) TableName() string {
	return "article"
}
