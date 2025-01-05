package handlers

import (
	. "easyjapanese/db"
	"easyjapanese/internal/models"
	"github.com/gin-gonic/gin"
	"net/http"
)

type SectionHandler struct{}

func (h *SectionHandler) SectionRoutes(router *gin.Engine) {
	v1 := router.Group("/section")
	v1.GET("/:target", h.get)
}
func (h *SectionHandler) get(c *gin.Context) {
	target := c.Param("target")
	var Res []models.Section
	DB.Where("target = ?", target).Find(&Res)
	c.JSON(http.StatusOK, gin.H{
		"msg":  "Successfully obtained",
		"data": Res,
	})
}
