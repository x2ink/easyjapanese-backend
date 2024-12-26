package models

import (
	"gorm.io/gorm"
)

type Users struct {
	gorm.Model
	Nickname string
	Email    string
	Password string
	Wx       *string
	Qq       *string
	Os       string
	Device   string
	Ip       string
	RoleId   uint `gorm:"default:1"`
}
