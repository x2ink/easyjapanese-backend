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
	User           Users        `gorm:"gorm:foreignKey:UserID;references:ID" json:"user"`
	Done           bool         `gorm:"column:done;default:false;comment:是否掌握该单词" json:"done,omitempty"`
	Listen         bool         `gorm:"column:listen;default:false;comment:是否掌握该单词" json:"listen,omitempty"`
	Write          bool         `gorm:"column:write;default:false;comment:是否掌握该单词" json:"write,omitempty"`
	Type           string       `gorm:"column:type" json:"type,omitempty"`
}

func (ReviewProgress) TableName() string {
	return "review_progress"
}
