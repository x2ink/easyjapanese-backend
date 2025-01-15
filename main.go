package main

import (
	"easyjapanese/db"
	"easyjapanese/internal/handlers"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
)

func main() {
	db.InitMysql()
	db.InitRedis()
	router := gin.Default()
	router.OPTIONS("/*any", func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		c.JSON(http.StatusOK, gin.H{})
	})

	router.Use(cors.New(cors.Config{
		AllowOrigins: []string{"http://192.168.0.101:5173", "http://localhost:5173"},
		MaxAge:       12 * time.Hour,
	}))
	handlers.Execute(router)
	if err := router.Run(); err != nil {
		return
	}
}
