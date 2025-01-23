package handlers

import (
	. "easyjapanese/db"
	"easyjapanese/internal/models"
	"errors"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"net/http"
	"strconv"
	"time"
)

type ArticleHandler struct{}

func (h *ArticleHandler) ArticleRoutes(router *gin.Engine) {
	v1 := router.Group("/article")
	v1.GET("/info/:id", h.getInfo)
	v1.GET("/list/:page/:size", h.getList)
}

type ArticleListRes struct {
	UserID      uint      `json:"user_id"`
	User        userInfo  `gorm:"foreignKey:UserID" json:"user"`
	ID          uint      `json:"id"`
	CreatedAt   time.Time `json:"created_at"`
	Title       string    `json:"title"`
	Browse      int       `json:"browse"`
	Encapsulate string    `json:"encapsulate"`
	Icon        string    `json:"icon"`
}
type ArticleRes struct {
	UserID      uint                    `json:"user_id"`
	User        userInfo                `gorm:"foreignKey:UserID" json:"user"`
	ID          uint                    `json:"id"`
	CreatedAt   time.Time               `json:"created_at"`
	Grammer     string                  `json:"grammer"`
	Content     []models.ArticleContent `json:"content" gorm:"serializer:json"`
	Title       string                  `json:"title"`
	Browse      int                     `json:"browse"`
	Audio       string                  `json:"audio"`
	Words       []models.ArticleWords   `json:"words" gorm:"serializer:json"`
	Encapsulate string                  `json:"encapsulate"`
	Icon        string                  `json:"icon"`
}

func (h *ArticleHandler) getList(c *gin.Context) {
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
	Result := []ArticleListRes{}
	DB.Preload("User").Model(&models.Article{}).Order("id desc").Limit(size).Offset(size * (page - 1)).Find(&Result)
	var total int64
	DB.Model(&models.Article{}).Count(&total)
	c.JSON(http.StatusOK, gin.H{
		"data":  Result,
		"total": total,
	})
}
func (h *ArticleHandler) getInfo(c *gin.Context) {
	id := c.Param("id")
	var article ArticleRes
	result := DB.Debug().Preload("User").Model(&models.Article{}).First(&article, id).Error
	if errors.Is(result, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusNotFound, gin.H{
			"msg": "Id does not exist",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"data": article,
	})
}
