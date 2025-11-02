package models

type UserConfig struct {
	ID     uint     `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID uint     `gorm:"comment:'用户ID'" json:"user_id"`
	Book   WordBook `gorm:"foreignKey:BookID;" json:"book"`
	BookID uint     `gorm:"comment:'正在选择的词书';default:1" json:"book_id"`
}

func (UserConfig) TableName() string {
	return "user_config"
}
