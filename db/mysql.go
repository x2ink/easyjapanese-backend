package db

import (
	"easyjapanese/config"
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitMysql() {
	dsn := fmt.Sprintf("%s:%s@tcp(%s)/easyjapanese?charset=utf8mb4&parseTime=True&loc=Local", config.DBUsername, config.DBPassword, config.MysqlAddress)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		PrepareStmt: true,
	})
	if err != nil {
		return
	}
	DB = db
}
