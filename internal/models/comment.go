package models

import "gorm.io/gorm"

type Comment struct {
	gorm.Model
	To       uint
	ToUser   Users `gorm:"foreignKey:To" json:"to_user"`
	From     uint
	FromUser Users `gorm:"foreignKey:From" json:"From_user"`
	TrendID  uint
	Level    uint           `gorm:"default:0"`
	Like     []CommentLike  `gorm:"foreignKey:CommentID;references:ID" json:"like"`
	Images   []CommentImage `gorm:"foreignKey:CommentID;references:ID" json:"images"`
	Content  string
	ParentID *uint     `json:"parent_id"`
	Children []Comment `gorm:"foreignKey:ParentID" json:"children"`
	Likenum  int       `gorm:"default:0" json:"likenum"`
}

func (Comment) TableName() string {
	return "comment"
}
