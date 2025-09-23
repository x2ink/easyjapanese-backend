package models

type MemorySettings struct {
	Cycle  []int `json:"cycle"`
	Extent struct {
		Forgotten int `json:"forgotten"`
		Vague     int `json:"vague"`
		Partial   int `json:"partial"`
	} `json:"extent"`
}
type UserConfig struct {
	ID          uint           `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID      uint           `gorm:"comment:'用户ID'" json:"user_id"`
	LearnGroup  int            `gorm:"comment:'每组的数量';default:10" json:"learn_group"`
	ReviewGroup int            `gorm:"comment:'每组的数量';default:10" json:"review_group"`
	WriteGroup  int            `gorm:"comment:'每组的数量';default:10" json:"write_group"`
	SoundGroup  int            `gorm:"comment:'每组的数量';default:10" json:"sound_group"`
	Book        WordBook       `gorm:"foreignKey:BookID;" json:"book"`
	BookID      uint           `gorm:"comment:'正在选择的词书';default:1" json:"book_id"`
	CycleConfig MemorySettings `json:"cycle_config" gorm:"serializer:json"`
}

func (UserConfig) TableName() string {
	return "user_config"
}
