package models

type Wordbook struct {
	ID       uint               `json:"id"`
	Name     string             `json:"name"`
	Category string             `json:"category"`
	Describe string             `json:"describe"`
	Icon     string             `json:"icon"`
	Tag      string             `json:"tag"`
	Words    []WordBookRelation `gorm:"foreignKey:BookId" json:"words"`
}

func (Wordbook) TableName() string {
	return "wordbook"
}
