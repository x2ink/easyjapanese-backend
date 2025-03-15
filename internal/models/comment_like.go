package models

import (
	"gorm.io/gorm"
)

type CommentLike struct {
	gorm.Model
	ID        uint    `json:"id"`
	CommentID uint    `json:"comment_id"`
	UserID    uint    `json:"user_id"`
	User      Users   `gorm:"foreignKey:UserID" json:"user"`
	Comment   Comment `gorm:"foreignKey:CommentID" json:"comment"`
}

func (CommentLike) TableName() string {
	return "comment_like"
}
