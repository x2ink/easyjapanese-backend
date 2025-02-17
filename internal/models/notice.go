package models

import "gorm.io/gorm"

type Notice struct {
	gorm.Model
	Tag     string
	Title   string
	Content string
	Type    string
	Data    string
	Icon    string
}

func (Notice) TableName() string {
	return "notice" // 使用你想要的表名
}
