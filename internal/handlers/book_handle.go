package handlers

import (
	. "easyjapanese/db"
	"easyjapanese/internal/middleware"
	"easyjapanese/internal/models"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type BookHandler struct{}

func (h *BookHandler) BookRoutes(router *gin.Engine) {
	v1 := router.Group("/book").Use(middleware.User())
	v1.POST("/add", h.addBook)
	v1.GET("/list", h.getWordBookList)
	v1.GET("/word", h.getContainBook)
	v1.POST("/info", h.setBook)
	v1.POST("/release", h.release)
	v1.POST("/delete", h.delBook)
	v2 := router.Group("/word").Use(middleware.User())
	v2.POST("/add", h.addWord)
	v2.GET("/list", h.getWordList)
	v2.POST("/delete", h.delWords)
	v2.POST("/done", h.setDone)
}
func (h *BookHandler) getContainBook(c *gin.Context) {
	wordId := c.Query("wordId")
	UserId, _ := c.Get("UserId")
	var ids []uint
	DB.Debug().Model(&models.WordBooksRelation{}).Where("word_id=? and user_id=?", wordId, UserId).Pluck("book_id", &ids)
	c.JSON(http.StatusOK, gin.H{
		"data": ids,
	})
}
func (h *BookHandler) setDone(c *gin.Context) {
	var Req struct {
		WordIds []uint `json:"word_ids" binding:"required"`
		Type    string `json:"type"`
	}
	UserId, _ := c.Get("UserId")
	if err := c.ShouldBindJSON(&Req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"err": err.Error()})
		return
	}
	DB.Where("user_id = ? AND word_id IN ?", UserId, Req.WordIds).
		Delete(&models.ReviewProgress{})
	if Req.Type == "mark" {
		var words []models.ReviewProgress
		for _, v := range Req.WordIds {
			words = append(words, models.ReviewProgress{
				UserID:         UserId.(uint),
				WordID:         v,
				Quality:        5,
				Done:           true,
				NextReviewDate: time.Date(9999, 12, 31, 23, 59, 59, 0, time.UTC),
			})
		}
		DB.Create(&words)
	}
	c.JSON(http.StatusOK, gin.H{})

}
func (h *BookHandler) release(c *gin.Context) {
	var Req struct {
		ID uint `json:"id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&Req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"err": err.Error()})
		return
	}
	UserId, _ := c.Get("UserId")
	book := models.WordBook{}
	DB.First(&book, Req.ID)
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
	UserId, _ := c.Get("UserId")
	var Req struct {
		ID       uint   `json:"id" binding:"required"`
		Name     string `json:"name" binding:"required"`
		Describe string `json:"describe" binding:"required"`
		Icon     string `json:"icon"`
	}
	if err := c.ShouldBindJSON(&Req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"err": err.Error()})
		return
	}
	book := models.WordBook{}
	DB.First(&book, Req.ID)
	result := DB.Where("name = ? and user_id = ?", Req.Name, UserId.(uint)).First(&models.WordBook{})
	if errors.Is(result.Error, gorm.ErrRecordNotFound) || book.Name == Req.Name {
		if book.UserID == UserId {
			book.ID = Req.ID
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

func handleWordBookRelation(words []models.WordBooksRelation) []JapaneseDictRes {
	result := make([]JapaneseDictRes, 0)
	for _, v := range words {
		result = append(result, JapaneseDictRes{
			ID:          v.Word.ID,
			Words:       v.Word.Words,
			Kana:        v.Word.Kana,
			Rome:        v.Word.Rome,
			Tone:        v.Word.Tone,
			Description: v.Word.Description,
		})
	}
	return result
}
func (h *BookHandler) getWordList(c *gin.Context) {
	id, _ := strconv.Atoi(c.Query("id"))
	tab, _ := strconv.Atoi(c.DefaultQuery("tab", "0"))
	page, _ := strconv.Atoi(c.Query("page"))
	size, _ := strconv.Atoi(c.Query("page_size"))
	if size <= 0 || size > 100 {
		size = 20
	}
	val := c.Query("val")
	UserId, _ := c.Get("UserId")
	offset := (page - 1) * size
	// 搜索到的单词列表
	var wordIDs []uint
	var useWordIDs bool
	if val != "" {
		if err := DB.Model(&models.JapaneseDict{}).
			Select("id").
			Where("MATCH(search_text) AGAINST(? IN NATURAL LANGUAGE MODE)", val).
			Pluck("id", &wordIDs).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "search failed"})
			return
		}
		useWordIDs = true
	}
	var notInIDs, inIDs []uint
	baseQuery := DB.Model(&models.WordBooksRelation{}).Where("book_id = ?", id)
	switch tab {
	case 1: // 未学习
		DB.Model(&models.ReviewProgress{}).
			Where("user_id = ?", UserId).
			Pluck("word_id", &notInIDs)
		baseQuery = baseQuery.Where("word_id NOT IN ?", notInIDs)
	case 2: // 已掌握
		DB.Model(&models.ReviewProgress{}).
			Where("user_id = ? AND done = ?", UserId, true).
			Pluck("word_id", &inIDs)
		baseQuery = baseQuery.Where("word_id IN ?", inIDs)
	case 3: // 学习中
		DB.Model(&models.ReviewProgress{}).
			Where("user_id = ? AND done = ?", UserId, false).
			Pluck("word_id", &inIDs)
		baseQuery = baseQuery.Where("word_id IN ?", inIDs)
	}
	if useWordIDs {
		if len(wordIDs) == 0 {
			c.JSON(http.StatusOK, gin.H{"data": []interface{}{}, "total": 0})
			return
		}
		baseQuery = baseQuery.Where("word_id IN ?", wordIDs)
	}
	// if len(inIDs) > 0 {
	// 	if useWordIDs {
	// 		inIDs = slice.Intersection(inIDs, wordIDs)
	// 		if len(inIDs) == 0 {
	// 			c.JSON(http.StatusOK, gin.H{"data": []interface{}{}, "total": 0})
	// 			return
	// 		}
	// 	}
	// }
	var total int64
	baseQuery.Count(&total)
	var relations []models.WordBooksRelation
	if total > 0 {
		DB.Preload("Word").
			Where("book_id = ?", id).
			Scopes(func(db *gorm.DB) *gorm.DB {
				db = db.Where("book_id = ?", id)
				if useWordIDs && len(wordIDs) > 0 {
					db = db.Where("word_id IN ?", wordIDs)
				}
				if len(notInIDs) > 0 {
					db = db.Where("word_id NOT IN ?", notInIDs)
				}
				if len(inIDs) > 0 {
					db = db.Where("word_id IN ?", inIDs)
				}
				return db
			}).
			Limit(size).Offset(offset).
			Find(&relations)
	}
	result := handleWordBookRelation(relations)
	c.JSON(http.StatusOK, gin.H{
		"data":  result,
		"total": total,
	})
}
func (h *BookHandler) delBook(c *gin.Context) {
	UserId, _ := c.Get("UserId")
	var Req struct {
		ID uint `json:"id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&Req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"err": err.Error()})
		return
	}
	book := models.WordBook{}
	DB.First(&book, Req.ID)
	if book.Status == 1 {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 4001,
			"msg":  "Prohibit deletion",
		})
		return
	}
	DB.Where("user_id = ? and id = ?", UserId, Req.ID).Delete(&models.WordBook{})
	DB.Where("book_id = ?", Req.ID).Delete(&models.WordBooksRelation{})
	c.JSON(http.StatusOK, gin.H{
		"msg": "Deleted successfully",
	})
}

func (h *BookHandler) delWords(c *gin.Context) {
	UserId, _ := c.Get("UserId")
	var Req struct {
		BookId int   `json:"book_id"`
		IDs    []int `json:"ids"`
	}
	if err := c.ShouldBindJSON(&Req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"err": err.Error()})
		return
	}
	DB.Where("user_id = ? and word_id in ? and book_id = ?", UserId, Req.IDs, Req.BookId).Delete(&models.WordBooksRelation{})
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
	err := DB.Where("word_id = ? AND book_id = ? AND user_id = ?", Req.WordID, Req.BookID, UserId).First(&models.WordBooksRelation{}).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		DB.Create(&models.WordBooksRelation{
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
		Name     string `json:"name" binding:"required"`
		Describe string `json:"describe" binding:"required"`
		Icon     string `json:"icon"`
	}
	if err := c.ShouldBindJSON(&Req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"err": err.Error()})
		return
	}
	UserId, _ := c.Get("UserId")
	var count int64
	DB.Model(&models.WordBook{}).Where("user_id=?", UserId).Count(&count)
	if count > 5 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 4002})
		return
	}
	result := DB.Where("name = ? and user_id = ?", Req.Name, UserId.(uint)).First(&models.WordBook{})
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		DB.Create(&models.WordBook{
			Name:     Req.Name,
			Describe: Req.Describe,
			UserID:   UserId.(uint),
			Icon:     Req.Icon,
			Tag:      "自制",
			Category: "自制",
		})
		c.JSON(http.StatusOK, gin.H{})
		return
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"code": 4001})
		return
	}
}

type PublicBookRes struct {
	ID        uint      `json:"id"`
	Name      string    `json:"name"`
	Category  string    `json:"category"`
	Describe  string    `json:"describe"`
	Current   bool      `json:"current"`
	LearnNum  int       `json:"learn_num"`
	WordNum   int       `json:"word_num"`
	Icon      string    `json:"icon"`
	CreatedAt time.Time `json:"created_at"`
	Status    int       `json:"status"`
	UserID    uint      `json:"user_id"`
}

func (h *BookHandler) getWordBookList(c *gin.Context) {
	UserId, _ := c.Get("UserId")
	config := models.UserConfig{}
	DB.Where("user_id=?", UserId).First(&config)
	result := make([]PublicBookRes, 0)
	DB.Model(&models.WordBook{}).Select("word_books.*,COUNT(DISTINCT user_config.user_id) AS learn_num,COUNT(DISTINCT word_books_relation.id) AS word_num").Joins("left join user_config ON word_books.id = user_config.book_id").Joins("left join word_books_relation ON word_books.id = word_books_relation.book_id and word_books_relation.deleted_at IS NULL").Where("word_books.status=1 or word_books.user_id = ?", UserId).Group("word_books.id").Order("word_books.id desc").Find(&result)
	for k, item := range result {
		if config.BookID == item.ID {
			result[k].Current = true
		}
	}
	c.JSON(http.StatusOK, gin.H{
		"data": result,
	})
}
