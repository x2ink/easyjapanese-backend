package handlers

import (
	. "easyjapanese/db"
	"easyjapanese/internal/middleware"
	"easyjapanese/internal/models"
	"easyjapanese/utils"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

func Execute(router *gin.Engine) {
	router.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "404 page not found",
		})
	})
	Userhandler := &UserHandler{}
	Userhandler.UserRoutes(router)
	Captchahandler := &CaptchaHandler{}
	Captchahandler.CaptchaRoutes(router)
	Wordhandler := &WordHandler{}
	Wordhandler.WordRoutes(router)
	Trendhandler := &TrendHandler{}
	Trendhandler.TrendRoutes(router)
	Filehandler := &FileHandler{}
	Filehandler.FileRoutes(router)
	Bookhandler := &BookHandler{}
	Bookhandler.BookRoutes(router)
	Messagehandler := &MessageHandler{}
	Messagehandler.MessageRoutes(router)
	Noteshandler := &NotesHandler{}
	Noteshandler.NotesRoutes(router)
	router.GET("/config", middleware.User(), getUserConfig)
	router.POST("/config", middleware.User(), setUserConfig)
	router.POST("/feedback", middleware.User(), feedback)
	router.GET("/verbtrans/:word", verbTrans)
	router.GET("/grammar/search/:page/:size/:val", searchGrammar)
	router.GET("/grammar/search/:page/:size", searchGrammar)
	router.GET("/grammar/list/:level/:page/:size", getGrammarList)
	router.GET("/grammar/:id", getGrammarInfo)
	router.GET("/unread", middleware.User(), getUnread)
	router.GET("/ranking", middleware.User(), getRanking)
	router.GET("/dailytalk/:page/:size", getDailyTalk)
	//	随机获取谚语
	router.GET("/sentence", getSentence)
}
func getSentence(c *gin.Context) {
	sentence := models.Sentence{}
	DB.Order("RAND()").First(&sentence)
	c.JSON(http.StatusOK, gin.H{
		"data": sentence,
	})
}
func getDailyTalk(c *gin.Context) {
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
	var res []models.DailyTalk
	var total int64 = 0
	DB.Limit(size).Offset(size * (page - 1)).Find(&res)
	DB.Model(&models.DailyTalk{}).Count(&total)
	c.JSON(http.StatusOK, gin.H{
		"msg":   "Successfully obtained",
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
	DB.Preload("User").Select("COUNT(id) as word_count", "user_id").Model(&models.LearnRecord{}).Group("user_id").Order("word_count desc").Find(&res)
	c.JSON(http.StatusOK, gin.H{
		"data": res,
	})
}
func getUnread(c *gin.Context) {
	UserId, _ := c.Get("UserId")
	var likeTotal int64 = 0
	var commentTotal int64 = 0
	var noticeTotal int64 = 0
	DB.Model(&models.Message{}).Where("to_id = ? and status = 0 and type = 'like'", UserId.(uint)).Count(&likeTotal)
	DB.Model(&models.Message{}).Where("to_id = ? and status = 0 and type = 'msg'", UserId.(uint)).Count(&commentTotal)
	DB.Model(&models.Message{}).Where("to_id = ? and status = 0 and type = 'notice'", UserId.(uint)).Count(&noticeTotal)
	c.JSON(http.StatusOK, gin.H{
		"like_total":   likeTotal,
		"msg_total":    commentTotal,
		"notice_total": noticeTotal,
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

type GrammarRes struct {
	Grammar string           `json:"grammar"`
	Id      uint             `json:"id"`
	Level   string           `json:"level"`
	Explain string           `json:"explain"`
	Example []models.Example `json:"example" gorm:"serializer:json"`
}

func getGrammarInfo(c *gin.Context) {
	id := c.Param("id")
	result := models.Grammar{}
	DB.Find(&result, id)
	c.JSON(http.StatusOK, gin.H{
		"data": result,
	})
}
func searchGrammar(c *gin.Context) {
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
	var total int64
	result := make([]GrammarRes, 0)
	searchTerm := fmt.Sprintf("%%%s%%", val)
	DB.Model(&models.Grammar{}).Where("grammar like ?", searchTerm).Limit(size).Offset(size * (page - 1)).Find(&result)
	DB.Model(&models.Grammar{}).Where("grammar like ?", searchTerm).Count(&total)
	c.JSON(http.StatusOK, gin.H{
		"data":  result,
		"total": total,
	})
}
func getGrammarList(c *gin.Context) {
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
	level := c.Param("level")
	var total int64
	result := make([]GrammarRes, 0)
	DB.Model(&models.Grammar{}).Where("level = ?", level).Limit(size).Offset(size * (page - 1)).Find(&result)
	DB.Model(&models.Grammar{}).Where("level = ?", level).Count(&total)
	c.JSON(http.StatusOK, gin.H{
		"data":  result,
		"total": total,
	})
}
func verbTrans(c *gin.Context) {
	word := c.Param("word")
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
		ID           uint   `json:"id" binding:"required"`
		LearnGroup   int    `json:"learn_group" binding:"required"`
		ReviewGroup  int    `json:"review_group" binding:"required"`
		Mode         string `json:"mode" binding:"required"`
		BookID       int    `json:"book_id" binding:"required"`
		Remind       string `json:"remind" binding:"required"`
		ListenSelect *bool  `json:"listen_select" binding:"required"`
	}
	if err := c.ShouldBindJSON(&Req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"err": err.Error()})
		return
	}
	DB.Model(&models.UserConfig{ID: Req.ID}).Updates(&Req)
	c.JSON(http.StatusOK, gin.H{
		"msg": "update success",
	})
}

func getUserConfig(c *gin.Context) {
	UserId, _ := c.Get("UserId")
	var config models.UserConfig
	DB.Preload("Book").First(&config, "user_id = ?", UserId)
	c.JSON(http.StatusOK, gin.H{
		"data": config,
		"msg":  "Successfully obtained",
	})
}
