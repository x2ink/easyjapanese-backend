package models

import (
	"gorm.io/gorm"
)

type CommentImage struct {
	gorm.Model
	ID       uint   `json:"id"`
	Url      string `gorm:"type:varchar(255);default:null" json:"url"`
	TargetID uint   `json:"target_id"`
}

func (CommentImage) TableName() string {
	return "comment_image"
}
