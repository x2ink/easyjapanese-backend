package models

import (
	"gorm.io/gorm"
)

type Trend struct {
	gorm.Model
	Content   string       `gorm:"type:longtext;default:null" json:"content" binding:"required"`
	UserID    uint         `gorm:"index" json:"user_id"`
	User      Users        `gorm:"foreignKey:UserID" json:"user"`
	Browse    int          `gorm:"default:0" json:"browse"`
	Likenum   int          `gorm:"default:0" json:"likenum"`
	SectionID int          `gorm:"default:0" json:"section_id"`
	Like      []TrendLike  `gorm:"foreignKey:TrendID;references:ID" json:"like"`
	Images    []TrendImage `gorm:"foreignKey:TrendID;references:ID" json:"images"`
}

func (Trend) TableName() string {
	return "trend"
}
