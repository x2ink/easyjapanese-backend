package models

type UserConfig struct {
	ID           uint     `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID       uint     `gorm:"comment:'用户ID'" json:"user_id"`
	LearnGroup   int      `gorm:"comment:'每组的数量';default:10" json:"learn_group"`
	ReviewGroup  int      `gorm:"comment:'每组的数量';default:10" json:"review_group"`
	Mode         string   `gorm:"comment:'背单词模式';default:'学习模式'" json:"mode"`
	Book         WordBook `gorm:"foreignKey:BookID;" json:"book"`
	BookID       uint     `gorm:"comment:'正在选择的词书';default:1" json:"book_id"`
	ListenSelect bool     `gorm:"default:false" json:"listen_select"`
	Remind       string   `gorm:"default:'20:30'" json:"remind"`
}

func (UserConfig) TableName() string {
	return "user_config" // 自定义表名
}
