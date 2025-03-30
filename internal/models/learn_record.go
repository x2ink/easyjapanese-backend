package models

import (
	"gorm.io/gorm"
)

type LearnRecord struct {
	gorm.Model
	WordID      uint   `gorm:"column:word_id" json:"word_id"`
	UserID      uint   `gorm:"column:user_id" json:"user_id"`
	ReviewTime  int64  `gorm:"column:review_time" json:"review_time"`
	ReviewCount int    `gorm:"column:review_count;default:0" json:"review_count"`
	ErrorCount  int    `gorm:"column:error_count;default:0" json:"error_count"`
	Pattern     int    `gorm:"column:pattern;default:0" json:"pattern"`
	Word        Jcdict `gorm:"foreignKey:WordID" json:"word"`
	User        Users  `gorm:"foreignKey:UserID" json:"user"`
	Done        bool   `gorm:"column:done;default:false" json:"done"`
}

func (LearnRecord) TableName() string {
	return "learn_record"
}
