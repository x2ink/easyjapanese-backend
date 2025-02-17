package models

import (
	"gorm.io/gorm"
)

type Chdict struct {
	gorm.Model
	ID     uint     `json:"id"`
	Ch     string   `json:"ch"`
	Pinyin string   `json:"pinyin"`
	Browse int      `json:"browse"`
	Ja     []string `json:"ja" gorm:"serializer:json"`
}

func (Chdict) TableName() string {
	return "chdict" // 使用你想要的表名
}
