package main

import (
	"easyjapanese/db"
	"easyjapanese/internal/handlers"
	"io"
	"net/http"
	"os"
	"runtime"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	if runtime.GOOS == "linux" {
		f, _ := os.Create("gin.log")
		gin.DefaultWriter = io.MultiWriter(f)
		errorLog, _ := os.Create("error.log")
		gin.DefaultErrorWriter = io.MultiWriter(errorLog, os.Stderr)
	}
	db.InitMysql()
	router := gin.Default()
	router.Static("/file", "./file")
	router.OPTIONS("/*any", func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		c.JSON(http.StatusOK, gin.H{})
	})

	router.Use(cors.New(cors.Config{
		AllowOrigins: []string{"http://192.168.1.2:8080", "http://192.168.1.3:8080", "http://localhost:5173", "http://localhost:3000"},
		MaxAge:       12 * time.Hour,
	}))
	handlers.Execute(router)
	if err := router.Run("0.0.0.0:8080"); err != nil {
		return
	}
}
