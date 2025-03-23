package models

type WordBookRelation struct {
	ID     uint     `json:"id"`
	WordID uint     `json:"word_id"`
	BookID uint     `json:"book_id"`
	Word   Jcdict   `gorm:"gorm:foreignKey:WordID;references:ID" json:"word"`
	Book   WordBook `gorm:"gorm:foreignKey:BookID;references:ID" json:"book"`
}

func (WordBookRelation) TableName() string {
	return "word_book_relation"
}
