package models

type Role struct {
	ID   uint `json:"id"`
	Name string
}

func (Role) TableName() string {
	return "role" // 使用你想要的表名
}
