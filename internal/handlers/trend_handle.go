package handlers

import (
	. "easyjapanese/db"
	"easyjapanese/internal/middleware"
	"easyjapanese/internal/models"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"net/http"
	"strconv"
	"time"
)

type TrendHandler struct{}
type userInfo struct {
	Id       uint   `json:"id"`
	Avatar   string `json:"avatar"`
	Nickname string `json:"nickname"`
	Role     string `json:"role"`
}

func (userInfo) TableName() string {
	return "users"
}

type trendResp struct {
	Content   string    `json:"content"`
	Browse    int       `json:"browse"`
	Like      int       `json:"like"`
	SectionId int       `json:"section_id"`
	Images    []string  `json:"images"`
	CreatedAt time.Time `json:"created_at"`
	User      userInfo  `json:"user"`
	Id        uint      `json:"id"`
}

func (h *TrendHandler) TrendRoutes(router *gin.Engine) {
	v1 := router.Group("/trend").Use(middleware.User())
	v1.POST("", h.addTrend)
	v1.DELETE("/:id", h.deleteTrend)
	v1.GET("/:id", h.getInfo)
	v1.GET("/search/:page/:size/:val", h.search)
	v1.GET("/list/:page/:size/:section", h.getList)
	v1.POST("/like/:id", h.like)
	v1.GET("/like/:id", h.getLike)
}
func (h *TrendHandler) getLike(c *gin.Context) {
	targetId, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}
	UserId, _ := c.Get("UserId")
	result := DB.Where("target_id = ? AND user_id = ?", uint(targetId), UserId).First(&models.TrendLike{}).Error
	if errors.Is(result, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusOK, gin.H{"msg": "Successfully obtained", "data": false})
		return
	} else {
		c.JSON(http.StatusOK, gin.H{"msg": "Successfully obtained", "data": true})
		return
	}
}
func (h *TrendHandler) like(c *gin.Context) {
	targetId, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}
	UserId, _ := c.Get("UserId")
	var like models.TrendLike
	result := DB.Where("target_id = ? AND user_id = ?", uint(targetId), UserId).First(&like).Error
	if errors.Is(result, gorm.ErrRecordNotFound) {
		DB.Create(&models.TrendLike{
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

func trendList(c *gin.Context, trends []models.Trend, total int64) {
	searchRes := []trendResp{}
	for _, trend := range trends {
		images := []string{}
		for _, image := range trend.Images {
			images = append(images, image.Url)
		}
		trendRes := trendResp{
			Id:        trend.ID,
			Images:    images,
			Content:   trend.Content,
			Browse:    trend.Browse,
			Like:      len(trend.Like),
			CreatedAt: trend.CreatedAt,
			SectionId: trend.SectionID,
			User: userInfo{
				Id:       trend.UserID,
				Avatar:   trend.User.Avatar,
				Nickname: trend.User.Nickname,
				Role:     trend.User.Role.Name,
			},
		}
		searchRes = append(searchRes, trendRes)
	}
	c.JSON(http.StatusOK, gin.H{
		"msg":   "Successfully obtained",
		"data":  searchRes,
		"total": total,
	})
}
func (h *TrendHandler) getList(c *gin.Context) {
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
	section, err := strconv.Atoi(c.Param("section"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"err": "The section format is incorrect"})
		return
	}
	var total int64
	trends := []models.Trend{}
	if section == 0 {
		DB.Order("id desc").Preload("User.Role").Preload("Images").Preload("Like").Limit(size).Offset(size * (page - 1)).Find(&trends)
		DB.Model(&models.Trend{}).Count(&total)
	} else {
		DB.Order("id desc").Preload("User.Role").Preload("Images").Preload("Like").Where("section_id = ?", section).Limit(size).Offset(size * (page - 1)).Find(&trends)
		DB.Model(&models.Trend{}).Where("section_id = ?", section).Count(&total)
	}
	trendList(c, trends, total)
}
func (h *TrendHandler) search(c *gin.Context) {
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
	val := c.Param("val")
	searchTerm := fmt.Sprintf("%%%s%%", val)
	var total int64
	trends := []models.Trend{}
	DB.Preload("User.Role").Preload("Images").Preload("Like").Where("content LIKE ?", searchTerm).Limit(size).Offset(size * (page - 1)).Find(&trends)
	DB.Model(&models.Trend{}).Where("title LIKE ? OR content LIKE ?", searchTerm, searchTerm).Count(&total)
	trendList(c, trends, total)
}
func (h *TrendHandler) addTrend(c *gin.Context) {
	var Req struct {
		Content   string   `json:"content" binding:"required"`
		SectionId int      `json:"section_id" binding:"required"`
		Images    []string `json:"images" binding:"required"`
	}
	UserId, _ := c.Get("UserId")
	if err := c.ShouldBindJSON(&Req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"err": err.Error()})
		return
	}
	trend := models.Trend{
		Content:   Req.Content,
		UserID:    UserId.(uint),
		SectionID: Req.SectionId,
	}
	DB.Create(&trend)
	for _, v := range Req.Images {
		DB.Create(&models.TrendImage{TargetID: trend.ID, Url: v})
	}
	c.JSON(http.StatusOK, gin.H{"msg": "Created successfully", "data": trend.ID})
}
func (h *TrendHandler) deleteTrend(c *gin.Context) {
	trendId := c.Param("id")
	DB.Delete(&models.Trend{}, trendId)
	DB.Where("target = ? AND target_id = ?", "trend", trendId).Delete(&models.TrendImage{})
	c.JSON(http.StatusOK, gin.H{"msg": "Deteled successfully"})
}

func (h *TrendHandler) getInfo(c *gin.Context) {
	trendId := c.Param("id")
	parsedId, err := strconv.ParseUint(trendId, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}
	var Trend models.Trend
	result := DB.Preload("User.Role").Preload("Images").Preload("Like").First(&Trend, uint(parsedId)).Error
	if errors.Is(result, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusNotFound, gin.H{
			"msg": "Trend does not exist",
		})
		return
	}
	Trend.Browse += 1
	DB.Save(&Trend)
	images := []string{}
	for _, image := range Trend.Images {
		images = append(images, image.Url)
	}
	trendResp := trendResp{
		CreatedAt: Trend.CreatedAt,
		Content:   Trend.Content,
		Browse:    Trend.Browse,
		Like:      len(Trend.Like),
		SectionId: Trend.SectionID,
		Images:    images,
		User: userInfo{
			Id:       Trend.UserID,
			Avatar:   Trend.User.Avatar,
			Nickname: Trend.User.Nickname,
			Role:     Trend.User.Role.Name,
		},
	}
	c.JSON(http.StatusOK, gin.H{"msg": "Successfully obtained", "data": trendResp})
}
