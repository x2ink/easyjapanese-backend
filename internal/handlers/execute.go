package handlers

import (
	. "easyjapanese/db"
	"easyjapanese/internal/middleware"
	"easyjapanese/internal/models"
	"easyjapanese/utils"
	"github.com/gin-gonic/gin"
	"net/http"
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
	router.GET("/config", middleware.User(), getUserConfig)
	router.POST("/config", middleware.User(), setUserConfig)
	router.GET("/verbtrans/:word", verbTrans)
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
