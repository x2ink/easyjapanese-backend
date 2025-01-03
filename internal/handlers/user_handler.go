package handlers

import (
	"context"
	. "easyjapanese/db"
	"easyjapanese/internal/middleware"
	"easyjapanese/internal/models"
	"easyjapanese/utils"
	"encoding/json"
	"errors"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"net/http"
)

type UserHandler struct{}

func (h *UserHandler) UserRoutes(router *gin.Engine) {
	router.POST("/register", h.register)
	router.POST("/test", h.test)
	router.POST("/login/:type", h.login)
	router.GET("/token/reset/:userId", h.resetToken)
	rg := router.Group("/user").Use(middleware.User())
	rg.GET("/info/simple", h.getSimpleUserInfo)
}
func (h *UserHandler) test(c *gin.Context) {
	var results []map[string]interface{}
	DB.Model(&models.Users{}).Find(&results)
	c.JSON(http.StatusOK, results)
}
func (h *UserHandler) resetToken(c *gin.Context) {
	UserId := c.Param("userId")
	token := c.GetHeader("Authorization")
	_, err := utils.DecryptToken(token)
	user := models.Users{}
	err = DB.Select("role_id", "id").First(&user, UserId).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusNotFound, gin.H{
			"msg": "Id does not exist",
		})
		return
	}
	if err != nil {
		if err.Error() == "token has invalid claims: token is expired" {
			tokenData := utils.Token{
				RoleId: user.RoleID,
				UserId: user.ID,
			}
			token := utils.EncryptToken(tokenData)
			c.JSON(http.StatusResetContent, gin.H{"msg": "Successful reset", "data": token})
			return
		}
		c.JSON(http.StatusUnauthorized, gin.H{
			"msg": "Authentication failed",
		})
		return
	} else {
		tokenData := utils.Token{
			RoleId: user.RoleID,
			UserId: user.ID,
		}
		token := utils.EncryptToken(tokenData)
		c.JSON(http.StatusResetContent, gin.H{"msg": "Successful reset", "data": token})
		return
	}
}
func (h *UserHandler) getSimpleUserInfo(c *gin.Context) {
	UserId, _ := c.Get("UserId")
	var User models.Users
	err := DB.Preload("Role").Select("nickname", "email", "avatar", "ip", "role_id").First(&User, UserId).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusNotFound, gin.H{
			"msg": "User does not exist",
		})
		return
	}
	address, err := utils.GetIpAddress(User.Ip)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"err": "Address acquisition failed"})
	}
	c.JSON(http.StatusOK, gin.H{"msg": "Successfully obtained", "data": map[string]interface{}{
		"nickname": User.Nickname,
		"email":    User.Email,
		"avatar":   User.Avatar,
		"address":  address,
		"role":     User.Role.Name,
	}})
}

type LoginReq struct {
	Password string `json:"password"`
	Email    string `json:"email" binding:"required,email"`
	Captcha  string `json:"captcha"`
	Os       string `json:"os" binding:"required"`
	Device   string `json:"device" binding:"required"`
}

func (h *UserHandler) login(c *gin.Context) {
	loginType := c.Param("type")
	var Req LoginReq
	if err := c.ShouldBindJSON(&Req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"err": err.Error()})
		return
	}
	if loginType == "pwd" {
		password := utils.EncryptionPassword(Req.Password)
		err := DB.Select("id").Find(&models.Users{}, "email=? AND password=?", Req.Email, password).First(&models.Users{}).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusBadRequest, gin.H{"msg": "Email or password is wrong", "code": 4002})
			return
		}
		loginSuccess(Req, c)
	} else if loginType == "capt" {
		ctx := context.Background()
		var capt Captcha
		captcha, err := Rdb.Get(ctx, Req.Email).Result()
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
			loginSuccess(Req, c)
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"err": "Captcha validation error", "code": 4001})
		}
	}
}
func loginSuccess(req LoginReq, c *gin.Context) {
	user := models.Users{}
	DB.Select("role_id", "id").Model(&models.Users{}).Where("email=?", req.Email).First(&user)
	DB.Save(&models.Users{ID: user.ID, Os: req.Os, Device: req.Device})
	tokenData := utils.Token{
		RoleId: user.RoleID,
		UserId: user.ID,
	}
	token := utils.EncryptToken(tokenData)
	c.JSON(http.StatusOK, gin.H{"msg": "Successful login", "data": token})
}
func (h *UserHandler) register(c *gin.Context) {
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
	captcha, err := Rdb.Get(ctx, Req.Email).Result()
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
		DB.Select("id").Find(&models.Users{}, "email=?", Req.Email).Count(&count)
		if count > 0 {
			c.JSON(http.StatusBadRequest, gin.H{"err": "email already exists", "code": 4002})
			return
		}
		user := models.Users{Nickname: Req.Nickname, Email: Req.Email, Password: password, Os: Req.Os, Device: Req.Device, Ip: ip}
		DB.Create(&user)
		tokenData := utils.Token{
			RoleId: 1,
			UserId: user.ID,
		}
		token := utils.EncryptToken(tokenData)
		c.JSON(http.StatusOK, gin.H{
			"msg":  "Successfully registered",
			"data": token,
		})
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"err": "Captcha validation error", "code": 4001})
	}
}
