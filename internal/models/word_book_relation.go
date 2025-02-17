package models

type WordBookRelation struct {
	ID     uint     `json:"id"`
	WordId uint     `json:"word_id"`
	BookId uint     `json:"book_id"`
	Word   Jadict   `gorm:"gorm:foreignKey:WordId;references:ID" json:"word"`
	Book   Wordbook `gorm:"gorm:foreignKey:BookId;references:ID" json:"book"`
}

func (WordBookRelation) TableName() string {
	return "word_book_relation"
}
