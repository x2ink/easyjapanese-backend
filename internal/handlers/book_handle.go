package handlers

import (
	. "easyjapanese/db"
	"easyjapanese/internal/middleware"
	"easyjapanese/internal/models"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/duke-git/lancet/v2/slice"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type BookHandler struct{}

func (h *BookHandler) BookRoutes(router *gin.Engine) {
	v1 := router.Group("/book").Use(middleware.User())
	v1.POST("", h.addBook)
	//查看自己的单词本
	v1.GET("/self/:id", h.getSelfBookList)
	v1.GET("/list", h.getWordBookList)
	//编辑单词本
	v1.PUT("/:id", h.setBook)
	//发布单词
	v1.POST("/release/:id", h.release)
	//删除单词本
	v1.DELETE("/:id", h.delBook)
	//加入单词本
	v2 := router.Group("/book/word").Use(middleware.User())
	v2.POST("", h.addWord)
	v2.GET("/:id/:tab/:page/:size/:val", h.getWordList)
	v2.GET("/:id/:tab/:page/:size", h.getWordList)
	v2.DELETE("/:bookid", h.delWords)
	v2.DELETE("/:bookid/:wordid", h.delWord)
}
func (h *BookHandler) release(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	UserId, _ := c.Get("UserId")
	book := models.WordBook{}
	DB.First(&book, id)
	if book.Status == 2 {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 4001,
			"msg":  "Please do not resubmit",
		})
		return
	}
	book.Status = 2
	DB.Save(&book)
	if book.UserID == UserId {
		c.JSON(http.StatusOK, gin.H{
			"msg": "Submited",
		})
	} else {
		c.JSON(http.StatusForbidden, gin.H{
			"msg": "Unknown error",
		})
	}
}
func (h *BookHandler) setBook(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	UserId, _ := c.Get("UserId")
	var Req struct {
		Name     string      `json:"name" binding:"required"`
		Describe string      `json:"describe" binding:"required"`
		Icon     models.Icon `json:"icon" binding:"required"`
	}
	if err := c.ShouldBindJSON(&Req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"err": err.Error()})
		return
	}
	book := models.WordBook{}
	DB.First(&book, id)
	result := DB.Where("name = ? and user_id = ?", Req.Name, UserId.(uint)).First(&models.WordBook{})
	if errors.Is(result.Error, gorm.ErrRecordNotFound) || book.Name == Req.Name {
		if book.UserID == UserId {
			book.ID = uint(id)
			book.Name = Req.Name
			book.Describe = Req.Describe
			book.Icon = Req.Icon
			DB.Updates(book)
			c.JSON(http.StatusOK, gin.H{
				"msg": "Updated successfully",
			})
		} else {
			c.JSON(http.StatusForbidden, gin.H{
				"msg": "Unknown error",
			})
		}
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"msg": "Word book noun repetition", "code": 4001})
		return
	}
}
func handleWordBookRelation(words []models.WordBookRelation) []JcdictRes {
	result := make([]JcdictRes, 0)
	for _, v := range words {
		meanings := make([]string, 0)
		for _, meaning := range v.Word.Meaning {
			meanings = append(meanings, meaning.Meaning)
		}
		result = append(result, JcdictRes{
			Word:    v.Word.Word,
			Kana:    v.Word.Kana,
			ID:      v.Word.ID,
			Browse:  v.Word.Browse,
			Meaning: meanings,
		})
	}
	return result
}
func (h *BookHandler) getWordList(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	tab, _ := strconv.Atoi(c.Param("tab"))
	page, _ := strconv.Atoi(c.Param("page"))
	size, _ := strconv.Atoi(c.Param("size"))
	val := c.Param("val")
	UserId, _ := c.Get("UserId")
	result := make([]JcdictRes, 0)
	words := make([]models.WordBookRelation, 0)
	userConfig := &models.UserConfig{}
	DB.Model(&models.UserConfig{}).Where("user_id = ?", UserId).First(&userConfig)
	tempIDs := make([]uint, 0)
	var total int64 = 0
	ids := make([]uint, 0)
	if val == "" {
		//不是搜索
		if tab == 0 {
			DB.Preload("Word.Meaning").Where("book_id = ?", id).Limit(size).Offset(size * (page - 1)).Find(&words)
			DB.Model(models.WordBookRelation{}).Where("book_id = ?", id).Count(&total)
		} else if tab == 1 {
			DB.Model(&models.LearnRecord{}).Select("word_id").Where("user_id = ?", UserId).Find(&ids)
			DB.Preload("Word.Meaning").Where("word_book_relation.book_id = ? and word_id not in ?", id, ids).Limit(size).Offset(size * (page - 1)).Find(&words)
			DB.Model(models.WordBookRelation{}).Where("book_id = ? and word_id not in ?", id, ids).Count(&total)
		} else if tab == 3 {
			DB.Model(&models.LearnRecord{}).Select("word_id").Where("user_id = ? and done is false", UserId).Find(&ids)
			DB.Preload("Word.Meaning").Where("word_book_relation.book_id = ? and word_id in ?", id, ids).Limit(size).Offset(size * (page - 1)).Find(&words)
			DB.Model(models.WordBookRelation{}).Where("book_id = ? and word_id in ?", id, ids).Count(&total)
		} else if tab == 2 {
			DB.Model(&models.LearnRecord{}).Select("word_id").Where("user_id = ? and done is true", UserId).Find(&ids)
			DB.Preload("Word.Meaning").Where("word_book_relation.book_id = ? and word_id in ?", id, ids).Limit(size).Offset(size * (page - 1)).Find(&words)
			DB.Model(models.WordBookRelation{}).Where("book_id = ? and word_id in ?", id, ids).Count(&total)
		}
	} else {
		//搜索
		searchTerm := fmt.Sprintf("%%%s%%", val)
		DB.Model(models.Jcdict{}).Where("word LIKE ? or kana LIKE ?", searchTerm, searchTerm).Pluck("id", &tempIDs)
		if tab == 0 {
			DB.Preload("Word.Meaning").Where("book_id = ? and word_id in ?", id, tempIDs).Limit(size).Offset(size * (page - 1)).Find(&words)
			DB.Model(models.WordBookRelation{}).Where("book_id = ? and word_id in ?", id, tempIDs).Count(&total)
		} else if tab == 1 {
			DB.Model(&models.LearnRecord{}).Select("word_id").Where("user_id = ?", UserId).Find(&ids)
			DB.Preload("Word.Meaning").Where("book_id = ? and word_id in ? and word_id not in ?", id, tempIDs, ids).Limit(size).Offset(size * (page - 1)).Find(&words)
			DB.Model(models.WordBookRelation{}).Where("book_id = ? and word_id in ? and word_id not in ?", id, tempIDs, ids).Count(&total)
		} else if tab == 3 {
			DB.Model(&models.LearnRecord{}).Select("word_id").Where("user_id = ? and done is false", UserId).Find(&ids)
			intersection := slice.Intersection(ids, tempIDs)
			DB.Preload("Word.Meaning").Where("word_book_relation.book_id = ? and word_id in ?", id, intersection).Limit(size).Offset(size * (page - 1)).Find(&words)
			DB.Model(models.WordBookRelation{}).Where("book_id = ? and word_id in ?", id, ids).Count(&total)
		} else if tab == 2 {
			DB.Model(&models.LearnRecord{}).Select("word_id").Where("user_id = ? and done is true", UserId).Find(&ids)
			intersection := slice.Intersection(ids, tempIDs)
			DB.Preload("Word.Meaning").Where("word_book_relation.book_id = ? and word_id in ?", id, intersection).Limit(size).Offset(size * (page - 1)).Find(&words)
			DB.Model(models.WordBookRelation{}).Where("book_id = ? and word_id in ?", id, ids).Count(&total)

		}
	}
	result = handleWordBookRelation(words)
	c.JSON(http.StatusOK, gin.H{
		"data":  result,
		"total": total,
	})
}
func (h *BookHandler) delBook(c *gin.Context) {
	UserId, _ := c.Get("UserId")
	id, _ := strconv.Atoi(c.Param("id"))
	book := models.WordBook{}
	DB.First(&book, id)
	if book.Status == 1 {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 4001,
			"msg":  "Prohibit deletion",
		})
		return
	}
	DB.Where("user_id = ? and id = ?", UserId, id).Delete(&models.WordBook{})
	DB.Where("book_id = ?", id).Delete(&models.WordBookRelation{})
	c.JSON(http.StatusOK, gin.H{
		"msg": "Deleted successfully",
	})
}
func (h *BookHandler) delWords(c *gin.Context) {
	UserId, _ := c.Get("UserId")
	bookid, _ := strconv.Atoi(c.Param("bookid"))
	var Req struct {
		IDs []int `json:"ids"`
	}
	if err := c.ShouldBindJSON(&Req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"err": err.Error()})
		return
	}
	DB.Where("user_id = ? and word_id in ? and book_id = ?", UserId, Req.IDs, bookid).Delete(&models.WordBookRelation{})
	c.JSON(http.StatusOK, gin.H{
		"msg": "Deleted successfully",
	})
}
func (h *BookHandler) delWord(c *gin.Context) {
	UserId, _ := c.Get("UserId")
	wordid, _ := strconv.Atoi(c.Param("wordid"))
	bookid, _ := strconv.Atoi(c.Param("bookid"))
	DB.Where("user_id = ? and word_id = ? and book_id = ?", UserId, wordid, bookid).Delete(&models.WordBookRelation{})
	c.JSON(http.StatusOK, gin.H{
		"msg": "Deleted successfully",
	})
}
func (h *BookHandler) addWord(c *gin.Context) {
	var Req struct {
		WordID uint `json:"word_id" binding:"required"`
		BookID uint `json:"book_id" binding:"required"`
	}
	UserId, _ := c.Get("UserId")
	if err := c.ShouldBindJSON(&Req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"err": err.Error()})
		return
	}
	err := DB.Where("word_id = ? AND book_id = ? AND user_id = ?", Req.WordID, Req.BookID, UserId).First(&models.WordBookRelation{}).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		DB.Create(&models.WordBookRelation{
			WordID: Req.WordID,
			BookID: Req.BookID,
			UserID: UserId.(uint),
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
func (h *BookHandler) addBook(c *gin.Context) {
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

type SelfBookRes struct {
	ID        uint        `json:"id"`
	Name      string      `json:"name"`
	Describe  string      `json:"describe"`
	Has       bool        `json:"has" gorm:"-"`
	Icon      models.Icon `json:"icon" gorm:"serializer:json"`
	Word      int         `json:"word" gorm:"-"`
	CreatedAt time.Time   `json:"created_at"`
	Status    int         `json:"status"`
}
type PublicBookRes struct {
	ID        uint        `json:"id"`
	Name      string      `json:"name"`
	Category  string      `json:"category"`
	Describe  string      `json:"describe"`
	Current   bool        `json:"current"`
	LearnNum  int         `json:"learn_num"`
	WordNum   int         `json:"word_num"`
	Icon      models.Icon `json:"icon" gorm:"serializer:json"`
	CreatedAt time.Time   `json:"created_at"`
	Status    int         `json:"status"`
}

func (h *BookHandler) getWordBookList(c *gin.Context) {
	UserId, _ := c.Get("UserId")
	config := models.UserConfig{}
	DB.Where("user_id=?", UserId).First(&config)
	result := make([]PublicBookRes, 0)
	DB.Debug().Model(&models.WordBook{}).Select("word_book.*,COUNT(DISTINCT user_config.user_id) AS learn_num,COUNT(DISTINCT word_book_relation.id) AS word_num").Joins("left join user_config ON word_book.id = user_config.book_id").Joins("left join word_book_relation ON word_book.id = word_book_relation.book_id").Where("word_book.status=1").Group("word_book.id").Find(&result)
	for k, item := range result {
		if config.BookID == item.ID {
			result[k].Current = true
		}
	}
	c.JSON(http.StatusOK, gin.H{
		"msg":  "Successfully obtained",
		"data": result,
	})
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
	result := make([]SelfBookRes, 0)
	DB.Model(&models.WordBookRelation{}).Where("user_id = ? and word_id = ?", UserId, wordId).Pluck("book_id", &bookIds)
	DB.Order("id desc").Preload("Words").Where("user_id = ?", UserId).Find(&books)
	for _, book := range books {
		item := SelfBookRes{}
		if containsBookId(bookIds, book.ID) {
			item.Has = true
		} else {
			item.Has = false
		}
		item.Word = len(book.Words)
		item.ID = book.ID
		item.Name = book.Name
		item.CreatedAt = book.CreatedAt
		item.Describe = book.Describe
		item.Icon = book.Icon
		item.Status = book.Status
		result = append(result, item)
	}
	c.JSON(http.StatusOK, gin.H{
		"msg":  "Successfully obtained",
		"data": result,
	})
}
