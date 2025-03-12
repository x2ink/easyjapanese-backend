package handlers

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"os"
	"path/filepath"
)

type FileHandler struct{}

func (h *FileHandler) FileRoutes(router *gin.Engine) {
	router.POST("/upload", h.upload)
}
func (h *FileHandler) upload(c *gin.Context) {
	file, _ := c.FormFile("file")
	if file.Size > 1024*1024*10 {
		c.JSON(http.StatusBadRequest, gin.H{"err": "The file is too large"})
		return
	}
	rootDir, _ := os.Getwd()
	path := fmt.Sprintf("%s/%s", rootDir, "file")
	dst := filepath.Join(path, file.Filename)
	err := c.SaveUploadedFile(file, dst)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"err": err.Error()})
		return
	}
	url := fmt.Sprintf("http://%s/file/%s", c.Request.Host, file.Filename)
	c.JSON(http.StatusOK, gin.H{"msg": "上传成功", "data": url})
	return
}
