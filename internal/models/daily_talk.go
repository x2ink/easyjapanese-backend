package models

type DailyTalk struct {
	ID    uint   `json:"id"`
	Jp    string `json:"jp"`
	Zh    string `json:"zh"`
	Voice string `json:"voice" gorm:"-"`
	Ruby  []Ruby `json:"ruby" gorm:"serializer:json"`
}

func (DailyTalk) TableName() string {
	return "daily_talk"
}
