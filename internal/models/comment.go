package models

import "gorm.io/gorm"

type Comment struct {
	gorm.Model
	To       uint
	ToUser   Users `gorm:"foreignKey:To" json:"to_user"`
	From     uint
	FromUser Users `gorm:"foreignKey:From" json:"From_user"`
	Target   string
	TargetID int
	Like     []CommentLike  `gorm:"foreignKey:TargetID;references:ID" json:"like"`
	Images   []CommentImage `gorm:"foreignKey:TargetID;references:ID" json:"images"`
	Content  string
	ParentID *int      `json:"parent_id"`
	Children []Comment `gorm:"foreignKey:ParentID" json:"children"`
}

func (Comment) TableName() string {
	return "comment"
}
