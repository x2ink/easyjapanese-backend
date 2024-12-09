package main

import (
	"easyjapanese/db"
	"easyjapanese/internal/handlers"
	"github.com/gin-gonic/gin"
)

func main() {
	db.Init()
	router := gin.Default()
	handlers.Execute(router)
	if err := router.Run(":8080"); err != nil {
		return
	}
}
