package models

import "time"

type LikeRecord struct {
	ID        uint      `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	TargetID  uint      `gorm:"null" json:"target_id"`
	Target    string    `gorm:"size:100;null" json:"target"`
	FromID    uint      `gorm:"null" json:"from_id"`
	ToID      uint      `gorm:"null" json:"to_id"`
	Status    int       `gorm:"default:0" json:"status"`
	Content   string    `gorm:"type:text" json:"content"`
	FromUser  Users     `gorm:"foreignKey:FromID" json:"from_user"`
	ToUser    Users     `gorm:"foreignKey:ToID" json:"to_user"`
}

func (LikeRecord) TableName() string {
	return "like_record"
}
