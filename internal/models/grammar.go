package models

type Grammar struct {
	ID      uint   `gorm:"primary_key" json:"id"`
	Grammar string `json:"grammar"`
	Explain string `json:"explain"`
	Level   string `json:"level"`
	Struct  string `json:"struct"`
	Scene   string `json:"scene"`
	Warning string `json:"warning"`
	// Example []Example `json:"example" gorm:"serializer:json"`
	Summary string `json:"summary"`
}

func (Grammar) TableName() string {
	return "grammar"
}
