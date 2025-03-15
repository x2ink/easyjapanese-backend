package models

import (
	"gorm.io/gorm"
)

type CommentImage struct {
	gorm.Model
	ID        uint   `json:"id"`
	Url       string `gorm:"type:varchar(255);default:null" json:"url"`
	CommentID uint   `json:"comment_id"`
}

func (CommentImage) TableName() string {
	return "comment_image"
}
