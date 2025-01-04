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
)

type LikeHandler struct{}

func (h *LikeHandler) LikeRoutes(router *gin.Engine) {
	v1 := router.Group("/like").Use(middleware.User())
	v1.POST("/:target/:id", h.like)
}
func (h *LikeHandler) like(c *gin.Context) {
	target := c.Param("target")
	targetId, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}
	UserId, _ := c.Get("UserId")
	var like models.Like
	result := DB.Where("target_id = ? AND target = ? AND user_id = ?", uint(targetId), target, UserId).First(&like).Error
	if errors.Is(result, gorm.ErrRecordNotFound) {
		DB.Create(&models.Like{
			Target:   target,
			TargetID: uint(targetId),
			UserID:   UserId.(uint),
		})
		c.JSON(http.StatusOK, gin.H{"msg": "like"})
		return
	} else {
		DB.Unscoped().Delete(&like)
		c.JSON(http.StatusOK, gin.H{"msg": "dislike"})
		return
	}
}
