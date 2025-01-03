package models

import (
	"gorm.io/gorm"
)

type Trend struct {
	gorm.Model
	ID        uint    `json:"id"`
	Title     string  `gorm:"type:varchar(255);default:null" json:"title" binding:"required"`
	Content   string  `gorm:"type:longtext;default:null" json:"content" binding:"required"`
	UserID    uint    `gorm:"index" json:"user_id"`
	Browse    int     `gorm:"default:0" json:"browse"`
	Like      int     `gorm:"default:0" json:"like"`
	SectionID int     `gorm:"index" json:"section_id" binding:"required"`
	Images    []Image `gorm:"foreignKey:TargetID" json:"images"`
}

func (Trend) TableName() string {
	return "trend"
}
