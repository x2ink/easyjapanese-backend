package models

type DailyTalk struct {
	ID   uint   `json:"id"`
	Ja   string `json:"ja"`
	Ch   string `json:"ch"`
	Read []Read `json:"read" gorm:"serializer:json"`
}

func (DailyTalk) TableName() string {
	return "daily_talk"
}
