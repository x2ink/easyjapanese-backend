package handlers

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

type UserHandler struct{}

func (h *UserHandler) RegisterRoutes(router *gin.Engine) {
	router.POST("/register", h.Register)
	//rg := router.Group("/user")
	//rg.POST("/register")
}
func (h *UserHandler) Register(c *gin.Context) {
	var req struct {
		Nickname string `json:"nickname" binding:"required,min=2,max=7"`
		Password string `json:"password" binding:"required"`
		Email    string `json:"email" binding:"required,email"`
		Os       string `json:"os" binding:"required"`
		Device   string `json:"device" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	//user := models.Users{Nickname: "小王", Email: "97400220@qq.com", Password: "1234", Os: "web", Device: "iphone", Ip: "127.0.0.1"}
	//result := db.DB.Create(&user)
	//log.Println(result.Error)
	c.JSON(http.StatusOK, gin.H{
		"msg": "pong",
	})
}
