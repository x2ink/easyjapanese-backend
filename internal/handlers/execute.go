package handlers

import (
	"easyjapanese/config"
	. "easyjapanese/db"
	"easyjapanese/internal/middleware"
	"easyjapanese/internal/models"
	"easyjapanese/utils"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/ikawaha/kagome-dict/uni"
	"github.com/ikawaha/kagome/v2/tokenizer"
	"gorm.io/gorm"
)

func Execute(router *gin.Engine) {
	router.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "404 page not found",
		})
	})
	Userhandler := &UserHandler{}
	Userhandler.UserRoutes(router)
	Wordhandler := &WordHandler{}
	Wordhandler.WordRoutes(router)
	router.POST("/like", middleware.User(), like)
	Filehandler := &FileHandler{}
	Filehandler.FileRoutes(router)
	Bookhandler := &BookHandler{}
	Bookhandler.BookRoutes(router)
	Noteshandler := &NotesHandler{}
	Noteshandler.NotesRoutes(router)
	router.GET("/config", middleware.User(), getUserConfig)
	router.POST("/config", middleware.User(), setUserConfig)
	router.POST("/feedback", middleware.User(), feedback)
	router.GET("/verbtrans", verbTrans)
	router.GET("/grammar/list", getGrammarList)
	router.GET("/grammar/info", getGrammarInfo)
	router.GET("/ranking", getRanking)
	router.GET("/dailytalk", getDailyTalk)
	router.GET("/sentence", getSentence)
	router.POST("/tools/break-sentence", sentenceBreakdown)
	router.GET("/culture", getCultureList)
}

type SentenceTokenData struct {
	Surface       string   `json:"surface"`
	BaseForm      string   `json:"base_form"`
	Pos           []string `json:"pos"`
	Pronunciation string   `json:"pronunciation"`
	Features      []string `json:"features"`
}

// 词句拆解
func sentenceBreakdown(c *gin.Context) {
	var Req struct {
		Sentence string `json:"sentence" binding:"required"`
	}
	if err := c.ShouldBindJSON(&Req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"err": err.Error()})
		return
	}
	t, err := tokenizer.New(uni.Dict(), tokenizer.OmitBosEos())
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}
	tokens := t.Tokenize(Req.Sentence)
	result := make([]SentenceTokenData, 0)
	for _, v := range tokens {
		if v.Class == tokenizer.DUMMY {
			continue
		}
		r := tokenizer.NewTokenData(v)
		result = append(result, SentenceTokenData{
			Surface:       r.Surface,
			BaseForm:      r.BaseForm,
			Pos:           r.POS,
			Pronunciation: r.Pronunciation,
			Features:      r.Features,
		})
	}
	c.JSON(http.StatusOK, gin.H{
		"data": result,
	})
}

type userInfo struct {
	Id       uint   `json:"id"`
	Avatar   string `json:"avatar"`
	Nickname string `json:"nickname"`
	Role     string `json:"role"`
}

func (userInfo) TableName() string {
	return "users"
}
func like(c *gin.Context) {
	var Req struct {
		ID   uint   `json:"id"  binding:"required"`
		Type string `json:"type"  binding:"required"`
		Like bool   `json:"like"`
	}
	if err := c.ShouldBindJSON(&Req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"err": err.Error()})
		return
	}
	if Req.Type == "read" {
		var readData models.WordRead
		result := DB.First(&readData, Req.ID).Error
		if errors.Is(result, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"err": "Id not exits",
			})
		} else {
			if Req.Like {
				readData.Like += 1
			} else {
				readData.Like -= 1
			}
			DB.Save(&readData)
			c.JSON(http.StatusOK, gin.H{})
		}
	}
}

func getSentence(c *gin.Context) {
	sentence := models.Sentence{}
	DB.Order("RAND()").First(&sentence)
	c.JSON(http.StatusOK, gin.H{
		"data": sentence,
	})
}

func getDailyTalk(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, err := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	if err != nil || page < 1 {
		page = 1
	}
	var res []models.DailyTalk
	var total int64 = 0
	DB.Limit(pageSize).Offset(pageSize * (page - 1)).Find(&res)
	DB.Model(&models.DailyTalk{}).Count(&total)
	var result []models.DailyTalk
	for _, v := range res {
		result = append(result, models.DailyTalk{
			Voice: fmt.Sprintf("%s/audio/talk/%d.wav", config.AliOssAddress, v.ID),
			ID:    v.ID,
			Jp:    v.Jp,
			Zh:    v.Zh,
			Ruby:  v.Ruby,
		})
	}
	c.JSON(http.StatusOK, gin.H{
		"data":  result,
		"total": total,
	})
}
func getCultureList(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, err := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	if err != nil || page < 1 {
		page = 1
	}
	var res []models.Culture
	var total int64 = 0
	DB.Limit(pageSize).Offset(pageSize * (page - 1)).Find(&res)
	DB.Model(&models.Culture{}).Count(&total)
	c.JSON(http.StatusOK, gin.H{
		"data":  res,
		"total": total,
	})
}

type RankingRes struct {
	User      userInfo `json:"user"`
	WordCount int      `json:"word_count"`
	UserID    uint     `json:"user_id"`
}

func getRanking(c *gin.Context) {
	res := make([]RankingRes, 0)
	DB.Model(&models.ReviewProgress{}).
		Select("user_id, COUNT(DISTINCT word_id) AS word_count").
		Group("user_id").
		Order("word_count DESC").
		Preload("User").Limit(20).
		Find(&res)
	c.JSON(http.StatusOK, gin.H{
		"data": res,
	})
}

func feedback(c *gin.Context) {
	var Req struct {
		Content string `json:"content"  binding:"required"`
		Type    string `json:"type"  binding:"required"`
	}
	if err := c.ShouldBindJSON(&Req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"err": err.Error()})
		return
	}
	UserId, _ := c.Get("UserId")
	feedback := models.Feedback{
		Content: Req.Content,
		Type:    Req.Type,
		UserID:  UserId.(uint),
	}
	DB.Create(&feedback)
	c.JSON(http.StatusOK, gin.H{
		"msg": "Submitted successfully",
	})
}

func getGrammarInfo(c *gin.Context) {
	var GrammarInfo struct {
		ID          uint                    `json:"id"`
		Grammar     string                  `json:"grammar"`
		Level       string                  `json:"level"`
		Connect     string                  `json:"connect"`
		Meanings    []string                `json:"meanings" gorm:"serializer:json"`
		Explanation []string                `json:"explanation" gorm:"serializer:json"`
		Examples    []models.GrammarExample `json:"examples" gorm:"serializer:json"`
	}
	id := c.Query("id")
	DB.Model(&models.Grammar{}).Find(&GrammarInfo, id)
	c.JSON(http.StatusOK, gin.H{
		"data": GrammarInfo,
	})
}

func getGrammarList(c *gin.Context) {
	type GrammarRes struct {
		ID       uint     `json:"id"`
		Grammar  string   `json:"grammar"`
		Level    string   `json:"level"`
		Meanings []string `json:"meanings" gorm:"serializer:json"`
	}
	page, err := strconv.Atoi(c.Query("page"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"err": "The pageSize format is incorrect"})
		return
	}
	pageSize, err := strconv.Atoi(c.Query("pageSize"))
	if err != nil || pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}
	val := c.Query("val")
	level := c.Query("level")
	db := DB.Model(&models.Grammar{})
	var total int64
	result := make([]GrammarRes, 0)
	if val == "" {
		if level != "" {
			db = db.Where("level = ?", level)
		}
		db.Count(&total)
		db.Limit(pageSize).Offset(pageSize * (page - 1)).Find(&result)
	} else {
		db = db.Where("MATCH(grammar) AGAINST (? IN BOOLEAN MODE)", val)
		db.Limit(pageSize).Offset(pageSize * (page - 1)).Find(&result)
		db.Count(&total)
	}
	c.JSON(http.StatusOK, gin.H{
		"data":  result,
		"total": total,
	})
}
func verbTrans(c *gin.Context) {
	word := c.Query("word")
	res := utils.VerbTransfiguration(word)
	if res == nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"msg": "this word is not verb",
		})
	}
	c.JSON(http.StatusOK, gin.H{
		"data": res,
	})
}
func setUserConfig(c *gin.Context) {
	var Req struct {
		ID     uint `json:"id" binding:"required"`
		BookID int  `json:"book_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&Req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"err": err.Error()})
		return
	}
	DB.Model(&models.UserConfig{ID: Req.ID}).Updates(&Req)
	c.Status(http.StatusOK)
}

func getUserConfig(c *gin.Context) {
	UserId, _ := c.Get("UserId")
	var config models.UserConfig
	DB.Preload("Book").First(&config, "user_id = ?", UserId)
	c.JSON(http.StatusOK, gin.H{
		"data": config,
	})
}
