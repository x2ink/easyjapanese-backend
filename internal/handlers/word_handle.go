package handlers

import (
	"easyjapanese/db"
	"easyjapanese/internal/models"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"regexp"
	"strconv"
)

type WordHandler struct{}

func (h *WordHandler) WordRoutes(router *gin.Engine) {
	word := router.Group("/ja")
	{
		word.POST("/add", h.AddWord)
		word.GET("/search/:page/:size/:val", h.JaSearch)
		word.GET("/info/:id", h.JaInfo)
	}
}
func (h *WordHandler) AddWord(c *gin.Context) {
	var Word models.Jadict
	{
	}
	if err := c.ShouldBindJSON(&Word); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"err": err.Error()})
		return
	}
	result := db.DB.Create(&Word)
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

func (h *WordHandler) JaInfo(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"err": "The id format is incorrect"})
		return
	}
	var Word models.Jadict
	db.DB.Model(models.Jadict{}).First(&Word, id)
	c.JSON(http.StatusOK, gin.H{
		"msg":  "Successfully obtained",
		"data": Word,
	})
}
func (h *WordHandler) JaSearch(c *gin.Context) {
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
	db.DB.Model(models.Jadict{}).Select("word", "kana", "detail", "id", "deleted_at").Where("word LIKE ? OR kana LIKE ?", searchTerm, searchTerm).Limit(size).Offset(size * (page - 1)).Find(&Word)
	db.DB.Model(models.Jadict{}).Select("word", "kana", "detail", "id", "deleted_at").Where("word LIKE ? OR kana LIKE ?", searchTerm, searchTerm).Count(&total)
	if total > 0 {
		for _, v := range Word {
			Res1.Meaning = GetMeaning(v.Detail)
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
func removeParenthesesContent(input string) string {
	// 正则表达式，匹配所有括号和括号内的内容
	re := regexp.MustCompile(`[（(\[【〔［].*?[］）)\]】〕]`)

	// 替换匹配到的部分为空字符串
	result := re.ReplaceAllString(input, "")

	// 去除多余的空格
	result = regexp.MustCompile(`\s+`).ReplaceAllString(result, " ")

	return result
}
func GetMeaning(detail []models.Detail) []string {
	var res []string
	for _, v := range detail {
		for _, v1 := range v.Detail {
			res = append(res, removeParenthesesContent(v1.Meaning))
		}
	}
	return res
}
