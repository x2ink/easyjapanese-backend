package models

type MybooksWordRelation struct {
	ID     uint   `json:"id"`
	WordId uint   `json:"word_id"`
	BookId uint   `json:"book_id"`
	Word   Jadict `gorm:"foreignKey:WordId" json:"word"`
}

func (MybooksWordRelation) TableName() string {
	return "mybooks_word_relation"
}
