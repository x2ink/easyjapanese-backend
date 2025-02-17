package handlers

import (
	. "easyjapanese/db"
	"easyjapanese/internal/middleware"
	"easyjapanese/internal/models"
	"errors"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"net/http"
	"strconv"
	"time"
)

type CommentHandler struct{}

func (h *CommentHandler) CommentRoutes(router *gin.Engine) {
	v1 := router.Group("/comment").Use(middleware.User())
	v1.POST("", h.add)
	v1.GET("/:target/:target_id/:page/:size", h.getList)
	v1.POST("/like/:id", h.like)
	v1.GET("/child/:parent_id/:page/:size", h.getChild)
}
func (h *CommentHandler) getChild(c *gin.Context) {
	parentId, err := strconv.ParseUint(c.Param("parent_id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ParentId format"})
		return
	}
	page, err := strconv.Atoi(c.Param("page"))
	UserId, _ := c.Get("UserId")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"err": "The page format is incorrect"})
		return
	}
	size, err := strconv.Atoi(c.Param("size"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"err": "The size format is incorrect"})
		return
	}
	var comments []models.Comment
	DB.Debug().Order("id desc").Preload("ToUser").Preload("Images").Preload("Like").Preload("FromUser").Model(&models.Comment{}).Where("parent_id = ?", parentId).Limit(size).Offset(size * (page - 1)).Find(&comments)
	var total int64
	DB.Model(&models.Comment{}).Where("parent_id = ?", parentId).Count(&total)
	var listres []listRes
	for _, comment := range comments {
		var images []string
		for _, image := range comment.Images {
			images = append(images, image.Url)
		}
		res := listRes{
			Id:      comment.ID,
			Content: comment.Content,
			ToUser: userInfo{
				Id:       comment.ToUser.ID,
				Avatar:   comment.ToUser.Avatar,
				Nickname: comment.ToUser.Nickname,
			},
			FromUser: userInfo{
				Id:       comment.FromUser.ID,
				Avatar:   comment.FromUser.Avatar,
				Nickname: comment.FromUser.Nickname,
			},
			Images:    images,
			CreatedAt: comment.CreatedAt,
			LikeCount: len(comment.Like),
			HasLike:   HasLike(comment.Like, UserId.(uint)),
		}
		listres = append(listres, res)
	}
	c.JSON(http.StatusOK, gin.H{"data": listres, "total": total})
}
func (h *CommentHandler) like(c *gin.Context) {
	targetId, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}
	UserId, _ := c.Get("UserId")
	var like models.CommentLike
	result := DB.Where("target_id = ? AND user_id = ?", uint(targetId), UserId).First(&like).Error
	if errors.Is(result, gorm.ErrRecordNotFound) {
		DB.Create(&models.CommentLike{
			TargetID: uint(targetId),
			UserID:   UserId.(uint),
		})
		var Comment models.Comment
		DB.First(&Comment, uint(targetId))
		DB.Create(&models.LikeRecord{
			Content:  TruncateString(Comment.Content, 100),
			TargetID: uint(targetId),
			Target:   "comment",
			ToID:     Comment.From,
			FromID:   UserId.(uint),
		})
		c.JSON(http.StatusOK, gin.H{"msg": "like"})
		return
	} else {
		DB.Unscoped().Delete(&like)
		DB.Unscoped().Delete(&models.LikeRecord{}, "target = ? and target_id = ?", "comment", uint(targetId))
		c.JSON(http.StatusOK, gin.H{"msg": "dislike"})
		return
	}
}

type listRes struct {
	Id         uint      `json:"id"`
	Content    string    `json:"content"`
	ToUser     userInfo  `json:"to_user"`
	FromUser   userInfo  `json:"from_user"`
	Images     []string  `json:"images"`
	CreatedAt  time.Time `json:"created_at"`
	Children   []listRes `json:"children"`
	LikeCount  int       `json:"like_count"`
	HasLike    bool      `json:"has_like"`
	ChildCount int       `json:"child_count"`
}

func HasLike(likes []models.CommentLike, userId uint) bool {
	for _, like := range likes {
		if like.UserID == userId {
			return true
		}
	}
	return false
}
func (h *CommentHandler) getList(c *gin.Context) {
	page, err := strconv.Atoi(c.Param("page"))
	UserId, _ := c.Get("UserId")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"err": "The page format is incorrect"})
		return
	}
	size, err := strconv.Atoi(c.Param("size"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"err": "The size format is incorrect"})
		return
	}
	target := c.Param("target")
	targetId, err := strconv.Atoi(c.Param("target_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"err": "The page format is incorrect"})
		return
	}
	var comments []models.Comment
	DB.Order("id desc").Preload("ToUser").Preload("Images").Preload("Like").Preload("FromUser").Preload("Children", func(db *gorm.DB) *gorm.DB {
		return db.Order("id desc").Preload("ToUser").Preload("FromUser").Preload("Images").Preload("Like")
	}).Model(&models.Comment{}).Where("target_id = ? AND target = ? AND parent_id is NULL", targetId, target).Limit(size).Offset(size * (page - 1)).Find(&comments)
	var total int64
	DB.Model(&models.Comment{}).Where("target_id = ? AND target = ? AND parent_id is NULL", targetId, target).Count(&total)
	listres := make([]listRes, 0)
	for _, comment := range comments {
		children := make([]listRes, 0)
		for k, child := range comment.Children {
			if k >= 10 {
				break
			}
			var childimage []string
			for _, image := range child.Images {
				childimage = append(childimage, image.Url)
			}
			children = append(children, listRes{
				Id:      child.ID,
				Content: child.Content,
				ToUser: userInfo{
					Id:       child.ToUser.ID,
					Avatar:   child.ToUser.Avatar,
					Nickname: child.ToUser.Nickname,
				},
				FromUser: userInfo{
					Id:       child.FromUser.ID,
					Avatar:   child.FromUser.Avatar,
					Nickname: child.FromUser.Nickname,
				},
				CreatedAt: comment.CreatedAt,
				Images:    childimage,
				LikeCount: len(child.Like),
				HasLike:   HasLike(child.Like, UserId.(uint)),
			})
		}
		var images []string
		for _, image := range comment.Images {
			images = append(images, image.Url)
		}
		res := listRes{
			ChildCount: len(comment.Children),
			Id:         comment.ID,
			Content:    comment.Content,
			ToUser: userInfo{
				Id:       comment.ToUser.ID,
				Avatar:   comment.ToUser.Avatar,
				Nickname: comment.ToUser.Nickname,
			},
			FromUser: userInfo{
				Id:       comment.FromUser.ID,
				Avatar:   comment.FromUser.Avatar,
				Nickname: comment.FromUser.Nickname,
			},
			Images:    images,
			CreatedAt: comment.CreatedAt,
			Children:  children,
			LikeCount: len(comment.Like),
			HasLike:   HasLike(comment.Like, UserId.(uint)),
		}
		listres = append(listres, res)
	}
	c.JSON(http.StatusOK, gin.H{"data": listres, "total": total})
}

type addReq struct {
	Content  string   `json:"content"`
	To       int64    `json:"to"`
	Target   string   `json:"target"`
	TargetID int      `json:"target_id"`
	ParentID int      `json:"parent_id"`
	Images   []string `json:"images"`
}

func (h *CommentHandler) add(c *gin.Context) {
	var Req addReq
	UserId, _ := c.Get("UserId")
	if err := c.ShouldBindJSON(&Req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"err": err.Error()})
		return
	}
	var ParentID *int
	content := ""
	if Req.ParentID == 0 {
		ParentID = nil
		var trend models.Trend
		DB.First(&trend, Req.TargetID)
		content = trend.Content
	} else {
		ParentID = &Req.ParentID
		var comment models.Comment
		DB.First(&comment, ParentID)
		content = comment.Content
	}
	comment := &models.Comment{
		Content:  Req.Content,
		To:       uint(Req.To),
		From:     UserId.(uint),
		Target:   Req.Target,
		TargetID: Req.TargetID,
		ParentID: ParentID,
	}
	DB.Create(&comment)
	if uint(Req.To) == UserId.(uint) {
		sendMessage(uint(Req.To), UserId.(uint), comment.ID, content)
	}
	for _, v := range Req.Images {
		DB.Create(&models.CommentImage{TargetID: comment.ID, Url: v})
	}
	c.JSON(http.StatusOK, gin.H{"msg": "Created successfully", "data": comment.ID})
}
