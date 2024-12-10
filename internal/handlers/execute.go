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
	Userhandler.RegisterRoutes(router)
	Captchahandler := &CaptchaHandler{}
	Captchahandler.CaptchaRoutes(router)
}
