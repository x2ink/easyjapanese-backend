package handlers

import (
	. "easyjapanese/db"
	"easyjapanese/internal/middleware"
	"easyjapanese/internal/models"
	"errors"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"net/http"
	"time"
)

type BookHandler struct{}

func (h *BookHandler) BookRoutes(router *gin.Engine) {
	v1 := router.Group("/book").Use(middleware.User())
	v1.POST("", h.addbook)
	v1.GET("/self/:id", h.getSelfBookList)
	//v1.DELETE("/:id", h.delbook)
	//v1.PUT("/:id", h.setbook)
	//v2 := router.Group("/mybooks/word").Use(middleware.User())
	//v2.POST("", h.addword)
	//v2.GET("/:id/:page/:size", h.getWordList)
	//v2.DELETE("/:wordid/:bookid", h.delword)
}

//	func (h *BookHandler) setbook(c *gin.Context) {
//		id, _ := strconv.Atoi(c.Param("id"))
//		UserId, _ := c.Get("UserId")
//		var Req struct {
//			Name     string `json:"name" binding:"required"`
//			Describe string `json:"describe" binding:"required"`
//		}
//		if err := c.ShouldBindJSON(&Req); err != nil {
//			c.JSON(http.StatusBadRequest, gin.H{"err": err.Error()})
//			return
//		}
//		DB.Model(&models.MyBooks{}).Where("user_id = ? and id = ?", UserId, id).Updates(map[string]interface{}{"name": Req.Name, "describe": Req.Describe})
//		c.JSON(http.StatusOK, gin.H{
//			"msg": "Updated successfully",
//		})
//	}
//
//	func (h *BookHandler) getWordList(c *gin.Context) {
//		id, _ := strconv.Atoi(c.Param("id"))
//		result := make([]JcdictRes, 0)
//		page, err := strconv.Atoi(c.Param("page"))
//		if err != nil {
//			c.JSON(http.StatusBadRequest, gin.H{"err": "The page format is incorrect"})
//			return
//		}
//		size, err := strconv.Atoi(c.Param("size"))
//		if err != nil {
//			c.JSON(http.StatusBadRequest, gin.H{"err": "The size format is incorrect"})
//			return
//		}
//		words := make([]models.BookWordRelation, 0)
//		var total int64
//		DB.Preload("Word.Meaning").Where("book_id = ?", id).Limit(size).Offset(size * (page - 1)).Find(&words)
//		DB.Model(models.BookWordRelation{}).Where("book_id = ?", id).Count(&total)
//		for _, v := range words {
//			meanings := make([]string, 0)
//			for _, meaning := range v.Word.Meaning {
//				meanings = append(meanings, meaning.Meaning)
//			}
//			result = append(result, JcdictRes{
//				Word:    v.Word.Word,
//				Kana:    v.Word.Kana,
//				ID:      v.Word.ID,
//				Browse:  v.Word.Browse,
//				Meaning: meanings,
//			})
//		}
//		c.JSON(http.StatusOK, gin.H{
//			"data":  result,
//			"total": total,
//		})
//	}
//
//	func (h *BookHandler) delbook(c *gin.Context) {
//		UserId, _ := c.Get("UserId")
//		id, _ := strconv.Atoi(c.Param("id"))
//		DB.Where("user_id = ? and id = ?", UserId, id).Delete(&models.MyBooks{})
//		DB.Where("book_id = ?", id).Delete(&models.BookWordRelation{})
//		c.JSON(http.StatusOK, gin.H{
//			"msg": "Deleted successfully",
//		})
//	}
//
//	func (h *BookHandler) delword(c *gin.Context) {
//		UserId, _ := c.Get("UserId")
//		wordid, _ := strconv.Atoi(c.Param("wordid"))
//		bookid, _ := strconv.Atoi(c.Param("bookid"))
//		DB.Debug().Where("user_id = ? and word_id = ? and book_id = ?", UserId, wordid, bookid).Delete(&models.BookWordRelation{})
//		c.JSON(http.StatusOK, gin.H{
//			"msg": "Deleted successfully",
//		})
//	}
//
//	func (h *BookHandler) addword(c *gin.Context) {
//		var Req struct {
//			WordID uint `json:"word_id" binding:"required"`
//			BookID uint `json:"book_id" binding:"required"`
//		}
//		UserId, _ := c.Get("UserId")
//		if err := c.ShouldBindJSON(&Req); err != nil {
//			c.JSON(http.StatusBadRequest, gin.H{"err": err.Error()})
//			return
//		}
//		err := DB.Where("word_id = ? AND book_id = ?", Req.WordID, Req.BookID).First(&models.BookWordRelation{}).Error
//		if errors.Is(err, gorm.ErrRecordNotFound) {
//			DB.Create(&models.BookWordRelation{
//				WordID: Req.WordID,
//				BookID: Req.BookID,
//				UserID: UserId.(uint),
//			})
//			c.JSON(http.StatusOK, gin.H{
//				"msg": "Submitted successfully",
//			})
//		} else {
//			c.JSON(http.StatusBadRequest, gin.H{
//				"msg": "The word is already in the vocabulary",
//			})
//		}
//	}
func (h *BookHandler) addbook(c *gin.Context) {
	var Req struct {
		Name     string      `json:"name" binding:"required"`
		Describe string      `json:"describe" binding:"required"`
		Icon     models.Icon `json:"icon" binding:"required"`
	}
	if err := c.ShouldBindJSON(&Req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"err": err.Error()})
		return
	}
	UserId, _ := c.Get("UserId")
	result := DB.Where("name = ? and user_id = ?", Req.Name, UserId.(uint)).First(&models.WordBook{})
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		DB.Create(&models.WordBook{
			Name:     Req.Name,
			Describe: Req.Describe,
			UserID:   UserId.(uint),
			Icon:     Req.Icon,
		})
		c.JSON(http.StatusOK, gin.H{
			"msg": "Submitted successfully",
		})
		return
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"msg": "Word book noun repetition", "code": 4001})
		return
	}
}

type BookRes struct {
	Id        uint        `json:"id"`
	Name      string      `json:"name"`
	Describe  string      `json:"describe"`
	Has       bool        `json:"has" gorm:"-"`
	Icon      models.Icon `json:"icon" gorm:"serializer:json"`
	Word      int         `json:"word" gorm:"-"`
	CreatedAt time.Time   `json:"created_at"`
}

func containsBookId(m []uint, value uint) bool {
	for _, k := range m {
		if k == value {
			return true
		}
	}
	return false
}
func (h *BookHandler) getSelfBookList(c *gin.Context) {
	UserId, _ := c.Get("UserId")
	wordId := c.Param("id")
	books := make([]models.WordBook, 0)
	bookIds := make([]uint, 0)
	result := make([]BookRes, 0)
	DB.Model(&models.WordBookRelation{}).Where("user_id = ? and word_id = ?", UserId, wordId).Pluck("book_id", &bookIds)
	DB.Order("id desc").Preload("Words").Where("user_id = ?", UserId).Find(&books)
	for _, book := range books {
		item := BookRes{}
		if containsBookId(bookIds, book.ID) {
			item.Has = true
		} else {
			item.Has = false
		}
		item.Word = len(book.Words)
		item.Id = book.ID
		item.Name = book.Name
		item.CreatedAt = book.CreatedAt
		item.Describe = book.Describe
		item.Icon = book.Icon
		result = append(result, item)
	}
	c.JSON(http.StatusOK, gin.H{
		"msg":  "Successfully obtained",
		"data": result,
	})
}
