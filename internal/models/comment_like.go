package models

import (
	"gorm.io/gorm"
)

type CommentLike struct {
	gorm.Model
	ID       uint    `json:"id"`
	TargetID uint    `json:"target_id"`
	UserID   uint    `json:"user_id"`
	User     Users   `gorm:"foreignKey:UserID" json:"user"`
	Comment  Comment `gorm:"foreignKey:TargetID" json:"comment"`
}

func (CommentLike) TableName() string {
	return "comment_like"
}
