package handlers

import (
	"easyjapanese/db"
	"easyjapanese/internal/models"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
)

type WordHandler struct{}

func (h *WordHandler) WordRoutes(router *gin.Engine) {
	router.POST("/word", h.AddWord)
}
func (h *WordHandler) AddWord(c *gin.Context) {
	var Word models.Jadict
	{
	}
	if err := c.ShouldBindJSON(&Word); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"err": err.Error()})
		return
	}
	log.Println(Word.Detail[0].Detail[0].Example[0].Read)
	db.DB.Create(&Word)
}
