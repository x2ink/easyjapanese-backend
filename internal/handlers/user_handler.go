package handlers

import (
	"context"
	"easyjapanese/db"
	"easyjapanese/internal/models"
	"easyjapanese/utils"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
)

type UserHandler struct{}

func (h *UserHandler) RegisterRoutes(router *gin.Engine) {
	router.POST("/register", h.Register)
	router.POST("/login/:type", h.Login)
	router.POST("/test", h.Test)
	//rg := router.Group("/user")
	//rg.POST("/register")
}
func (h *UserHandler) Test(c *gin.Context) {
	ctx := context.Background()
	err := db.Rdb.Set(ctx, "mykey", "Hello, Redis!", 60*time.Second).Err()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"err": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"msg": "Key set successfully"})
}
func (h *UserHandler) Login(c *gin.Context) {
	loginType := c.Param("type")
	var Req struct {
		Password string `json:"password"`
		Email    string `json:"email" binding:"required,email"`
		Captcha  string `json:"captcha"`
	}
	if err := c.ShouldBindJSON(&Req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"err": err.Error()})
		return
	}
	if loginType == "pwd" {
		//	密码登录
		password := utils.EncryptionPassword(Req.Password)
		var count int64
		db.DB.Model(&models.Users{}).Where("email=? AND password=?", Req.Email, password).Count(&count)
		if count > 0 {
			c.JSON(http.StatusOK, gin.H{"msg": "Successful login"})
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"msg": "Email or password is wrong", "code": 4002})
		}
	} else if loginType == "capt" {
		ctx := context.Background()
		//	验证码登录
		var capt Captcha
		captcha, err := db.Rdb.Get(ctx, Req.Email).Result()
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"err": "Captcha validation error", "code": 4001})
			return
		}
		err = json.Unmarshal([]byte(captcha), &capt)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"err": err.Error()})
			return
		}
		if capt.Type == "login" && capt.Value == Req.Captcha {
			c.JSON(http.StatusOK, gin.H{"msg": "Successful login"})
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"err": "Captcha validation error", "code": 4001})
		}
	}
}

func (h *UserHandler) Register(c *gin.Context) {
	var Req struct {
		Nickname string `json:"nickname" binding:"required,min=2,max=7"`
		Password string `json:"password" binding:"required"`
		Email    string `json:"email" binding:"required,email"`
		Os       string `json:"os" binding:"required"`
		Device   string `json:"device" binding:"required"`
		Captcha  string `json:"captcha" binding:"required"`
	}
	if err := c.ShouldBindJSON(&Req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"err": err.Error()})
		return
	}
	ip := c.ClientIP()
	password := utils.EncryptionPassword(Req.Password)
	ctx := context.Background()
	//验证验证码是否正确
	var capt Captcha
	captcha, err := db.Rdb.Get(ctx, Req.Email).Result()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"err": "Captcha validation error", "code": 4001})
		return
	}
	err = json.Unmarshal([]byte(captcha), &capt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"err": err.Error()})
		return
	}
	if capt.Type == "register" && capt.Value == Req.Captcha {
		var count int64
		db.DB.Model(&models.Users{}).Where("email=?", Req.Email).Count(&count)
		if count > 0 {
			c.JSON(http.StatusBadRequest, gin.H{"err": "email already exists", "code": 4002})
			return
		}
		user := models.Users{Nickname: Req.Nickname, Email: Req.Email, Password: password, Os: Req.Os, Device: Req.Device, Ip: ip}
		db.DB.Create(&user)
		c.JSON(http.StatusOK, gin.H{
			"msg": "Successfully registered",
		})
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"err": "Captcha validation error", "code": 4001})
	}
}
