package handlers

import (
	. "easyjapanese/db"
	"easyjapanese/internal/middleware"
	"easyjapanese/internal/models"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

type MybooksHandler struct{}

func (h *MybooksHandler) MybooksRoutes(router *gin.Engine) {
	v1 := router.Group("/mybooks").Use(middleware.User())
	v1.POST("", h.add)
	v1.GET("", h.getList)
	v1.POST("/add", h.addword)
	v1.GET("/list/:id/:page/:size", h.getWordList)
	v1.POST("/del/:id", h.delword)
}
func (h *MybooksHandler) getWordList(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var Res1 Res
	var Res2 []Res
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
	words := []models.MybooksWordRelation{}
	var total int64
	DB.Preload("Word").Where("book_id = ?", id).Limit(size).Offset(size * (page - 1)).Find(&words)
	DB.Model(models.WordBookRelation{}).Where("book_id = ?", id).Count(&total)
	if total > 0 {
		for _, v := range words {
			Res1.Meaning = getMeaning(v.Word.Detail)
			Res1.ID = v.Word.ID
			Res1.Kana = v.Word.Kana
			Res1.Word = v.Word.Word
			Res2 = append(Res2, Res1)
		}
	}
	c.JSON(http.StatusOK, gin.H{
		"data":  Res2,
		"total": total,
	})
}
func (h *MybooksHandler) delword(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	DB.Delete(&models.MybooksWordRelation{}, id)
	c.JSON(http.StatusOK, gin.H{
		"msg": "Deleted successfully",
	})
}
func (h *MybooksHandler) addword(c *gin.Context) {
	var Req struct {
		WordId uint `json:"word_id" binding:"required"`
		BookId uint `json:"book_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&Req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"err": err.Error()})
		return
	}
	DB.Create(&models.MybooksWordRelation{
		WordId: Req.WordId,
		BookId: Req.BookId,
	})
	c.JSON(http.StatusOK, gin.H{
		"msg": "Submitted successfully",
	})
}
func (h *MybooksHandler) add(c *gin.Context) {
	var Req struct {
		Name     string `json:"name" binding:"required"`
		Describe string `json:"describe" binding:"required"`
	}
	if err := c.ShouldBindJSON(&Req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"err": err.Error()})
		return
	}
	UserId, _ := c.Get("UserId")
	var total int64
	DB.Model(&models.MyBooks{}).Where("user_id = ?", UserId).Count(&total)
	if total >= 5 {
		c.JSON(http.StatusBadRequest, gin.H{
			"msg": "The vocabulary book has a maximum of five words",
		})
	} else {
		mybooks := models.MyBooks{
			Name:     Req.Name,
			Describe: Req.Describe,
			UserID:   UserId.(uint),
		}
		DB.Create(&mybooks)
		c.JSON(http.StatusOK, gin.H{
			"msg": "Submitted successfully",
		})
	}
}

type MybooksRes struct {
	Id       uint   `json:"id"`
	Name     string `json:"name"`
	Describe string `json:"describe"`
}

func (h *MybooksHandler) getList(c *gin.Context) {
	UserId, _ := c.Get("UserId")
	var books []MybooksRes
	DB.Model(&models.MyBooks{}).Where("user_id = ?", UserId).Find(&books)
	c.JSON(http.StatusOK, gin.H{
		"msg":  "Successfully obtained",
		"data": books,
	})
}
