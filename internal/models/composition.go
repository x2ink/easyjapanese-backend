package models

type Composition struct {
	ID        int    `gorm:"primaryKey;autoIncrement" json:"id"`
	Title     string `gorm:"type:text" json:"title"`
	Topic     string `gorm:"type:text" json:"topic"`
	Demand    string `gorm:"type:text" json:"demand"`
	Article   string `gorm:"type:text" json:"article"`
	Translate string `gorm:"type:text" json:"translate"`
	Tag       string `gorm:"type:varchar(10)" json:"tag"`
}

func (Composition) TableName() string {
	return "composition"
}
