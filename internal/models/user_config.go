package models

type UserConfig struct {
	ID            uint   `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID        int64  `gorm:"comment:'用户ID'" json:"user_id"`
	Dailylearning int    `gorm:"comment:'每日学习的数量';default:10" json:"daily_learning"`
	Mode          string `gorm:"comment:'背单词模式';default:'学习模式'" json:"mode"`
	BookID        int    `gorm:"comment:'正在选择的词书'" json:"book_id"`
}

func (UserConfig) TableName() string {
	return "user_config" // 自定义表名
}
