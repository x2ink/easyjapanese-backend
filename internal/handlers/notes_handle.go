package handlers

import (
	. "easyjapanese/db"
	"easyjapanese/internal/middleware"
	"easyjapanese/internal/models"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type NotesHandler struct{}

func (h *NotesHandler) NotesRoutes(router *gin.Engine) {
	v1 := router.Group("/notes").Use(middleware.User())
	v1.POST("", h.add)
	v1.DELETE("/:id", h.del)
	v1.DELETE("quote/:id", h.delQuote)
	v1.GET("/info/:id", h.getInfo)
	v1.GET("/like/:type/:id", h.like)
	v1.GET("/list/:id/:page/:size", h.getList)
	v1.GET("/self/:page/:size", h.getSelfList)
}

type NoteRes struct {
	Word      string    `json:"word"`
	Kana      string    `json:"kana"`
	Content   string    `json:"content"`
	User      userInfo  `json:"user"`
	CreatedAt time.Time `json:"created_at"`
	Like      int       `json:"like"`
	ID        uint      `json:"id"`
	Cite      CiteInfo  `json:"cite"`
	WordID    uint      `json:"word_id"`
}

func (h *NotesHandler) getSelfList(c *gin.Context) {
	page, err := strconv.Atoi(c.Param("page"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"err": "The page format is incorrect"})
		return
	}
	size, err := strconv.Atoi(c.Param("size"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"err": "The size format is incorrect"})
		return
	}
	var total int64
	result := make([]NoteRes, 0)
	notes := make([]models.Notes, 0)
	UserId, _ := c.Get("UserId")
	DB.Preload("Word").Preload("Cite.User").Model(&models.Notes{}).Where("user_id= ?", UserId).Limit(size).Offset(size * (page - 1)).Find(&notes)
	DB.Model(&models.Notes{}).Where("user_id= ?", UserId).Count(&total)
	for _, note := range notes {
		var cite CiteInfo
		if note.Cite == nil {
			cite = CiteInfo{}
		} else {
			cite = CiteInfo{
				Content:  note.Cite.Content,
				ID:       &note.Cite.ID,
				Nickname: note.Cite.User.Nickname,
			}
		}
		result = append(result, NoteRes{
			ID:        note.ID,
			Content:   note.Content,
			CreatedAt: note.CreatedAt,
			Like:      note.Like,
			Cite:      cite,
			Word:      note.Word.Word,
			Kana:      note.Word.Kana,
			WordID:    note.WordID,
		})
	}
	c.JSON(http.StatusOK, gin.H{
		"data":  result,
		"total": total,
	})
}
func (h *NotesHandler) getList(c *gin.Context) {
	page, err := strconv.Atoi(c.Param("page"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"err": "The page format is incorrect"})
		return
	}
	size, err := strconv.Atoi(c.Param("size"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"err": "The size format is incorrect"})
		return
	}
	id := c.Param("id")
	var total int64
	result := make([]NoteRes, 0)
	notes := make([]models.Notes, 0)
	DB.Preload("Cite.User").Preload("User.Role").Model(&models.Notes{}).Where("public = ? and word_id= ?", true, id).Limit(size).Offset(size * (page - 1)).Find(&notes)
	DB.Model(&models.Notes{}).Where("public = ? and word_id= ?", true, id).Count(&total)
	for _, note := range notes {
		var cite CiteInfo
		if note.Cite == nil {
			cite = CiteInfo{}
		} else {
			cite = CiteInfo{
				Content:  note.Cite.Content,
				ID:       &note.Cite.ID,
				Nickname: note.Cite.User.Nickname,
			}
		}
		result = append(result, NoteRes{
			ID:      note.ID,
			Content: note.Content,
			User: userInfo{
				Id:       note.User.ID,
				Avatar:   note.User.Avatar,
				Nickname: note.User.Nickname,
				Role:     note.User.Role.Name,
			},
			CreatedAt: note.CreatedAt,
			Like:      note.Like,
			Cite:      cite,
		})
	}
	c.JSON(http.StatusOK, gin.H{
		"data":  result,
		"total": total,
	})
}
func (h *NotesHandler) like(c *gin.Context) {
	id := c.Param("id")
	t := c.Param("type")
	note := models.Notes{}
	result := DB.First(&note, id)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"err": "Not found"})
	} else {
		if t == "unlike" {
			if note.Like > 0 {
				note.Like -= 1
			}
		} else {
			note.Like += 1
		}
		DB.Save(&note)
		c.JSON(http.StatusOK, gin.H{"msg": "Updated successfully"})
	}
}

type NoteInfo struct {
	Content string   `json:"content"`
	Public  bool     `json:"public"`
	ID      uint     `json:"id"`
	Cite    CiteInfo `json:"cite"`
	Word    string   `json:"word"`
	Kana    string   `json:"kana"`
	Meaning []string `json:"meaning"`
}
type CiteInfo struct {
	Content  string `json:"content"`
	ID       *uint  `json:"id"`
	Nickname string `json:"nickname"`
}

func (h *NotesHandler) getInfo(c *gin.Context) {
	id := c.Param("id")
	UserId, _ := c.Get("UserId")
	note := models.Notes{}
	word := models.Jcdict{}
	DB.Preload("Meaning").First(&word, id)
	meanings := make([]string, 0)
	for _, meaning := range word.Meaning {
		meanings = append(meanings, meaning.Meaning)
	}
	result := DB.Preload("Cite.User").Where("word_id = ? and user_id = ?", id, UserId).First(&note)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusOK, gin.H{"data": NoteInfo{
			Word:    note.Word.Word,
			Kana:    note.Word.Kana,
			Meaning: meanings,
		}})
	} else {
		meanings := make([]string, 0)
		for _, meaning := range word.Meaning {
			meanings = append(meanings, meaning.Meaning)
		}

		if note.Cite == nil {
			c.JSON(http.StatusOK, gin.H{"data": NoteInfo{
				Content: note.Content,
				Public:  note.Public,
				ID:      note.ID,
				Cite: CiteInfo{
					Content:  "",
					ID:       nil,
					Nickname: "",
				},
				Word:    word.Word,
				Kana:    word.Kana,
				Meaning: meanings,
			}})
		} else {
			c.JSON(http.StatusOK, gin.H{"data": NoteInfo{
				Content: note.Content,
				Public:  note.Public,
				ID:      note.ID,
				Cite: CiteInfo{
					Content:  note.Cite.Content,
					ID:       &note.Cite.ID,
					Nickname: note.Cite.User.Nickname,
				},
				Word:    word.Word,
				Kana:    word.Kana,
				Meaning: meanings,
			}})
		}

	}
}
func (h *NotesHandler) delQuote(c *gin.Context) {
	UserId, _ := c.Get("UserId")
	id := c.Param("id")
	note := models.Notes{}
	DB.Where("id = ? and user_id = ?", id, UserId).First(&note)
	note.CiteID = nil // 删除引用
	DB.Save(&note)    // 保存更改
	c.JSON(http.StatusOK, gin.H{"msg": "Deleted successfully"})
}
func (h *NotesHandler) del(c *gin.Context) {
	UserId, _ := c.Get("UserId")
	id := c.Param("id")
	DB.Where("id = ? and user_id = ?", id, UserId).Delete(&models.Notes{})
	c.JSON(http.StatusOK, gin.H{"msg": "Deleted successfully"})
}
func (h *NotesHandler) add(c *gin.Context) {
	UserId, _ := c.Get("UserId")
	var Req struct {
		Content string `json:"content"`
		WordID  uint   `json:"word_id"`
		Public  bool   `json:"public"`
		CiteID  uint   `json:"cite_id"`
	}
	if err := c.ShouldBindJSON(&Req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"err": err.Error()})
		return
	}
	note := models.Notes{}
	DB.Where("user_id = ? and word_id = ?", UserId, Req.WordID).First(&note)
	note.Content = Req.Content
	note.WordID = Req.WordID
	note.UserID = UserId.(uint)
	note.Public = Req.Public
	note.CiteID = &Req.CiteID
	DB.Save(&note)
	c.JSON(http.StatusOK, gin.H{"msg": "Submited successfully"})
}
