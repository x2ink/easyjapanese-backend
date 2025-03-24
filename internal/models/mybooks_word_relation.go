package models

type MybooksWordRelation struct {
	ID     uint   `json:"id"`
	WordID uint   `json:"word_id"`
	BookID uint   `json:"book_id"`
	UserID uint   `json:"user_id"`
	Word   Jcdict `gorm:"foreignKey:WordID" json:"word"`
}

func (MybooksWordRelation) TableName() string {
	return "mybooks_word_relation"
}
