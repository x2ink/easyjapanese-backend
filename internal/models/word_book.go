package models

type WordBook struct {
	ID       uint               `json:"id"`
	Name     string             `json:"name"`
	Category string             `json:"category"`
	Describe string             `json:"describe"`
	Icon     string             `json:"icon"`
	Tag      string             `json:"tag"`
	Words    []WordBookRelation `gorm:"foreignKey:BookID;references:ID" json:"words"`
}

func (WordBook) TableName() string {
	return "word_book"
}
