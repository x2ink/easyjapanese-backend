package models

type Sentence struct {
	ID     int    `gorm:"primaryKey;autoIncrement" json:"id"`
	Ja     string `gorm:"type:text" json:"ja"`
	Ch     string `gorm:"type:text" json:"ch"`
	Source string `gorm:"type:text" json:"source"`
}

func (Sentence) TableName() string {
	return "sentence"
}
