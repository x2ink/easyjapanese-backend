package main

import (
	"easyjapanese/db"
	"easyjapanese/internal/handlers"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	db.InitMysql()
	db.InitRedis()
	// 配置开发环境和生产环境
	gin.SetMode(gin.ReleaseMode) // 生产环境
	// gin.SetMode(gin.DebugMode) // 开发环境
	fmt.Println(gin.Mode())
	if gin.Mode() == gin.ReleaseMode {
		// 确保启用控制台和文件日志
		f, _ := os.Create("gin.log")
		gin.DefaultWriter = io.MultiWriter(f)
	}
	router := gin.Default()
	router.Static("/file", "./file")
	router.OPTIONS("/*any", func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		c.JSON(http.StatusOK, gin.H{})
	})

	router.Use(cors.New(cors.Config{
		AllowOrigins: []string{"http://192.168.3.77:5173", "http://localhost:5173", "http://172.20.10.10:5173", "http://localhost:3000"},
		MaxAge:       12 * time.Hour,
	}))
	handlers.Execute(router)
	if err := router.Run("0.0.0.0:8080"); err != nil {
		return
	}
}
