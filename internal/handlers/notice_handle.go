package handlers

import (
	. "easyjapanese/db"
	"easyjapanese/internal/models"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
	"time"
)

type NoticeHandler struct{}

func (h *NoticeHandler) NoticeRoutes(router *gin.Engine) {
	v1 := router.Group("/notice")
	v1.GET("/list/:page/:size", h.getList)
	v1.GET("/info/:id", h.getInfo)
}

type NoticeRes struct {
	CreatedAt time.Time `json:"created_at"`
	ID        uint      `json:"id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	Type      string    `json:"type"`
	Tag       string    `json:"tag"`
	Icon      string    `json:"icon"`
}

func (h *NoticeHandler) getInfo(c *gin.Context) {
	Id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}
	var notice models.Notice
	DB.First(&notice, uint(Id))
	c.JSON(http.StatusOK, gin.H{
		"msg": "Successfully obtained",
		"data": map[string]any{
			"title":      notice.Title,
			"data":       notice.Data,
			"created_at": notice.CreatedAt,
		},
	})
}

func (h *NoticeHandler) getList(c *gin.Context) {
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
	var notices []NoticeRes
	var total int64 = 0
	DB.Order("id desc").Model(&models.Notice{}).Limit(size).Offset(size * (page - 1)).Find(&notices)
	DB.Model(&models.Notice{}).Count(&total)
	c.JSON(http.StatusOK, gin.H{
		"msg":   "Successfully obtained",
		"data":  notices,
		"total": total,
	})
}
