package handlers

import (
	"bytes"
	"context"
	"easyjapanese/config"
	"easyjapanese/internal/middleware"
	"encoding/base64"
	"fmt"
	"net/http"

	"github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss"
	"github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss/credentials"
	"github.com/gin-gonic/gin"
)

type FileHandler struct{}

var (
	region   = "cn-shanghai"
	bucket   = "jpx2ink"
	provider = credentials.NewStaticCredentialsProvider(config.AliAccessKeyId, config.AliAccessKeySecret)
)

func (h *FileHandler) FileRoutes(router *gin.Engine) {
	router.POST("/base64", middleware.User(), h.uploadBase64)
	router.POST("/upload", middleware.User(), h.uploadFile)
}
func (h *FileHandler) uploadFile(c *gin.Context) {
	file, header, err := c.Request.FormFile("file")
	const maxSize = 10 * 1024 * 1024
	if header.Size > maxSize {
		c.JSON(http.StatusRequestEntityTooLarge, gin.H{
			"err": "The file is too large",
		})
		return
	}
	fileName := c.PostForm("file_name")
	if fileName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"err": "file_name 不能为空"})
		return
	}
	cfg := oss.LoadDefaultConfig().WithCredentialsProvider(provider).WithRegion(region)
	client := oss.NewClient(cfg)

	if err != nil {
		c.String(http.StatusBadRequest, fmt.Sprintf("获取文件失败: %v", err))
		return
	}
	defer file.Close()
	request := &oss.PutObjectRequest{
		Bucket: oss.Ptr(bucket),
		Key:    oss.Ptr(fileName),
		Body:   file,
	}
	_, err = client.PutObject(context.TODO(), request)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"err": err.Error()})
		return
	}
	imageURL := fmt.Sprintf("https://%s.oss-%s.aliyuncs.com/%s", bucket, region, fileName)
	c.JSON(http.StatusOK, gin.H{"url": imageURL})
}

func (h *FileHandler) uploadBase64(c *gin.Context) {
	const maxSize = 10 * 1024 * 1024
	if c.Request.ContentLength > maxSize {
		c.JSON(http.StatusRequestEntityTooLarge, gin.H{
			"err": "The file is too large",
		})
		return
	}
	var Req struct {
		Data     string `json:"data"`
		FileName string `json:"file_name"`
	}
	if err := c.ShouldBindJSON(&Req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"err": err.Error()})
		return
	}
	cfg := oss.LoadDefaultConfig().WithCredentialsProvider(provider).WithRegion(region)
	client := oss.NewClient(cfg)
	imageData, _ := base64.StdEncoding.DecodeString(Req.Data)
	request := &oss.PutObjectRequest{
		Bucket: oss.Ptr(bucket),
		Key:    oss.Ptr(Req.FileName),
		Body:   bytes.NewReader(imageData),
	}
	_, err := client.PutObject(context.TODO(), request)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"err": err.Error()})
		return
	}
	imageURL := fmt.Sprintf("https://%s.oss-%s.aliyuncs.com/%s", bucket, region, Req.FileName)
	c.JSON(http.StatusOK, gin.H{"url": imageURL})
}
