package models

import (
	"gorm.io/gorm"
)

type Users struct {
	gorm.Model
	ID       uint   `json:"id"`
	Nickname string `gorm:"default:'默认昵称'"`
	Email    string
	Password string
	Wx       string
	Qq       string
	Os       string
	Device   string
	Ip       string
	Avatar   string
	RoleID   uint `gorm:"default:1"`
	Role     Role `gorm:"constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
}
