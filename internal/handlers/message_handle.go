package handlers

import (
	. "easyjapanese/db"
	"easyjapanese/internal/middleware"
	"easyjapanese/internal/models"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
	"time"
)

type MessageHandler struct{}

func (h *MessageHandler) MessageRoutes(router *gin.Engine) {
	v1 := router.Group("/message").Use(middleware.User())
	v1.GET("/:page/:size", h.getList)
}
func sendMessage(toId uint, fromId uint, commentId uint, content string) {
	DB.Create(&models.Message{
		Status:    0,
		ToID:      toId,
		FromID:    fromId,
		CommentId: commentId,
		Content:   content,
	})
}

type MessageRes struct {
	ID        uint        `json:"id"`
	CreatedAt time.Time   `json:"created_at"`
	CommentId uint        `json:"comment_id"`
	Comment   commentInfo `json:"comment"`
	FromID    uint        `json:"from_id"`
	ToID      uint        `json:"to_id"`
	Status    int         `json:"status"`
	FromUser  userInfo    `json:"from_user"`
	ToUser    userInfo    `json:"to_user"`
	Content   string      `json:"content"`
}

type commentInfo struct {
	ID       uint   `json:"id"`
	To       uint   `json:"to_id"`
	From     uint   `json:"from_id"`
	Target   string `json:"target"`
	TargetID int    `json:"target_id"`
	Content  string `json:"content"`
	ParentID *int   `json:"parent_id"`
	Like     bool   `json:"like"`
}

func (h *MessageHandler) getList(c *gin.Context) {
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
	DB.Preload("Comment.Like").Preload("FromUser.Role").Preload("ToUser.Role").Order("id desc").Where("to_id = ?", UserId.(uint)).Model(&models.Message{}).Limit(size).Offset(size * (page - 1)).Find(&messages)
	DB.Where("to_id = ?", UserId.(uint)).Model(&models.Message{}).Count(&total)
	res := make([]MessageRes, 0)
	for _, item := range messages {
		messageRes := MessageRes{
			ID:        item.ID,
			CreatedAt: item.CreatedAt,
			FromID:    item.FromID,
			ToID:      item.ToID,
			Status:    item.Status,
			Content:   item.Content,
			Comment: commentInfo{
				ID:       item.Comment.ID,
				To:       item.Comment.To,
				From:     item.Comment.From,
				Target:   item.Comment.Target,
				TargetID: item.Comment.TargetID,
				Content:  item.Comment.Content,
				ParentID: item.Comment.ParentID,
				Like:     HasLike(item.Comment.Like, UserId.(uint)),
			},
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
