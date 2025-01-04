package handlers

import (
	. "easyjapanese/db"
	"easyjapanese/internal/middleware"
	"easyjapanese/internal/models"
	"easyjapanese/utils"
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
	Address  string `json:"address"`
	Role     string `json:"role"`
}
type trendResp struct {
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	Browse    int       `json:"browse"`
	Like      int       `json:"like"`
	SectionId int       `json:"section_id"`
	Images    []string  `json:"images"`
	CreatedAt time.Time `json:"created_at"`
	User      userInfo  `json:"user"`
}

func (h *TrendHandler) TrendRoutes(router *gin.Engine) {
	v1 := router.Group("/trend").Use(middleware.User())
	v1.POST("/", h.addTrend)
	v1.DELETE("/:id", h.deleteTrend)
	v1.GET("/:id", h.getInfo)
	v1.GET("/search/:page/:size/:val", h.search)
	v1.GET("/list/:page/:size/:section", h.getList)
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
	var trends []models.Trend
	DB.Preload("User").Where("section_id = ?", section).Limit(size).Offset(size * (page - 1)).Find(&trends)
	DB.Model(&models.Trend{}).Where("section_id = ?", section).Count(&total)
	var images []models.Image
	DB.Where("target = ?", "trend").Find(&images)
	var likes []models.Like
	DB.Where("target = ?", "trend").Find(&likes)
	var searchRes []trendResp
	for _, trend := range trends {
		address, _ := utils.GetIpAddress(trend.User.Ip)
		image := []string{}
		likeCount := 0
		for _, v := range images {
			if trend.ID == v.TargetID {
				image = append(image, v.Url)
			}
		}
		for _, v := range likes {
			if trend.ID == v.TargetID {
				likeCount++
			}
		}
		trendRes := trendResp{
			Images:    image,
			Title:     trend.Title,
			Content:   trend.Content,
			Browse:    trend.Browse,
			Like:      likeCount,
			CreatedAt: trend.CreatedAt,
			SectionId: trend.SectionID,
			User: userInfo{
				Id:       trend.UserID,
				Avatar:   trend.User.Avatar,
				Nickname: trend.User.Nickname,
				Address:  address,
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
	var trends []models.Trend
	DB.Preload("User").Where("title LIKE ? OR content LIKE ?", searchTerm, searchTerm).Limit(size).Offset(size * (page - 1)).Find(&trends)
	DB.Model(&models.Trend{}).Where("title LIKE ? OR content LIKE ?", searchTerm, searchTerm).Count(&total)
	var images []models.Image
	DB.Where("target = ?", "trend").Find(&images)
	var likes []models.Like
	DB.Where("target = ?", "trend").Find(&likes)
	var searchRes []trendResp
	for _, trend := range trends {
		address, _ := utils.GetIpAddress(trend.User.Ip)
		image := []string{}
		likeCount := 0
		for _, v := range images {
			if trend.ID == v.TargetID {
				image = append(image, v.Url)
			}
		}
		for _, v := range likes {
			if trend.ID == v.TargetID {
				likeCount++
			}
		}
		trendRes := trendResp{
			Images:    image,
			Title:     trend.Title,
			Content:   trend.Content,
			Browse:    trend.Browse,
			Like:      likeCount,
			CreatedAt: trend.CreatedAt,
			SectionId: trend.SectionID,
			User: userInfo{
				Id:       trend.UserID,
				Avatar:   trend.User.Avatar,
				Nickname: trend.User.Nickname,
				Address:  address,
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
func (h *TrendHandler) addTrend(c *gin.Context) {
	var Req struct {
		Content   string   `json:"content" binding:"required"`
		Title     string   `json:"title" binding:"required"`
		SectionId int      `json:"section_id" binding:"required"`
		Images    []string `json:"images" binding:"required"`
	}
	UserId, _ := c.Get("UserId")
	if err := c.ShouldBindJSON(&Req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"err": err.Error()})
		return
	}
	trend := models.Trend{
		Title:     Req.Title,
		Content:   Req.Content,
		UserID:    UserId.(uint),
		SectionID: Req.SectionId,
	}
	DB.Create(&trend)
	for _, v := range Req.Images {
		DB.Create(&models.Image{Target: "trend", TargetID: trend.ID, Url: v})
	}
	c.JSON(http.StatusOK, gin.H{"msg": "Created successfully", "data": trend.ID})
}
func (h *TrendHandler) deleteTrend(c *gin.Context) {
	trendId := c.Param("id")
	DB.Delete(&models.Trend{}, trendId)
	DB.Where("target = ? AND target_id = ?", "trend", trendId).Delete(&models.Image{})
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
	result := DB.Preload("User.Role").First(&Trend, uint(parsedId)).Error
	if errors.Is(result, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusNotFound, gin.H{
			"msg": "Trend does not exist",
		})
		return
	}
	var images []string
	var image []models.Image
	DB.Where("target = ? AND target_id = ?", "trend", Trend.ID).Find(&image)
	for _, v := range image {
		images = append(images, v.Url)
	}
	address, err := utils.GetIpAddress(Trend.User.Ip)
	Trend.Browse += 1
	DB.Save(&Trend)
	var likeCount int64
	DB.Model(&models.Like{}).Where("target = ? AND target_id = ?", "trend", Trend.ID).Count(&likeCount)
	trendResp := trendResp{
		CreatedAt: Trend.CreatedAt,
		Title:     Trend.Title,
		Content:   Trend.Content,
		Browse:    Trend.Browse,
		Like:      int(likeCount),
		SectionId: Trend.SectionID,
		Images:    images,
		User: userInfo{
			Id:       Trend.UserID,
			Avatar:   Trend.User.Avatar,
			Nickname: Trend.User.Nickname,
			Address:  address,
			Role:     Trend.User.Role.Name,
		},
	}
	c.JSON(http.StatusOK, gin.H{"msg": "Successfully obtained", "data": trendResp})
}
