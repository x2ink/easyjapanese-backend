package handlers

import (
	. "easyjapanese/db"
	"easyjapanese/internal/middleware"
	"easyjapanese/internal/models"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type NotesHandler struct{}

func (h *NotesHandler) NotesRoutes(router *gin.Engine) {
	v1 := router.Group("/notes").Use(middleware.User())
	v1.POST("/submit", h.submit)
	v1.POST("/delete", h.del)
	v1.GET("/list", h.getList)
	v1.GET("/info", h.getInfo)
	v1.GET("/origin", h.queryOrigin)
}

type NoteItem struct {
	Content    string    `json:"content"`
	CreatedAt  time.Time `json:"created_at"`
	ID         uint      `json:"id"`
	TargetType string    `json:"target_type"`
}

func (h *NotesHandler) queryOrigin(c *gin.Context) {
	targetType := c.Query("type")
	targetID := c.Query("id")
	var result string
	switch targetType {
	case "word_jp":
		var jpword models.JapaneseDict
		DB.First(&jpword, targetID)
		result = jpword.Kana
		for _, v := range jpword.Words {
			result += fmt.Sprintf("【%s】", v)
		}
	case "talk":
		var talk models.DailyTalk
		DB.First(&talk, targetID)
		result = talk.Jp
	case "culture":
		var culture models.Culture
		DB.First(&culture, targetID)
		result = culture.Title
	}
	c.JSON(http.StatusOK, gin.H{"data": result})
}
func (h *NotesHandler) submit(c *gin.Context) {
	var Req struct {
		TargetType string `json:"target_type"`
		TargetID   uint   `json:"target_id"`
		Content    string `json:"content"`
	}
	if err := c.ShouldBindJSON(&Req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"err": err.Error()})
		return
	}
	UserId, _ := c.Get("UserId")
	var note models.Notes
	err := DB.Where("target_id = ? and target_type = ?", Req.TargetID, Req.TargetType).First(&note).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		note := models.Notes{
			Content:    Req.Content,
			UserID:     UserId.(uint),
			TargetID:   Req.TargetID,
			TargetType: Req.TargetType,
		}
		DB.Model(&models.Notes{}).Create(&note)
		c.JSON(http.StatusOK, gin.H{"data": note.ID})
	} else {
		note.Content = Req.Content
		DB.Save(&note)
		c.JSON(http.StatusOK, gin.H{})
	}
}

func (h *NotesHandler) getList(c *gin.Context) {
	page, err := strconv.Atoi(c.Query("page"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"err": "The pageSize format is incorrect"})
		return
	}
	pageSize, err := strconv.Atoi(c.Query("pageSize"))
	if err != nil || pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}
	UserId, _ := c.Get("UserId")
	var total int64
	result := make([]NoteItem, 0)
	query := DB.Model(&models.Notes{}).Where("user_id= ?", UserId)
	query.Count(&total)
	query.Limit(pageSize).Offset(pageSize * (page - 1)).Find(&result)
	c.JSON(http.StatusOK, gin.H{
		"data":  result,
		"total": total,
	})
}

type NoteInfo struct {
	Content    string    `json:"content"`
	TargetType string    `json:"target_type"`
	ID         uint      `json:"id"`
	TargetID   uint      `json:"target_id"`
	CreatedAt  time.Time `json:"created_at"`
}

func (h *NotesHandler) getInfo(c *gin.Context) {
	id := c.Query("id")
	targetType := c.Query("target_type")
	targetID := c.Query("target_id")
	UserId, _ := c.Get("UserId")
	note := NoteInfo{}
	query := DB.Model(&models.Notes{}).Where("user_id = ?", UserId)
	if id != "" {
		query = query.Where("id = ?", id)
	} else if targetType != "" && targetID != "" {
		query = query.Where("target_type = ? AND target_id = ?", targetType, targetID)
	}
	result := query.First(&note)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"err": "This id is not exist"})
	} else {
		c.JSON(http.StatusOK, gin.H{
			"data": note,
		})
	}
}

func (h *NotesHandler) del(c *gin.Context) {
	UserId, _ := c.Get("UserId")
	var Req struct {
		ID uint `json:"id"`
	}
	if err := c.ShouldBindJSON(&Req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"err": err.Error()})
		return
	}
	DB.Where("id = ? and user_id = ?", Req.ID, UserId).Delete(&models.Notes{})
	c.JSON(http.StatusOK, gin.H{"msg": "Deleted successfully"})
}
