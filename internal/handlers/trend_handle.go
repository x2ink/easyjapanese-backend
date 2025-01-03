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

type TrendHandler struct{}

func (h *TrendHandler) TrendRoutes(router *gin.Engine) {
	v1 := router.Group("/trend").Use(middleware.User())
	v1.POST("/", h.addTrend)
	v1.DELETE("/:id", h.deleteTrend)
	v1.GET("/:id", h.getInfo)
}
func (h *TrendHandler) addTrend(c *gin.Context) {
	var Trend models.Trend
	UserId, _ := c.Get("UserId")
	Trend.UserID = UserId.(uint)
	if err := c.ShouldBindJSON(&Trend); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"err": err.Error()})
		return
	}
	DB.Create(&Trend)
	for _, v := range Trend.Images {
		image := models.Image{Target: "trend", TargetID: Trend.ID, Url: v.Url}
		DB.Create(&image)
	}
	c.JSON(http.StatusOK, gin.H{"msg": "Created successfully", "data": Trend.ID})
}
func (h *TrendHandler) deleteTrend(c *gin.Context) {
	trendId := c.Param("id")
	DB.Delete(&models.Trend{}, trendId)
	DB.Where("target = ? AND target_id = ?", "trend", trendId).Delete(&models.Image{})
	c.JSON(http.StatusOK, gin.H{"msg": "Deteled successfully"})
}

type trendResp struct {
	Title    string   `json:"title"`
	Content  string   `json:"content"`
	Browse   int      `json:"browse"`
	Like     int      `json:"like"`
	SelectId int      `json:"select_id"`
	Images   []string `json:"images"`
}

func (h *TrendHandler) getInfo(c *gin.Context) {
	trendId := c.Param("id")
	parsedId, err := strconv.ParseUint(trendId, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}
	var Trend models.Trend
	result := DB.Preload("Images", func(db *gorm.DB) *gorm.DB {
		return db.Where("target = ?", "trend")
	}).First(&Trend, uint(parsedId)).Error
	if errors.Is(result, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusNotFound, gin.H{
			"msg": "Trend does not exist",
		})
		return
	}
	var images []string
	for _, v := range Trend.Images {
		images = append(images, v.Url)
	}
	trendResp := trendResp{
		Title:    Trend.Title,
		Content:  Trend.Content,
		Browse:   Trend.Browse,
		Like:     Trend.Like,
		SelectId: Trend.SectionID,
		Images:   images,
	}
	c.JSON(http.StatusOK, gin.H{"msg": "Successfully obtained", "data": trendResp})
}
