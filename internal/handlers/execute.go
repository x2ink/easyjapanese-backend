package handlers

import (
	. "easyjapanese/db"
	"easyjapanese/internal/middleware"
	"easyjapanese/internal/models"
	"easyjapanese/utils"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
	"time"
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
	Sectionhandler := &SectionHandler{}
	Sectionhandler.SectionRoutes(router)
	Filehandler := &FileHandler{}
	Filehandler.FileRoutes(router)
	Commenthandler := &CommentHandler{}
	Commenthandler.CommentRoutes(router)
	Articlehandler := &ArticleHandler{}
	Articlehandler.ArticleRoutes(router)
	Mybookshandler := &MybooksHandler{}
	Mybookshandler.MybooksRoutes(router)
	Messagehandler := &MessageHandler{}
	Messagehandler.MessageRoutes(router)
	Noticehandler := &NoticeHandler{}
	Noticehandler.NoticeRoutes(router)
	router.GET("/likerecord/:page/:size", middleware.User(), getLikeRecordList)
	router.GET("/config", middleware.User(), getUserConfig)
	router.POST("/config", middleware.User(), setUserConfig)
	router.POST("/feedback", middleware.User(), feedback)
	router.GET("/verbtrans/:word", verbTrans)
	router.GET("/grammar", getGrammarList)
	router.GET("/grammar/:id", getGrammarInfo)
	router.GET("/unread", middleware.User(), getUnread)
}
func getUnread(c *gin.Context) {
	UserId, _ := c.Get("UserId")
	var likeTotal int64 = 0
	var commentTotal int64 = 0
	DB.Model(&models.LikeRecord{}).Where("to_id = ? and status = 0", UserId.(uint)).Count(&likeTotal)
	DB.Model(&models.Message{}).Where("to_id = ? and status = 0", UserId.(uint)).Count(&commentTotal)
	c.JSON(http.StatusOK, gin.H{
		"like_total":    likeTotal,
		"comment_total": commentTotal,
	})
}

type LikeRecordRes struct {
	ID        uint      `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	TargetID  uint      `json:"target_id"`
	Target    string    `json:"target"`
	FromID    uint      `json:"from_id"`
	ToID      uint      `json:"to_id"`
	Status    int       `json:"status"`
	Content   string    `json:"content"`
	FromUser  userInfo  `json:"from_user"`
	ToUser    userInfo  `json:"to_user"`
}

func TruncateString(s string, maxLength int) string {
	if len(s) > maxLength {
		return s[:maxLength]
	}
	return s
}

func getLikeRecordList(c *gin.Context) {
	UserId, _ := c.Get("UserId")
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
	var likeRecords []models.LikeRecord
	var total int64 = 0
	DB.Preload("FromUser.Role").Preload("ToUser.Role").Order("id desc").Where("to_id = ?", UserId.(uint)).Model(&models.LikeRecord{}).Limit(size).Offset(size * (page - 1)).Find(&likeRecords)
	DB.Where("to_id = ?", UserId.(uint)).Model(&models.LikeRecord{}).Count(&total)
	res := make([]LikeRecordRes, 0)
	for _, item := range likeRecords {
		likeRecordRes := LikeRecordRes{
			ID:        item.ID,
			CreatedAt: item.CreatedAt,
			TargetID:  item.TargetID,
			Target:    item.Target,
			FromID:    item.FromID,
			ToID:      item.ToID,
			Status:    item.Status,
			Content:   item.Content,
			FromUser: userInfo{
				Id:       item.FromUser.ID,
				Avatar:   item.FromUser.Avatar,
				Nickname: item.FromUser.Nickname,
				Role:     item.FromUser.Role.Name,
			},
			ToUser: userInfo{
				Id:       item.ToUser.ID,
				Avatar:   item.ToUser.Avatar,
				Nickname: item.ToUser.Nickname,
				Role:     item.ToUser.Role.Name,
			},
		}
		res = append(res, likeRecordRes)
	}
	//刷新已读
	DB.Model(&models.LikeRecord{}).Where("to_id = ?", UserId.(uint)).Updates(models.LikeRecord{Status: 1})
	c.JSON(http.StatusOK, gin.H{
		"msg":   "Successfully obtained",
		"data":  res,
		"total": total,
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
func getGrammarList(c *gin.Context) {
	result := make([]GrammarRes, 0)
	DB.Model(&models.Grammar{}).Find(&result)
	c.JSON(http.StatusOK, gin.H{
		"data": result,
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
		ID          uint   `json:"id" binding:"required"`
		LearnGroup  int    `json:"learn_group" binding:"required"`
		ReviewGroup int    `json:"review_group" binding:"required"`
		Mode        string `json:"mode" binding:"required"`
		BookID      int    `json:"book_id" binding:"required"`
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
	DB.First(&config, "user_id = ?", UserId)
	c.JSON(http.StatusOK, gin.H{
		"data": config,
		"msg":  "Successfully obtained",
	})
}
