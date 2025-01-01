package models

import "time"

type Read struct {
	Top  string `json:"top"`
	Body string `json:"body"`
}

type Example struct {
	Ch    string `json:"ch"`
	Ja    string `json:"ja"`
	Read  []Read `json:"read"`
	Voice string `json:"voice"`
}

type DetailItem struct {
	Example []Example `json:"example"`
	Meaning string    `json:"meaning"`
}

type Detail struct {
	Detail   []DetailItem `json:"detail"`
	Wordtype string       `json:"wordtype"`
}

type Jadict struct {
	ID        uint       `json:"id"`
	CreatedAt time.Time  `json:"created_at,omitempty"`
	UpdatedAt time.Time  `json:"updated_at,omitempty"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
	Word      string     `json:"word"`
	Tone      string     `json:"tone"`
	Rome      string     `json:"rome"`
	Voice     string     `json:"voice"`
	Kana      string     `json:"kana"`
	Detail    []Detail   `json:"detail" gorm:"serializer:json"`
}

func (Jadict) TableName() string {
	return "jadict" // 使用你想要的表名
}
