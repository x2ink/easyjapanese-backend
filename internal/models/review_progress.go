package models

import (
	"time"

	"gorm.io/gorm"
)

type ReviewProgress struct {
	gorm.Model
	UserID         uint         `gorm:"column:user_id" json:"user_id,omitempty"`
	WordID         uint         `gorm:"column:word_id" json:"word_id,omitempty"`
	Quality        int          `gorm:"column:quality;comment:单词的熟悉度" json:"quality,omitempty"`
	NextReviewDate time.Time    `gorm:"column:next_review_date" json:"next_review_date,omitempty"`
	Easiness       float64      `gorm:"column:easiness" json:"easiness,omitempty"`
	Interval       int          `gorm:"column:interval" json:"interval,omitempty"`
	Repetitions    int          `gorm:"column:repetitions" json:"repetitions,omitempty"`
	Word           JapaneseDict `gorm:"gorm:foreignKey:WordID;references:ID" json:"word"`
	Done           bool         `gorm:"column:done;default:false;comment:是否掌握该单词" json:"done,omitempty"`
}

func (ReviewProgress) TableName() string {
	return "review_progress"
}
