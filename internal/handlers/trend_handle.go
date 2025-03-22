package handlers

import (
	. "easyjapanese/db"
	"easyjapanese/internal/middleware"
	"easyjapanese/internal/models"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"net/http"
	"strconv"
	"time"
)

type TrendHandler struct{}
type userInfo struct {
	Id       uint   `json:"id"`
	Avatar   string `json:"avatar"`
	Nickname string `json:"nickname"`
	Role     string `json:"role"`
}

func (userInfo) TableName() string {
	return "users"
}

func (h *TrendHandler) TrendRoutes(router *gin.Engine) {
	router.GET("/section", h.getSection)
	router.GET("/mytrend/:page/:size", middleware.User(), h.getMyList)
	v2 := router.Group("/trend").Use(middleware.User())
	{
		v2.POST("", h.addTrend)
		v2.GET("/info/:id", h.getInfo)
		v2.GET("/:section/:page/:size", h.getList)
		v2.GET("/search/:page/:size/:val", h.searchTrend)
		v2.DELETE("/:id", h.deleteTrend)
		v2.POST("/like/:id", h.likeTrend)
	}
	v3 := router.Group("/comment").Use(middleware.User())
	{
		v3.POST("", h.addComment)
		v3.PUT("/top/:id", h.topComment)
		v3.DELETE("/:id", h.delComment)
		v3.GET("/:trendid/:sort/:page/:size", h.getCommentList)
		v3.GET("/:trendid/:sort/:page/:size/:hideid", h.getCommentList)
		v3.POST("/like/:id", h.likeComment)
		v3.GET("/child/:parent_id/:page/:size/:sort", h.getChild)
		//v1.POST("/getone", h.getOne)
		//v1.GET("/:target/:target_id/:page/:size/:sort/:hide_id", h.getList)
		//
		//v1.GET("/child/:parent_id/:page/:size/:sort", h.getChild)
	}
	//
	//v1.GET("/search/:page/:size/:val", h.search)
	//
	//v1.GET("/mylist/:page/:size", h.getMyList)
	//v1.POST("/like/:id", h.like)
}
func (h *TrendHandler) getChild(c *gin.Context) {
	parentId, err := strconv.Atoi(c.Param("parent_id"))
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
	sort := c.Param("sort")
	if sort == "time" {
		DB.Order("id desc").Preload("ToUser").Preload("Images").Preload("FromUser").Preload("Like").Model(&models.Comment{}).Where("parent_id = ?", parentId).Limit(size).Offset(size * (page - 1)).Find(&comments)
	} else {
		DB.Order("likenum desc,id desc").Preload("ToUser").Preload("Images").Preload("FromUser").Preload("Like").Model(&models.Comment{}).Where("parent_id = ?", parentId).Limit(size).Offset(size * (page - 1)).Find(&comments)
	}
	var total int64
	DB.Model(&models.Comment{}).Where("parent_id = ?", parentId).Count(&total)
	var listres []commentRes
	for _, comment := range comments {
		var images []string
		for _, image := range comment.Images {
			images = append(images, image.Url)
		}
		res := commentRes{
			ToComment: comment.ToComment,
			Id:        comment.ID,
			Content:   comment.Content,
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
			LikeCount: comment.Likenum,
			HasLike:   HasLike(comment.Like, UserId.(uint)),
		}
		listres = append(listres, res)
	}
	c.JSON(http.StatusOK, gin.H{"data": listres, "total": total})
}

type commentRes struct {
	Id         uint         `json:"id"`
	Content    string       `json:"content"`
	ToUser     userInfo     `json:"to_user"`
	FromUser   userInfo     `json:"from_user"`
	Images     []string     `json:"images"`
	CreatedAt  time.Time    `json:"created_at"`
	Children   []commentRes `json:"children"`
	LikeCount  int          `json:"like_count"`
	ToComment  uint         `json:"to_comment"`
	HasLike    bool         `json:"has_like"`
	ChildCount int          `json:"child_count"`
	ParentId   *uint        `json:"parent_id"`
	Top        bool         `json:"top"`
}

func (h *TrendHandler) likeComment(c *gin.Context) {
	commentId, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}
	UserId, _ := c.Get("UserId")
	var like models.CommentLike
	result := DB.Where("comment_id = ? AND user_id = ?", uint(commentId), UserId).First(&like).Error
	var Comment models.Comment
	DB.First(&Comment, uint(commentId))
	if errors.Is(result, gorm.ErrRecordNotFound) {
		DB.Create(&models.CommentLike{
			CommentID: uint(commentId),
			UserID:    UserId.(uint),
		})
		Comment.Likenum = Comment.Likenum + 1
		DB.Save(&Comment)
		path := fmt.Sprintf("/trendpages/trenddetail/trenddetail?id=%d", Comment.TrendID)
		msg := Message{
			ToID:    Comment.From,
			FromID:  UserId.(uint),
			Content: captureContent(Comment.Content),
			Title:   "点赞了你的评论",
			Path:    path,
			Form:    "like",
		}
		SendMessage(msg)
		c.JSON(http.StatusOK, gin.H{"msg": "like"})
		return
	} else {
		DB.Unscoped().Delete(&like)
		Comment.Likenum = Comment.Likenum - 1
		DB.Save(&Comment)
		c.JSON(http.StatusOK, gin.H{"msg": "dislike"})
		return
	}
}
func HasLike(likes []models.CommentLike, userId uint) bool {
	for _, like := range likes {
		if like.UserID == userId {
			return true
		}
	}
	return false
}
func (h *TrendHandler) getCommentList(c *gin.Context) {
	hideIds := make([]int, 0)
	hideId, err := strconv.Atoi(c.Param("hideid"))
	if err == nil {
		hideIds = append(hideIds, hideId)
	}
	UserId, _ := c.Get("UserId")
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
	trendId, err := strconv.Atoi(c.Param("trendid"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"err": "The page format is incorrect"})
		return
	}
	var comments []models.Comment
	sort := c.Param("sort")
	if sort == "time" {
		DB.Order("level desc,id desc").Preload("ToUser").Preload("Images").Preload("FromUser").Preload("Like").Preload("Children", func(db *gorm.DB) *gorm.DB {
			return db.Order("id desc").Preload("ToUser").Preload("FromUser").Preload("Images").Preload("Like")
		}).Model(&models.Comment{}).Where("trend_id = ? AND parent_id is NULL AND id not in ?", trendId, hideIds).Limit(size).Offset(size * (page - 1)).Find(&comments)
	} else {
		DB.Order("level desc,likenum desc,id desc").Preload("ToUser").Preload("Images").Preload("FromUser").Preload("Like").Preload("Children", func(db *gorm.DB) *gorm.DB {
			return db.Order("likenum desc,id desc").Preload("ToUser").Preload("FromUser").Preload("Images").Preload("Like")
		}).Model(&models.Comment{}).Where("trend_id = ? AND parent_id is NULL AND id not in ?", trendId, hideIds).Limit(size).Offset(size * (page - 1)).Find(&comments)
	}
	var total int64
	DB.Model(&models.Comment{}).Where("trend_id = ? AND parent_id is NULL", trendId).Count(&total)
	listres := make([]commentRes, 0)
	for _, comment := range comments {
		children := make([]commentRes, 0)
		for k, child := range comment.Children {
			if k >= 10 {
				break
			}
			var childimage []string
			for _, image := range child.Images {
				childimage = append(childimage, image.Url)
			}
			children = append(children, commentRes{
				ToComment: child.ToComment,
				Id:        child.ID,
				Content:   child.Content,
				ParentId:  child.ParentID,
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
				CreatedAt: child.CreatedAt,
				Images:    childimage,
				LikeCount: child.Likenum,
				HasLike:   HasLike(child.Like, UserId.(uint)),
			})
		}
		var images []string
		for _, image := range comment.Images {
			images = append(images, image.Url)
		}
		var top bool
		if comment.Level == 0 {
			top = false
		} else {
			top = true
		}
		res := commentRes{
			ToComment:  comment.ToComment,
			Top:        top,
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
			ParentId:  comment.ParentID,
			Images:    images,
			CreatedAt: comment.CreatedAt,
			Children:  children,
			LikeCount: comment.Likenum,
			HasLike:   HasLike(comment.Like, UserId.(uint)),
		}
		listres = append(listres, res)
	}
	c.JSON(http.StatusOK, gin.H{"data": listres, "total": total})
}
func (h *TrendHandler) addComment(c *gin.Context) {
	var Req struct {
		Content   string   `json:"content"`
		To        uint     `json:"to"`
		TrentID   uint     `json:"trend_id"`
		ParentID  uint     `json:"parent_id"`
		Images    []string `json:"images"`
		ToComment uint     `json:"to_comment"`
	}
	UserId, _ := c.Get("UserId")
	if err := c.ShouldBindJSON(&Req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"err": err.Error()})
		return
	}
	trend := models.Trend{}
	DB.First(&trend, Req.TrentID)
	var ParentID *uint
	msg := Message{
		ToID:    trend.UserID,
		FromID:  UserId.(uint),
		Content: "",
		Form:    "msg",
	}
	if Req.ParentID == 0 {
		ParentID = nil
		msg.Path = fmt.Sprintf("/trendpages/trenddetail/trenddetail?id=%d", Req.TrentID)
		msg.Title = "回复了你的动态"
	} else {
		ParentID = &Req.ParentID
		msg.Path = fmt.Sprintf("/trendpages/trenddetail/trenddetail?id=%d", Req.TrentID)
		msg.Title = "回复了你的评论"
	}
	comment := &models.Comment{
		Content:   Req.Content,
		To:        Req.To,
		From:      UserId.(uint),
		TrendID:   Req.TrentID,
		ParentID:  ParentID,
		ToComment: Req.ToComment,
	}
	var msgId uint
	if Req.To != UserId.(uint) {
		msgId = SendMessage(msg)
	}
	DB.Create(&comment)
	if Req.To != UserId.(uint) {
		msgContent := struct {
			ParentID  uint   `json:"parent_id"`
			Title     string `json:"title"`
			Content   string `json:"content"`
			Like      bool   `json:"like"`
			CommentId uint   `json:"comment_id"`
			TrendId   uint   `json:"trend_id"`
		}{
			Title:     Req.Content,
			CommentId: comment.ID,
			TrendId:   Req.TrentID,
		}
		if Req.ParentID == 0 {
			msgContent.Content = captureContent(trend.Content)
			msgContent.ParentID = comment.ID
		} else {
			var commentContent models.Comment
			DB.First(&commentContent, Req.ParentID)
			msgContent.Content = captureContent(commentContent.Content)
			msgContent.ParentID = Req.ParentID
		}
		jsonData, _ := json.Marshal(msgContent)
		DB.Model(&models.Message{}).Where("id = ?", msgId).Update("content", string(jsonData))
	}
	for _, v := range Req.Images {
		DB.Create(&models.CommentImage{CommentID: comment.ID, Url: v})
	}
	c.JSON(http.StatusOK, gin.H{"msg": "Created successfully", "data": comment.ID})
}

func (h *TrendHandler) getSection(c *gin.Context) {
	var Res []models.Section
	DB.Find(&Res)
	c.JSON(http.StatusOK, gin.H{
		"msg":  "Successfully obtained",
		"data": Res,
	})
}
func captureContent(content string) string {
	if len(content) < 20 {
		return content
	} else {
		return content[:20] + "..."
	}
}
func (h *TrendHandler) likeTrend(c *gin.Context) {
	trendId, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"err": "The page format is incorrect"})
		return
	}
	UserId, _ := c.Get("UserId")
	var like models.TrendLike
	result := DB.Where("trend_id = ? AND user_id = ?", uint(trendId), UserId).First(&like).Error
	var Trend models.Trend
	DB.First(&Trend, uint(trendId))
	if errors.Is(result, gorm.ErrRecordNotFound) {
		DB.Create(&models.TrendLike{
			TrendID: uint(trendId),
			UserID:  UserId.(uint),
		})
		Trend.Likenum = Trend.Likenum + 1
		DB.Save(&Trend)
		path := fmt.Sprintf("/trendpages/trenddetail/trenddetail?id=%d", trendId)
		msg := Message{
			ToID:    Trend.UserID,
			FromID:  UserId.(uint),
			Content: captureContent(Trend.Content),
			Title:   "点赞了你的动态",
			Path:    path,
			Form:    "like",
		}
		SendMessage(msg)
		c.JSON(http.StatusOK, gin.H{"msg": "like"})
		return
	} else {
		DB.Unscoped().Delete(&like)
		Trend.Likenum = Trend.Likenum - 1
		DB.Save(&Trend)
		c.JSON(http.StatusOK, gin.H{"msg": "dislike"})
		return
	}
}

func trendList(c *gin.Context, trends []TrendCount, total int64) {
	searchRes := make([]trendResp, 0)
	for _, trend := range trends {
		images := make([]string, 0)
		for _, image := range trend.Images {
			images = append(images, image.Url)
		}
		trendRes := trendResp{
			Comment:   trend.Comment,
			Id:        trend.ID,
			Images:    images,
			Content:   trend.Content,
			Browse:    trend.Browse,
			Like:      trend.Likenum,
			CreatedAt: trend.CreatedAt,
			SectionId: trend.SectionID,
			User: userInfo{
				Id:       trend.UserID,
				Avatar:   trend.User.Avatar,
				Nickname: trend.User.Nickname,
				Role:     trend.User.Role.Name,
			},
		}
		searchRes = append(searchRes, trendRes)
	}
	c.JSON(http.StatusOK, gin.H{
		"msg":   "Successfully obtained",
		"data":  searchRes,
		"total": total,
	})
}

type trendResp struct {
	Comment   int       `json:"comment"`
	Content   string    `json:"content"`
	Browse    int       `json:"browse"`
	Like      int       `json:"like"`
	SectionId int       `json:"section_id"`
	Images    []string  `json:"images"`
	CreatedAt time.Time `json:"created_at"`
	User      userInfo  `json:"user"`
	Id        uint      `json:"id"`
	My        bool      `json:"my"`
	Has       bool      `json:"has"`
}
type TrendCount struct {
	models.Trend
	Comment int `json:"comment"`
}

func (h *TrendHandler) getMyList(c *gin.Context) {
	UserId, _ := c.Get("UserId")
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
	trends := make([]TrendCount, 0)
	DB.Preload("User.Role").Preload("Images").Order("id desc").Model(&models.Trend{}).Select("COUNT(comment.id) as comment,trend.*").Joins("left join comment on  comment.trend_id=trend.id").Where("user_id = ?", UserId).Group("trend.id").Limit(size).Offset(size * (page - 1)).Find(&trends)
	DB.Model(&models.Trend{}).Where("user_id = ?", UserId).Count(&total)
	trendList(c, trends, total)
}
func (h *TrendHandler) searchTrend(c *gin.Context) {
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
	val := c.Param("val")
	var total int64
	trends := make([]TrendCount, 0)
	likeitem := fmt.Sprintf("%%%s%%", val)
	DB.Preload("User.Role").Preload("Images").Order("id desc").Model(&models.Trend{}).Select("COUNT(comment.id) as comment,trend.*").Joins("left join comment on comment.trend_id=trend.id").Where("trend.content Like ?", likeitem).Group("trend.id").Limit(size).Offset(size * (page - 1)).Find(&trends)
	DB.Model(&models.Trend{}).Where("trend.content Like ?", likeitem).Count(&total)
	trendList(c, trends, total)
}
func (h *TrendHandler) getList(c *gin.Context) {
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
	section, err := strconv.Atoi(c.Param("section"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"err": "The section format is incorrect"})
		return
	}
	var total int64
	trends := make([]TrendCount, 0)
	if section == 0 {
		DB.Preload("User.Role").Preload("Images").Order("id desc").Model(&models.Trend{}).Select("COUNT(comment.id) as comment,trend.*").Joins("left join comment on comment.trend_id=trend.id").Group("trend.id").Limit(size).Offset(size * (page - 1)).Find(&trends)
		DB.Model(&models.Trend{}).Count(&total)
	} else {
		DB.Preload("User.Role").Preload("Images").Order("id desc").Model(&models.Trend{}).Select("COUNT(comment.id) as comment,trend.*").Joins("left join comment on comment.trend_id=trend.id").Where("trend.section_id = ?", section).Group("trend.id").Limit(size).Offset(size * (page - 1)).Find(&trends)
		DB.Model(&models.Trend{}).Where("trend.section_id = ?", section).Count(&total)
	}
	trendList(c, trends, total)
}
func (h *TrendHandler) search(c *gin.Context) {
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
	val := c.Param("val")
	searchTerm := fmt.Sprintf("%%%s%%", val)
	var total int64
	trends := make([]TrendCount, 0)
	DB.Preload("User.Role").Preload("Images").Order("id desc").Model(&models.Trend{}).Select("COUNT(comment.id) as comment,trend.*").Joins("left join comment on comment.target='trend' and comment.target_id=trend.id").Where("content LIKE ?", searchTerm).Group("trend.id").Limit(size).Offset(size * (page - 1)).Find(&trends)
	DB.Model(&models.Trend{}).Where("title LIKE ? OR content LIKE ?", searchTerm, searchTerm).Count(&total)
	trendList(c, trends, total)
}
func (h *TrendHandler) addTrend(c *gin.Context) {
	var Req struct {
		Content   string   `json:"content" binding:"required"`
		SectionId int      `json:"section_id" binding:"required"`
		Images    []string `json:"images" binding:"required"`
	}
	UserId, _ := c.Get("UserId")
	if err := c.ShouldBindJSON(&Req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"err": err.Error()})
		return
	}
	trend := models.Trend{
		Content:   Req.Content,
		UserID:    UserId.(uint),
		SectionID: Req.SectionId,
	}
	DB.Create(&trend)
	for _, v := range Req.Images {
		DB.Create(&models.TrendImage{TrendID: trend.ID, Url: v})
	}
	c.JSON(http.StatusOK, gin.H{"msg": "Created successfully", "data": trend.ID})
}
func (h *TrendHandler) topComment(c *gin.Context) {
	commentId := c.Param("id")
	comment := models.Comment{}
	DB.First(&comment, commentId)
	if comment.Level == 0 {
		comment.Level = 1
	} else {
		comment.Level = 0
	}
	DB.Save(&comment)
	c.JSON(http.StatusOK, gin.H{"msg": "Updated successfully"})
}
func (h *TrendHandler) delComment(c *gin.Context) {
	commentId := c.Param("id")
	DB.Delete(&models.Comment{}, commentId)
	DB.Where("comment_id = ?", commentId).Delete(&models.CommentImage{})
	DB.Where("comment_id = ?", commentId).Delete(&models.CommentLike{})
	c.JSON(http.StatusOK, gin.H{"msg": "Deteled successfully"})
}
func (h *TrendHandler) deleteTrend(c *gin.Context) {
	trendId := c.Param("id")
	DB.Delete(&models.Trend{}, trendId)
	DB.Where("trend_id = ?", trendId).Delete(&models.TrendImage{})
	DB.Where("trend_id = ?", trendId).Delete(&models.Comment{})
	c.JSON(http.StatusOK, gin.H{"msg": "Deteled successfully"})
}

func (h *TrendHandler) getInfo(c *gin.Context) {
	trendId := c.Param("id")
	UserId, _ := c.Get("UserId")
	var Trend models.Trend
	result := DB.Preload("User.Role").Preload("Images").Preload("Like").First(&Trend, trendId).Error
	if errors.Is(result, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusNotFound, gin.H{
			"msg": "Trend does not exist",
		})
		return
	}
	Trend.Browse += 1
	DB.Save(&Trend)
	images := make([]string, 0)
	for _, image := range Trend.Images {
		images = append(images, image.Url)
	}
	has := false
	for _, like := range Trend.Like {
		if like.UserID == UserId.(uint) {
			has = true
		}
	}
	trendResp := trendResp{
		Has:       has,
		Id:        Trend.ID,
		CreatedAt: Trend.CreatedAt,
		Content:   Trend.Content,
		Browse:    Trend.Browse,
		Like:      len(Trend.Like),
		SectionId: Trend.SectionID,
		Images:    images,
		My:        UserId.(uint) == Trend.UserID,
		User: userInfo{
			Id:       Trend.UserID,
			Avatar:   Trend.User.Avatar,
			Nickname: Trend.User.Nickname,
			Role:     Trend.User.Role.Name,
		},
	}
	c.JSON(http.StatusOK, gin.H{"msg": "Successfully obtained", "data": trendResp})
}
