package main

import (
	"easyjapanese/db"
	"easyjapanese/internal/handlers"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	db.InitMysql()
	db.InitRedis()
	router := gin.Default()
	err := router.SetTrustedProxies([]string{})
	if err != nil {
		return
	}
	router.Use(cors.Default())
	handlers.Execute(router)
	if err := router.Run(); err != nil {
		return
	}
}
