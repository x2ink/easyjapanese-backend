package models

import (
	"time"
)

type Users struct {
	ID        uint       `json:"id"`
	CreatedAt time.Time  `json:"created_at,omitempty"`
	UpdatedAt time.Time  `json:"updated_at,omitempty"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
	Nickname  string
	Email     string
	Password  string
	Wx        *string
	Qq        *string
	Os        string
	Device    string
	Ip        string
	Avatar    string
	RoleID    uint `gorm:"default:1"`
	Role      Role `gorm:"constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
}
