package models

type JcdictMeaning struct {
	ID      uint   `json:"id"`
	WordID  uint   `json:"word_id" gorm:"index"`
	Meaning string `json:"meaning"`
}

func (JcdictMeaning) TableName() string {
	return "jcdict_meaning"
}
