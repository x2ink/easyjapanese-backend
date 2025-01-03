package models

type Section struct {
	ID     uint   `json:"id"`
	Name   string `gorm:"type:varchar(20);default:null" json:"name"`
	Target string `gorm:"type:varchar(25);default:null" json:"target"`
}

func (Section) TableName() string {
	return "section"
}
