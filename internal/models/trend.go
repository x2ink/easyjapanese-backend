package models

import (
	"gorm.io/gorm"
)

type Trend struct {
	gorm.Model
	ID        uint   `json:"id"`
	Title     string `gorm:"type:varchar(255);default:null" json:"title" binding:"required"`
	Content   string `gorm:"type:longtext;default:null" json:"content" binding:"required"`
	UserID    uint   `gorm:"index" json:"user_id"`
	User      Users  `gorm:"foreignKey:UserID" json:"user"`
	Browse    int    `gorm:"default:0" json:"browse"`
	SectionID int    `gorm:"index" json:"section_id" binding:"required"`
}

func (Trend) TableName() string {
	return "trend"
}
