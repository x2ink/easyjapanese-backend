package models

type Read struct {
	Top  string `json:"top"`
	Body string `json:"body"`
}
type JcdictExample struct {
	ID     uint   `json:"id"`
	WordID uint   `json:"word_id" gorm:"index"`
	Ja     string `json:"ja"`
	Ch     string `json:"ch"`
	Read   []Read `json:"read" gorm:"serializer:json"`
	Voice  string `json:"voice"`
}

func (JcdictExample) TableName() string {
	return "jcdict_example"
}
