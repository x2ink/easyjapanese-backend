package handlers

import (
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
}
