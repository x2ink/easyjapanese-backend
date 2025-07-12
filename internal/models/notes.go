package models

import "gorm.io/gorm"

type Notes struct {
	gorm.Model
	UserID  uint   `gorm:"type:bigint" json:"user_id"`
	User    Users  `gorm:"foreignKey:UserID;references:ID" json:"user"`
	WordID  uint   `json:"word_id"`
	Word    Jcdict `gorm:"foreignKey:WordID;references:ID" json:"word"`
	Content string `gorm:"type:text" json:"content"`
	Public  bool   `gorm:"default:false" json:"public"`
	Like    int    `json:"like" gorm:"default:0"`
	CiteID  *uint  `json:"cite_id" gorm:"default:0"`
	Cite    *Notes `gorm:"foreignKey:CiteID" json:"cite"`
}

func (Notes) TableName() string {
	return "notes"
}
