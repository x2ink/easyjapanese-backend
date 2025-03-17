package handlers

import (
	. "easyjapanese/db"
	"easyjapanese/internal/middleware"
	"easyjapanese/internal/models"
	"errors"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"net/http"
	"strconv"
	"time"
)

type MessageHandler struct{}

func (h *MessageHandler) MessageRoutes(router *gin.Engine) {
	v1 := router.Group("/message").Use(middleware.User())
	v1.GET("/:form/:page/:size", h.getList)
	v1.GET("/read/:id", h.readMessage)
	v1.GET("/info/:id", h.getMessage)
}

type Message struct {
	ToID    uint   `json:"to_id"`
	FromID  uint   `json:"from_id"`
	Content string `json:"content"`
	Title   string `json:"title"`
	Path    string `json:"path"`
	Form    string `json:"form"`
	Cover   string `json:"cover"`
	Tag     string `json:"tag"`
}

func SendMessage(msg Message) uint {
	Msg := models.Message{
		Status:  0,
		ToID:    msg.ToID,
		FromID:  msg.FromID,
		Title:   msg.Title,
		Content: msg.Content,
		Path:    msg.Path,
		Type:    msg.Form,
		Cover:   msg.Cover,
		Tag:     msg.Tag,
	}
	DB.Create(&Msg)
	return Msg.ID
}
func (h *MessageHandler) getMessage(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"err": "The id format is incorrect"})
		return
	}
	var msg MessageRes
	result := DB.Model(&models.Message{}).Where("id = ?", id).First(&msg).Error
	if errors.Is(result, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusBadRequest, gin.H{"err": "The id not exist"})
		return
	} else {
		c.JSON(http.StatusOK, gin.H{"msg": "Success", "data": msg})
	}
}

// 设置消息已读

func (h *MessageHandler) readMessage(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	UserId, _ := c.Get("UserId")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"err": "The id format is incorrect"})
		return
	}
	var message models.Message
	result := DB.Where("to_id = ?", UserId.(uint)).First(&message, id).Error
	if errors.Is(result, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusBadRequest, gin.H{"err": "The id not exist"})
		return
	} else {
		message.Status = 1
		DB.Save(&message)
		c.JSON(http.StatusOK, gin.H{"msg": "Success"})
	}
}

type MessageRes struct {
	ID        uint      `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	FromID    uint      `json:"from_id"`
	ToID      uint      `json:"to_id"`
	Status    int       `json:"status"`
	FromUser  userInfo  `json:"from_user"`
	ToUser    userInfo  `json:"to_user"`
	Content   string    `json:"content"`
	Title     string    `json:"title"`
	Path      string    `json:"path"`
	Cover     string    `json:"cover"`
	Tag       string    `json:"tag"`
}

func (h *MessageHandler) getList(c *gin.Context) {
	form := c.Param("form")
	UserId, _ := c.Get("UserId")
	page, err := strconv.Atoi(c.Param("page"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"err": "The page format is incorrect"})
		return
	}
	size, err := strconv.Atoi(c.Param("size"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"err": "The size format is incorrect"})
		return
	}
	var messages []models.Message
	var total int64 = 0
	DB.Preload("FromUser.Role").Preload("ToUser.Role").Order("id desc").Where("to_id = ? and type=?", UserId.(uint), form).Model(&models.Message{}).Limit(size).Offset(size * (page - 1)).Find(&messages)
	DB.Where("to_id = ? and type=?", UserId.(uint), form).Model(&models.Message{}).Count(&total)
	res := make([]MessageRes, 0)
	for _, item := range messages {
		messageRes := MessageRes{
			Tag:       item.Tag,
			Cover:     item.Cover,
			ID:        item.ID,
			CreatedAt: item.CreatedAt,
			FromID:    item.FromID,
			ToID:      item.ToID,
			Status:    item.Status,
			Title:     item.Title,
			Path:      item.Path,
			Content:   item.Content,
			FromUser: userInfo{
				Id:       item.FromUser.ID,
				Avatar:   item.FromUser.Avatar,
				Nickname: item.FromUser.Nickname,
				Role:     item.FromUser.Role.Name,
			},
			ToUser: userInfo{
				Id:       item.ToUser.ID,
				Avatar:   item.ToUser.Avatar,
				Nickname: item.ToUser.Nickname,
				Role:     item.ToUser.Role.Name,
			},
		}
		res = append(res, messageRes)
	}
	c.JSON(http.StatusOK, gin.H{
		"msg":   "Successfully obtained",
		"data":  res,
		"total": total,
	})
}
