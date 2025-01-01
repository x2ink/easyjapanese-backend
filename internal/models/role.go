package models

import "time"

type Role struct {
	ID        uint       `json:"id"`
	CreatedAt time.Time  `json:"created_at,omitempty"`
	UpdatedAt time.Time  `json:"updated_at,omitempty"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
	Name      string
}

func (Role) TableName() string {
	return "role" // 使用你想要的表名
}
