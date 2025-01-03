package handlers

import (
	"bytes"
	"context"
	"easyjapanese/db"
	"easyjapanese/utils"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

type Captcha struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}
type CaptchaHandler struct{}

func (h *CaptchaHandler) CaptchaRoutes(router *gin.Engine) {
	router.GET("/captcha/:capType/:email", h.getCaptcha)
}
func (h *CaptchaHandler) getCaptcha(c *gin.Context) {
	capType := c.Param("capType")
	email := c.Param("email")
	code := utils.GenerateRandomString(6, "number")
	captcha := Captcha{Value: code, Type: capType}
	jsonData, err := json.Marshal(captcha)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"err": err.Error()})
		return
	}
	ctx := context.Background()
	//验证验证码是否发送频繁
	val, err := db.Rdb.TTL(ctx, email).Result()
	if val.Minutes() > 4 {
		c.JSON(http.StatusBadRequest, gin.H{"err": "Frequent requests"})
		return
	}
	//写入数据库
	err = db.Rdb.Set(ctx, email, string(jsonData), 5*60*time.Second).Err()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"err": err.Error()})
		return
	}
	SendCaptcha(code, email, c)
}
func SendCaptcha(code, email string, c *gin.Context) {
	currentDir, err := os.Getwd()
	tmpl, err := template.ParseFiles(filepath.Join(currentDir, "templates/captcha.html"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"err": "Failed to load template"})
		return
	}
	var tpl bytes.Buffer
	data := map[string]string{
		"Captcha": code,
		"Email":   email,
	}
	if err := tmpl.Execute(&tpl, data); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"err": "Failed to render template"})
		return
	}
	emailBody := tpl.String()
	SendEmail("轻松日语验证码", emailBody, email)
	c.JSON(http.StatusOK, gin.H{"msg": "Captcha sent successfully"})
}
