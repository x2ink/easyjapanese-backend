package models

import (
	"time"
)

type Chdict struct {
	ID        uint       `json:"id"`
	CreatedAt time.Time  `json:"created_at,omitempty"`
	UpdatedAt time.Time  `json:"updated_at,omitempty"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
	Ch        string     `json:"ch"`
	Pinyin    string     `json:"pinyin"`
	Ja        []string   `json:"ja" gorm:"serializer:json"`
}

func (Chdict) TableName() string {
	return "chdict" // 使用你想要的表名
}
