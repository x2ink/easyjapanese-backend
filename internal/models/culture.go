package models

import "gorm.io/gorm"

type Culture struct {
	gorm.Model
	Cover  string   `gorm:"type:varchar(255)" json:"cover"`
	Title  string   `gorm:"type:varchar(255)" json:"title"`
	URL    string   `gorm:"type:varchar(255)" json:"url"`
	Browse int      `gorm:"default:0" json:"browse"`
	Tags   []string `gorm:"serializer:json" json:"tags"`
}

func (Culture) TableName() string {
	return "culture"
}
