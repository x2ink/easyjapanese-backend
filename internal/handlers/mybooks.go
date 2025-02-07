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

type MybooksHandler struct{}

func (h *MybooksHandler) MybooksRoutes(router *gin.Engine) {
	v1 := router.Group("/mybooks").Use(middleware.User())
	v1.POST("", h.add)
	v1.GET("", h.getList)
	v1.POST("/add", h.addword)
	v1.POST("/set/:id", h.setbook)
	v1.GET("/list/:id/:page/:size", h.getWordList)
	v1.POST("/del/word/:wordid/:bookid", h.delword)
	v1.POST("/del/book/:id", h.delbook)
}
func (h *MybooksHandler) setbook(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	UserId, _ := c.Get("UserId")
	var Req struct {
		Name     string `json:"name" binding:"required"`
		Describe string `json:"describe" binding:"required"`
	}
	if err := c.ShouldBindJSON(&Req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"err": err.Error()})
		return
	}
	DB.Model(&models.MyBooks{}).Where("user_id = ? and id = ?", UserId, id).Updates(map[string]interface{}{"name": Req.Name, "describe": Req.Describe})
	c.JSON(http.StatusOK, gin.H{
		"msg": "Updated successfully",
	})
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
	DB.Model(models.MybooksWordRelation{}).Where("book_id = ?", id).Count(&total)
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
func (h *MybooksHandler) delbook(c *gin.Context) {
	UserId, _ := c.Get("UserId")
	id, _ := strconv.Atoi(c.Param("id"))
	DB.Where("user_id = ? and id = ?", UserId, id).Delete(&models.MyBooks{})
	DB.Where("book_id = ?", id).Delete(&models.MybooksWordRelation{})
	c.JSON(http.StatusOK, gin.H{
		"msg": "Deleted successfully",
	})
}
func (h *MybooksHandler) delword(c *gin.Context) {
	UserId, _ := c.Get("UserId")
	wordid, _ := strconv.Atoi(c.Param("wordid"))
	bookid, _ := strconv.Atoi(c.Param("bookid"))
	DB.Debug().Where("user_id = ? and word_id = ? and book_id = ?", UserId, wordid, bookid).Delete(&models.MybooksWordRelation{})
	c.JSON(http.StatusOK, gin.H{
		"msg": "Deleted successfully",
	})
}
func (h *MybooksHandler) addword(c *gin.Context) {
	var Req struct {
		WordId uint `json:"word_id" binding:"required"`
		BookId uint `json:"book_id" binding:"required"`
	}
	UserId, _ := c.Get("UserId")
	if err := c.ShouldBindJSON(&Req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"err": err.Error()})
		return
	}
	err := DB.Where("word_id = ? AND book_id = ?", Req.WordId, Req.BookId).First(&models.MybooksWordRelation{}).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		DB.Create(&models.MybooksWordRelation{
			WordId: Req.WordId,
			BookId: Req.BookId,
			UserId: UserId.(uint),
		})
		c.JSON(http.StatusOK, gin.H{
			"msg": "Submitted successfully",
		})
	} else {
		c.JSON(http.StatusBadRequest, gin.H{
			"msg": "The word is already in the vocabulary",
		})
	}
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
	DB.Order("id desc").Model(&models.MyBooks{}).Where("user_id = ?", UserId).Find(&books)
	c.JSON(http.StatusOK, gin.H{
		"msg":  "Successfully obtained",
		"data": books,
	})
}
