package models

import (
	"gorm.io/gorm"
	"time"
)

type Users struct {
	ID        uint `gorm:"primaryKey"`
	Nickname  string
	Email     string
	Password  string
	Wx        *string
	Qq        *string
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
	Os        string
	Device    string
	Ip        string
}
