package handlers

import (
	. "easyjapanese/db"
	"easyjapanese/internal/middleware"
	"easyjapanese/internal/models"
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
	v2 := router.Group("/trend").Use(middleware.User())
	{
		v2.POST("", h.addTrend)
		v2.GET("/info/:id", h.getInfo)
		v2.GET("/:section/:page/:size", h.getList)
		v2.DELETE("/:id", h.deleteTrend)
	}
	v3 := router.Group("/comment").Use(middleware.User())
	{
		v3.POST("", h.addComment)
		v3.PUT("/top/:id", h.topComment)
		v3.DELETE("/:id", h.delComment)
		v3.GET("/:trendid/:sort/:page/:size", h.getCommentList)
		v3.GET("/:trendid/:sort/:page/:size/:hideid", h.getCommentList)
		//v1.POST("/getone", h.getOne)
		//v1.GET("/:target/:target_id/:page/:size/:sort/:hide_id", h.getList)
		//v1.POST("/like/:id", h.like)
		//v1.GET("/child/:parent_id/:page/:size/:sort", h.getChild)
	}
	//
	//v1.GET("/search/:page/:size/:val", h.search)
	//
	//v1.GET("/mylist/:page/:size", h.getMyList)
	//v1.POST("/like/:id", h.like)
	//v1.GET("/like/:id", h.getLike)
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
	HasLike    bool         `json:"has_like"`
	ChildCount int          `json:"child_count"`
	ParentId   *uint        `json:"parent_id"`
	Top        bool         `json:"top"`
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
				Id:       child.ID,
				Content:  child.Content,
				ParentId: child.ParentID,
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
		Content  string   `json:"content"`
		To       uint     `json:"to"`
		TrentID  uint     `json:"trend_id"`
		ParentID uint     `json:"parent_id"`
		Images   []string `json:"images"`
	}
	UserId, _ := c.Get("UserId")
	if err := c.ShouldBindJSON(&Req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"err": err.Error()})
		return
	}
	var ParentID *uint
	//content := ""
	if Req.ParentID == 0 {
		ParentID = nil
		//var trend models.Trend
		//DB.First(&trend, Req.TargetID)
		//content = trend.Content
	} else {
		ParentID = &Req.ParentID
		//var comment models.Comment
		//DB.First(&comment, ParentID)
		//content = comment.Content
	}
	comment := &models.Comment{
		Content:  Req.Content,
		To:       Req.To,
		From:     UserId.(uint),
		TrendID:  Req.TrentID,
		ParentID: ParentID,
	}
	DB.Debug().Create(&comment)
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
func (h *TrendHandler) getLike(c *gin.Context) {
	targetId, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}
	UserId, _ := c.Get("UserId")
	result := DB.Where("target_id = ? AND user_id = ?", uint(targetId), UserId).First(&models.TrendLike{}).Error
	if errors.Is(result, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusOK, gin.H{"msg": "Successfully obtained", "data": false})
		return
	} else {
		c.JSON(http.StatusOK, gin.H{"msg": "Successfully obtained", "data": true})
		return
	}
}
func (h *TrendHandler) like(c *gin.Context) {
	//targetId, err := strconv.ParseUint(c.Param("id"), 10, 32)
	//if err != nil {
	//	c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
	//	return
	//}
	//UserId, _ := c.Get("UserId")
	//var like models.TrendLike
	//result := DB.Where("target_id = ? AND user_id = ?", uint(targetId), UserId).First(&like).Error
	//var Trend models.Trend
	//DB.First(&Trend, uint(targetId))
	//if errors.Is(result, gorm.ErrRecordNotFound) {
	//	DB.Create(&models.TrendLike{
	//		TargetID: uint(targetId),
	//		UserID:   UserId.(uint),
	//	})
	//	DB.Create(&models.LikeRecord{
	//		Content:  TruncateString(Trend.Content, 100),
	//		TargetID: int(targetId),
	//		Target:   "trend",
	//		ToID:     Trend.UserID,
	//		FromID:   UserId.(uint),
	//	})
	//	Trend.Likenum = Trend.Likenum + 1
	//	DB.Save(&Trend)
	//	c.JSON(http.StatusOK, gin.H{"msg": "like"})
	//	return
	//} else {
	//	DB.Unscoped().Delete(&like)
	//	Trend.Likenum = Trend.Likenum - 1
	//	DB.Save(&Trend)
	//	DB.Unscoped().Delete(&models.LikeRecord{}, "target = ? and target_id = ?", "trend", uint(targetId))
	//	c.JSON(http.StatusOK, gin.H{"msg": "dislike"})
	//	return
	//}
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
	DB.Debug().Preload("User.Role").Preload("Images").Order("id desc").Model(&models.Trend{}).Select("COUNT(comment.id) as comment,trend.*").Joins("left join comment on comment.target='trend' and comment.target_id=trend.id").Where("user_id = ?", UserId).Group("trend.id").Limit(size).Offset(size * (page - 1)).Find(&trends)
	DB.Model(&models.Trend{}).Where("user_id = ?", UserId).Count(&total)
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
		DB.Preload("User.Role").Preload("Images").Order("id desc").Model(&models.Trend{}).Select("COUNT(comment.id) as comment,trend.*").Joins("left join comment on comment.trend_id=trend.id").Where("section_id = ?", section).Group("trend.id").Limit(size).Offset(size * (page - 1)).Find(&trends)
		DB.Model(&models.Trend{}).Where("section_id = ?", section).Count(&total)
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
	DB.Debug().Create(&trend)
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
	trendResp := trendResp{
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
