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
	"regexp"
	"strconv"
)

type WordHandler struct{}
type wordBookRes struct {
	Name     string `json:"name"`
	Category string `json:"category"`
	Icon     string `json:"icon"`
	Words    int    `json:"words"`
	ID       uint   `json:"id"`
	Describe string `json:"describe"`
}

func (h *WordHandler) WordRoutes(router *gin.Engine) {
	router.GET("/wordbook", getWordBook)
	router.GET("/todayword", middleware.User(), getTodayWord)
	router.GET("/wordbook/:id/:page/:size", getWordBookList)
	jc := router.Group("/jc")
	{
		jc.POST("/add", h.jcAddWord)
		jc.GET("/search/:page/:size/:val", h.jcSearch)
		jc.GET("/info/:id", h.jcInfo)
	}
	cj := router.Group("/cj")
	{
		cj.POST("/add", h.jcAddWord)
		cj.GET("/search/:page/:size/:val", h.cjSearch)
		cj.GET("/info/:id", h.cjInfo)
	}
}
func getTodayWord(c *gin.Context) {
	UserId, _ := c.Get("UserId")
	var config models.UserConfig
	DB.First(&config, "user_id = ?", UserId)
	wordbooks := []models.WordBookRelation{}
	result := []Res{}
	DB.Order("id desc").Preload("Word").Where("book_id = ?", config.BookID).Limit(config.Dailylearning).Offset(1).Find(&wordbooks)
	for _, book := range wordbooks {
		result = append(result, Res{
			Word:    book.Word.Word,
			Kana:    book.Word.Kana,
			ID:      book.Word.ID,
			Meaning: getMeaning(book.Word.Detail),
		})
	}
	c.JSON(http.StatusOK, gin.H{
		"data": result,
		"msg":  "Successfully obtained",
	})
}
func getWordBookList(c *gin.Context) {
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
	bookId, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}
	wordbooks := []models.WordBookRelation{}
	result := []Res{}
	var total int64
	DB.Order("id desc").Preload("Word").Where("book_id = ?", bookId).Limit(size).Offset(size * (page - 1)).Find(&wordbooks)
	DB.Model(models.WordBookRelation{}).Where("book_id = ?", bookId).Count(&total)
	for _, book := range wordbooks {
		result = append(result, Res{
			Word:    book.Word.Word,
			Kana:    book.Word.Kana,
			ID:      book.Word.ID,
			Meaning: getMeaning(book.Word.Detail),
		})
	}
	c.JSON(http.StatusOK, gin.H{
		"data":  result,
		"total": total,
		"msg":   "Successfully obtained",
	})
}
func getWordBook(c *gin.Context) {
	res := []wordBookRes{}
	wordbook := []models.Wordbook{}
	DB.Model(&models.Wordbook{}).Preload("Words").Find(&wordbook)
	for _, word := range wordbook {
		res = append(res, wordBookRes{
			ID:       word.ID,
			Name:     word.Name,
			Icon:     word.Icon,
			Category: word.Category,
			Words:    len(word.Words),
			Describe: word.Describe,
		})
	}
	c.JSON(http.StatusOK, gin.H{
		"data": res,
		"msg":  "Successfully obtained",
	})
}

func (h *WordHandler) jcAddWord(c *gin.Context) {
	var Word models.Jadict
	if err := c.ShouldBindJSON(&Word); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"err": err.Error()})
		return
	}
	result := DB.Create(&Word)
	if result.Error != nil {
		c.JSON(http.StatusBadRequest, gin.H{"err": result.Error.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"msg": "Successfully Added",
	})
}

type Res struct {
	Word    string   `json:"word"`
	Kana    string   `json:"kana"`
	ID      uint     `json:"id"`
	Meaning []string `json:"meaning"`
}
type ChRes struct {
	Ch     string `json:"ch"`
	Pinyin string `json:"pinyin"`
	Result []Res  `json:"result"`
}

func (h *WordHandler) cjInfo(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"err": "The id format is incorrect"})
		return
	}
	var res ChRes
	var Word models.Chdict
	err = DB.Select("ch", "ja", "pinyin").Model(models.Chdict{}).First(&Word, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"err": "The word does not exist"})
		return
	}
	res.Ch = Word.Ch
	res.Pinyin = Word.Pinyin
	var JaWords []models.Jadict
	if len(Word.Ja) > 0 {
		DB.Select("detail", "id", "kana", "word").Model(&models.Jadict{}).Where("word IN ?", Word.Ja).Find(&JaWords)
	}
	var Res1 Res
	for _, v1 := range JaWords {
		Res1.Meaning = getMeaning(v1.Detail)
		Res1.ID = v1.ID
		Res1.Kana = v1.Kana
		Res1.Word = v1.Word
		res.Result = append(res.Result, Res1)
	}
	c.JSON(http.StatusOK, gin.H{
		"msg":  "Successfully obtained",
		"data": res,
	})
}
func (h *WordHandler) jcInfo(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"err": "The id format is incorrect"})
		return
	}
	var Word models.Jadict
	DB.Model(models.Jadict{}).Omit("created_at", "updated_at").First(&Word, id)
	c.JSON(http.StatusOK, gin.H{
		"msg":  "Successfully obtained",
		"data": Word,
	})
}
func (h *WordHandler) jcSearch(c *gin.Context) {
	var Res1 Res
	var Res2 []Res
	var Word []models.Jadict
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
	DB.Model(models.Jadict{}).Select("word", "kana", "detail", "id", "deleted_at").Where("word LIKE ? OR kana LIKE ?", searchTerm, searchTerm).Order("LENGTH(word) ASC").Limit(size).Offset(size * (page - 1)).Find(&Word)
	DB.Model(models.Jadict{}).Select("word", "kana", "detail", "id", "deleted_at").Where("word LIKE ? OR kana LIKE ?", searchTerm, searchTerm).Count(&total)
	if total > 0 {
		for _, v := range Word {
			Res1.Meaning = getMeaning(v.Detail)
			Res1.ID = v.ID
			Res1.Kana = v.Kana
			Res1.Word = v.Word
			Res2 = append(Res2, Res1)
		}
	}
	c.JSON(http.StatusOK, gin.H{
		"msg":   "Successfully obtained",
		"data":  Res2,
		"total": total,
	})
}
func (h *WordHandler) cjSearch(c *gin.Context) {
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
	var Word []models.Chdict
	val := c.Param("val")
	searchTerm := fmt.Sprintf("%%%s%%", val)
	var total int64
	DB.Model(models.Chdict{}).Select("ch", "id", "pinyin", "deleted_at").Where("ch LIKE ?", searchTerm).Limit(size).Offset(size * (page - 1)).Find(&Word)
	DB.Model(models.Chdict{}).Select("ch", "id", "pinyin", "deleted_at").Where("ch LIKE ?", searchTerm).Count(&total)
	c.JSON(http.StatusOK, gin.H{
		"msg":   "Successfully obtained",
		"data":  Word,
		"total": total,
	})
}
func removeParenthesesContent(input string) string {
	// 正则表达式，匹配所有括号和括号内的内容
	re := regexp.MustCompile(`[（(\[【〔［].*?[］）)\]】〕]`)

	// 替换匹配到的部分为空字符串
	result := re.ReplaceAllString(input, "")

	// 去除多余的空格
	result = regexp.MustCompile(`\s+`).ReplaceAllString(result, " ")

	return result
}
func getMeaning(detail []models.Detail) []string {
	var res []string
	for _, v := range detail {
		for _, v1 := range v.Detail {
			res = append(res, removeParenthesesContent(v1.Meaning))
		}
	}
	return res
}
