package models

import "time"

type Users struct {
	ID        uint       `json:"id"` // 将 ID 的 JSON 名称设置为小写的 "id"
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
	Nickname  string
	Email     string
	Password  string
	Wx        *string
	Qq        *string
	Os        string
	Device    string
	Ip        string
	RoleId    uint `gorm:"default:1"`
}
