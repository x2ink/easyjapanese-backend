package models

type UserConfig struct {
	ID          uint     `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID      int64    `gorm:"comment:'用户ID'" json:"user_id"`
	LearnGroup  int      `gorm:"comment:'每组的数量';default:10" json:"learn_group"`
	ReviewGroup int      `gorm:"comment:'每组的数量';default:10" json:"review_group"`
	Mode        string   `gorm:"comment:'背单词模式';default:'学习模式'" json:"mode"`
	Book        Wordbook `gorm:"foreignKey:BookID" json:"book"`
	BookID      int      `gorm:"comment:'正在选择的词书'" json:"book_id"`
}

func (UserConfig) TableName() string {
	return "user_config" // 自定义表名
}
