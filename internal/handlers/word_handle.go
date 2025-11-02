package handlers

import (
	. "easyjapanese/db"
	"easyjapanese/internal/middleware"
	"easyjapanese/internal/models"
	"easyjapanese/utils"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/duke-git/lancet/v2/slice"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type WordHandler struct{}

func (h *WordHandler) WordRoutes(router *gin.Engine) {
	jc := router.Group("/jc")
	{
		jc.GET("/list", h.jcList)
		jc.GET("/info", h.jcInfo)
	}
	router.POST("/review", middleware.User(), h.review)
	router.GET("/learn", middleware.User(), h.getNewWord)
	router.GET("/review", middleware.User(), h.getReviewWord)
	router.GET("/learnt", middleware.User(), h.getTodayWord)
	router.POST("/learnt", middleware.User(), h.setTodayWord)
	router.POST("/read", middleware.User(), h.addRead)
	router.GET("/read", h.getFollowRead)
	router.GET("/recommend", h.recommendWord)
	router.GET("/homeinfo", middleware.User(), h.getInfo)
}

type JapaneseDictRes struct {
	ID          uint     `json:"id"`
	Words       []string `json:"words"`
	Kana        string   `json:"kana"`
	Tone        string   `json:"tone"`
	Rome        string   `json:"rome"`
	Description string   `json:"description"`
}

func (h *WordHandler) setTodayWord(c *gin.Context) {
	var Req struct {
		WordId uint   `json:"word_id"`
		Type   string `json:"type"`
	}
	if err := c.ShouldBindJSON(&Req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"err": err.Error()})
		return
	}
	now := time.Now()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	endOfDay := startOfDay.Add(24*time.Hour - time.Nanosecond)
	UserId, _ := c.Get("UserId")
	word := models.ReviewProgress{}
	result := DB.Where("(created_at BETWEEN ? AND ?) and user_id=? and word_id = ?",
		startOfDay, endOfDay, UserId, Req.WordId).First(&word).Error

	if errors.Is(result, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusBadRequest, gin.H{"err": gorm.ErrRecordNotFound.Error()})
		return
	}
	switch Req.Type {
	case "write":
		word.Write = true
		DB.Save(&word)
	case "listen":
		word.Listen = true
		DB.Save(&word)
	default:
		c.JSON(http.StatusBadRequest, gin.H{"err": "未知的类型: " + Req.Type})
		return
	}
	c.JSON(http.StatusOK, gin.H{})
}
func (h *WordHandler) getTodayWord(c *gin.Context) {
	UserId, _ := c.Get("UserId")
	filter := c.Query("filter")
	reviewProgress := make([]models.ReviewProgress, 0)
	now := time.Now()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	endOfDay := startOfDay.Add(24*time.Hour - time.Nanosecond)
	db := DB.Where("(created_at BETWEEN ? AND ?) and user_id=?", startOfDay, endOfDay, UserId)
	switch filter {
	case "write":
		db.Where("write=?", false)
	case "listen":
		db.Where("listen=?", false)
	}
	db.Preload("Word").Limit(10).Find(&reviewProgress)
	result := make([]WordInfo, 0)
	for _, word := range reviewProgress {
		wordinfo := WordInfo{
			ID:          word.Word.ID,
			Words:       word.Word.Words,
			Tone:        word.Word.Tone,
			Rome:        word.Word.Rome,
			Browse:      word.Word.Browse,
			Detail:      word.Word.Detail,
			Kana:        word.Word.Kana,
			Description: word.Word.Description,
			UpdatedAt:   word.Word.UpdatedAt,
		}
		result = append(result, wordinfo)
	}
	c.JSON(200, gin.H{"data": result})
}
func (h *WordHandler) getReviewWord(c *gin.Context) {
	UserId, _ := c.Get("UserId")
	reviewProgress := make([]models.ReviewProgress, 0)
	now := time.Now()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	endOfDay := startOfDay.Add(24*time.Hour - time.Nanosecond)
	DB.Preload("Word").Where("(next_review_date BETWEEN ? AND ?) and user_id=?", startOfDay, endOfDay, UserId).Limit(10).Find(&reviewProgress)
	result := make([]WordInfo, 0)
	for _, word := range reviewProgress {
		wordinfo := WordInfo{
			ID:          word.Word.ID,
			Words:       word.Word.Words,
			Tone:        word.Word.Tone,
			Rome:        word.Word.Rome,
			Browse:      word.Word.Browse,
			Detail:      word.Word.Detail,
			Kana:        word.Word.Kana,
			Description: word.Word.Description,
			UpdatedAt:   word.Word.UpdatedAt,
		}
		result = append(result, wordinfo)
	}
	c.JSON(200, gin.H{"data": result})
}
func (h *WordHandler) review(c *gin.Context) {
	var Req struct {
		WordId  uint `json:"word_id"`
		Quality int  `json:"quality"`
	}
	if err := c.ShouldBindJSON(&Req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"err": err.Error()})
		return
	}

	UserId, _ := c.Get("UserId")
	reviewProgress := models.ReviewProgress{}
	result := DB.Where("user_id = ? and word_id = ?", UserId, Req.WordId).Order("id DESC").First(&reviewProgress).Error
	if errors.Is(result, gorm.ErrRecordNotFound) {
		t := time.Now()
		todayZero := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
		info := utils.FirstReview(Req.Quality, todayZero)
		updateReviewProgress(&reviewProgress, info, Req.Quality)
		reviewProgress.WordID = Req.WordId
		reviewProgress.UserID = UserId.(uint)
		DB.Create(&reviewProgress)
	} else {
		fmt.Print(reviewProgress.ID)
		info := utils.Review(Req.Quality, reviewProgress.Easiness, reviewProgress.Interval, reviewProgress.Repetitions, reviewProgress.NextReviewDate)
		updateReviewProgress(&reviewProgress, info, Req.Quality)
		reviewProgress.ID = 0
		DB.Create(&reviewProgress)
	}
	c.JSON(200, gin.H{})

}
func updateReviewProgress(rp *models.ReviewProgress, info utils.ReviewResult, quality int) {
	rp.Easiness = info.Easiness
	rp.Interval = info.Interval
	rp.Repetitions = info.Repetitions
	rp.Quality = quality
	rp.NextReviewDate = info.ReviewDateTime
}

func (h *WordHandler) jcList(c *gin.Context) {

	val := c.Query("val")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, err := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	if err != nil || page < 1 {
		page = 1
	}
	if err != nil || pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}
	var wordList = make([]models.JapaneseDict, 0)
	var total int64 = 0
	db := DB.Model(&models.JapaneseDict{}).Select("words", "tone", "rome", "kana", "id", "description")
	if val != "" {
		db = db.Where("MATCH(search_text) AGAINST(? IN NATURAL LANGUAGE MODE)", val)
	}
	db.Count(&total)
	db.Limit(pageSize).Offset((page - 1) * pageSize).Order("LENGTH(kana) asc").Find(&wordList)
	var result []JapaneseDictRes
	for _, v := range wordList {
		result = append(result, JapaneseDictRes{
			ID:          v.ID,
			Words:       v.Words,
			Kana:        v.Kana,
			Rome:        v.Rome,
			Tone:        v.Tone,
			Description: v.Description,
		})
	}
	c.JSON(http.StatusOK, gin.H{
		"data":  result,
		"total": total,
	})
}

type FollowReadRes struct {
	ID     uint      `json:"id"`
	UserID uint      `gorm:"column:user_id" json:"user_id"`
	Voice  string    `gorm:"column:voice;size:255" json:"voice"`
	Like   uint      `gorm:"column:like;default:0" json:"like"`
	WordID uint      `gorm:"column:word_id" json:"word_id"`
	User   userInfo  `gorm:"foreignKey:UserID" json:"user"`
	Time   time.Time `gorm:"column:created_at" json:"time"`
}

func (h *WordHandler) getFollowRead(c *gin.Context) {
	page, err := strconv.Atoi(c.Query("page"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"err": "The page format is incorrect"})
		return
	}
	pageSize, err := strconv.Atoi(c.Query("pageSize"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"err": "The size format is incorrect"})
		return
	}
	wordId, err := strconv.Atoi(c.Query("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"err": "The id format is incorrect"})
		return
	}
	result := make([]FollowReadRes, 0)
	var total int64 = 0
	db := DB.Debug().Model(&models.WordRead{}).Preload("User").Where("word_id = ? and status = 1", wordId)
	db.Limit(pageSize).Offset(pageSize * (page - 1)).Order("id desc").Find(&result)
	db.Count(&total)
	c.JSON(http.StatusOK, gin.H{
		"data":  result,
		"total": total,
		"msg":   "Successfully obtained",
	})
}
func (h *WordHandler) addRead(c *gin.Context) {
	var Req struct {
		Voice  string `json:"voice"`
		WordID uint   `json:"word_id"`
	}
	UserId, _ := c.Get("UserId")
	if err := c.ShouldBindJSON(&Req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"err": err.Error()})
		return
	}
	DB.Where("user_id = ? and word_id and ?", UserId.(uint), Req.WordID).Delete(&models.WordRead{})
	DB.Create(&models.WordRead{
		UserID: UserId.(uint),
		Voice:  Req.Voice,
		WordID: Req.WordID,
	})
	c.JSON(http.StatusOK, gin.H{
		"msg": "Submitted successfully",
	})
}

type BookInfo struct {
	Name     string `json:"name"`
	ID       uint   `json:"id"`
	Describe string `json:"describe"`
	Icon     string `json:"icon"`
}

func (h *WordHandler) getInfo(c *gin.Context) {
	UserId, _ := c.Get("UserId")
	learnRecords := make([]models.ReviewProgress, 0)
	DB.Order("created_at desc").Where("user_id = ?", UserId).Find(&learnRecords)
	// 判断日期是否连续
	day := 0
	now := time.Now()
	midnight := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	dayTimestamp := midnight.Unix()
	dayGroups := make([]time.Time, 0)
	var dayReview int64 = 0
	var dayLearn int64 = 0
	todayStart := time.Now().Truncate(24 * time.Hour)
	todayEnd := todayStart.Add(24*time.Hour - 1*time.Second)
	year, month, _ := now.Date()
	dates := make([]string, 0)
	for k, v := range learnRecords {
		if v.CreatedAt.Year() == year || v.CreatedAt.Month() == month {
			dates = slice.AppendIfAbsent(dates, v.CreatedAt.Format("01-02"))
		}
		if v.UpdatedAt.Before(todayEnd) && v.UpdatedAt.After(todayStart) && !v.UpdatedAt.Equal(v.CreatedAt) {
			dayReview++
		}
		if v.CreatedAt.Before(todayEnd) && v.CreatedAt.After(todayStart) {
			dayLearn++
		}
		if slice.Contain(dayGroups, v.CreatedAt) {
			dayGroups = append(dayGroups, v.CreatedAt)
			timestamp := time.Date(v.CreatedAt.Year(), v.CreatedAt.Month(), v.CreatedAt.Day(), 0, 0, 0, 0, v.CreatedAt.Location()).Unix()
			diffDays := int((dayTimestamp - timestamp) / 86400)
			// 判断是否连续
			if k == 0 && diffDays == 0 {
				day = 1
			} else if k == 0 && diffDays == 1 {
				day = 1
			} else {
				//第一个是不是今天
				onetimestamp := time.Date(dayGroups[0].Year(), dayGroups[0].Month(), dayGroups[0].Day(), 0, 0, 0, 0, dayGroups[0].Location()).Unix()
				if onetimestamp == dayTimestamp {
					if diffDays == k {
						day++
					} else {
						break
					}
				} else {
					if diffDays == k+1 {
						day++
					} else {
						break
					}
				}
			}
		}
	}
	var userConfig models.UserConfig
	DB.Where("user_id = ?", UserId).First(&userConfig)
	var bookInfo BookInfo
	DB.Model(models.WordBook{}).First(&bookInfo, userConfig.BookID)
	wordids := make([]uint, 0)
	DB.Model(models.WordBooksRelation{}).Where("book_id = ?", userConfig.BookID).Pluck("word_id", &wordids)
	learnnum := 0
	review := 0
	learn := 0
	endOfDay := midnight.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
	for _, v := range learnRecords {
		if v.NextReviewDate.Unix() < endOfDay.Unix() {
			review++
		}
		if slice.Contain(wordids, v.WordID) {
			learnnum++
		}
	}
	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			// 一共学习的单词
			"learnt": len(learnRecords),
			// 今日复习
			"day_review": dayReview,
			// 今日新学
			"day_learn": dayLearn,
			// 选择的单词书数量
			"wordnum": len(wordids),
			// 等待复习的单词
			"review": review,
			// 今日学习的单词
			"learn": learn,
			// 选择的单词书学习的单词数量
			"learnnum":  learnnum,
			"day":       day,
			"book_info": bookInfo,
			"dates":     dates,
		},
	})
}

func (h *WordHandler) getNewWord(c *gin.Context) {
	UserId, _ := c.Get("UserId")
	var config models.UserConfig
	DB.First(&config, "user_id = ?", UserId)
	wordbooks := make([]models.WordBooksRelation, 0)
	result := make([]WordInfo, 0)
	DB.Preload("Word").Joins("LEFT JOIN review_progress lp ON lp.word_id = word_books_relation.word_id").
		Where("lp.word_id IS NULL AND word_books_relation.book_id = ?", config.BookID).
		Order("word_books_relation.id DESC").
		Limit(20).
		Find(&wordbooks)
	for _, word := range wordbooks {
		wordinfo := WordInfo{
			ID:          word.Word.ID,
			Words:       word.Word.Words,
			Tone:        word.Word.Tone,
			Rome:        word.Word.Rome,
			Browse:      word.Word.Browse,
			Detail:      word.Word.Detail,
			Kana:        word.Word.Kana,
			Description: word.Word.Description,
			UpdatedAt:   word.Word.UpdatedAt,
		}
		result = append(result, wordinfo)
	}
	c.JSON(http.StatusOK, gin.H{
		"data":  result,
		"total": len(result),
	})
}

type WordInfo struct {
	ID          uint            `json:"id"`
	Words       []string        `json:"words" gorm:"serializer:json"`
	Kana        string          `json:"kana"`
	Tone        string          `json:"tone"`
	Rome        string          `json:"rome"`
	Detail      []models.Detail `json:"detail" gorm:"serializer:json"`
	UpdatedAt   time.Time       `json:"updated_at"`
	Browse      uint            `json:"browse"`
	Description string          `json:"description"`
}

func (h *WordHandler) jcInfo(c *gin.Context) {
	var wordInfo WordInfo
	id, err := strconv.Atoi(c.Query("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"err": "The id format is incorrect"})
		return
	}
	result := DB.Model(models.JapaneseDict{}).First(&wordInfo, id).Error
	if errors.Is(result, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusNotFound, gin.H{
			"err": "Word does not exist",
		})
		return
	}
	DB.Model(&models.JapaneseDict{Model: gorm.Model{ID: wordInfo.ID}}).Update("browse", wordInfo.Browse+1)
	c.JSON(http.StatusOK, gin.H{
		"data": wordInfo,
	})
}

func (h *WordHandler) recommendWord(c *gin.Context) {
	recommendWords := make([]models.JapaneseDict, 0)
	DB.Order("browse desc").Limit(10).Find(&recommendWords)
	var Result []string
	for _, v := range recommendWords {
		Result = append(Result, v.Words[0])
	}
	c.JSON(http.StatusOK, gin.H{
		"msg":  "Successfully obtained",
		"data": Result,
	})
}
