package handlers

import (
	. "easyjapanese/db"
	"easyjapanese/internal/middleware"
	"easyjapanese/internal/models"
	"easyjapanese/utils"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type UserHandler struct{}

func (h *UserHandler) UserRoutes(router *gin.Engine) {
	router.POST("/wxlogin", h.wxLogin)
	router.GET("/token", h.resetToken)
	rg := router.Group("/user").Use(middleware.User())
	rg.GET("/info", h.getSimpleUserInfo)
	rg.POST("/info", h.setUserInfo)

}
func (h *UserHandler) wxLogin(c *gin.Context) {
	var Res struct {
		SessionKey string `json:"session_key"`
		OpenID     string `json:"openid"`
	}
	var Req struct {
		Code     string `json:"code"`
		Avatar   string `json:"avatar"`
		Nickname string `json:"nickname"`
		Os       string `json:"os"`
		Device   string `json:"device"`
	}
	if err := c.ShouldBindJSON(&Req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"err": err.Error()})
		return
	}
	url := fmt.Sprintf("https://api.weixin.qq.com/sns/jscode2session?appid=wx4979399128b06403&secret=d4c5c81322267eafa792934452eb4a16&js_code=%s&grant_type=authorization_code", Req.Code)
	resp, err := http.Get(url)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"err": err.Error()})
		return
	}
	defer func(Body io.ReadCloser) {
		if err := Body.Close(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"err": err.Error()})
			return
		}
	}(resp.Body)
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"err": err.Error()})
		return
	}
	if err := json.Unmarshal(body, &Res); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"err": err.Error()})
		return
	}
	if Res.OpenID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"err": "Code parsing failed"})
		return
	}
	var user models.Users
	result := DB.First(&user, "open_id=?", Res.OpenID).Error
	if errors.Is(result, gorm.ErrRecordNotFound) {
		ip := c.ClientIP()
		newUser := models.Users{Avatar: Req.Avatar, Nickname: Req.Nickname, OpenID: Res.OpenID, Os: Req.Os, Device: Req.Device, Ip: ip}
		DB.Create(&newUser)
		newUser.Os = Req.Os
		newUser.Device = Req.Device
		DB.Create(&models.UserConfig{UserID: newUser.ID})
		wxGetToken(newUser, c)
	} else {
		user.Os = Req.Os
		user.Device = Req.Device
		wxGetToken(user, c)
	}
}
func wxGetToken(user models.Users, c *gin.Context) {
	DB.Save(&user)
	tokenData := utils.Token{
		RoleId: user.RoleID,
		UserId: user.ID,
	}
	token := utils.EncryptToken(tokenData)
	c.JSON(http.StatusOK, gin.H{"msg": "Successful login", "data": token})
}

func (h *UserHandler) setUserInfo(c *gin.Context) {
	UserId, _ := c.Get("UserId")
	var Req struct {
		Nickname string `json:"nickname"`
		Avatar   string `json:"avatar"`
	}
	if err := c.ShouldBindJSON(&Req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"err": err.Error()})
		return
	}
	var user models.Users
	result := DB.First(&user, UserId)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"err": "User not exist"})
	} else {
		user.Nickname = Req.Nickname
		user.Avatar = Req.Avatar
		DB.Debug().Save(&user)
		c.JSON(http.StatusOK, gin.H{"msg": "Update success"})
	}
}
func (h *UserHandler) resetToken(c *gin.Context) {
	UserId := c.Query("userId")
	token := c.GetHeader("Authorization")
	user := models.Users{}
	err := DB.Select("role_id", "id").First(&user, UserId).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusNotFound, gin.H{
			"msg": "Id does not exist",
		})
		return
	}
	_, err = utils.DecryptToken(token)
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
		c.JSON(http.StatusOK, gin.H{"msg": "Successful reset", "data": token})
		return
	}
}

func (h *UserHandler) getSimpleUserInfo(c *gin.Context) {
	UserId, _ := c.Get("UserId")
	var User models.Users
	err := DB.Preload("Role").Select("created_at", "nickname", "email", "avatar", "ip", "role_id", "id").First(&User, UserId).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusNotFound, gin.H{
			"msg": "User does not exist",
		})
		return
	}
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"err": "Address acquisition failed"})
	}
	c.JSON(http.StatusOK, gin.H{"msg": "Successfully obtained", "data": map[string]interface{}{
		"id":         User.ID,
		"nickname":   User.Nickname,
		"email":      User.Email,
		"avatar":     User.Avatar,
		"role":       User.Role.Name,
		"created_at": User.CreatedAt,
	}})
}
