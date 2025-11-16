package handlers

import (
	. "easyjapanese/db"
	"easyjapanese/internal/middleware"
	"easyjapanese/internal/models"
	"easyjapanese/utils"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"sort"
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
	edit := router.Group("/edit").Use(middleware.User())
	{
		edit.POST("/submit", h.submitEdit)
		edit.GET("/history", h.getEditHistory)
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
	router.GET("/listen/options", middleware.User(), h.getListenOptions)
}
func (h WordHandler) submitEdit(c *gin.Context) {
	UserId, _ := c.Get("UserId")
	var Req struct {
		WordID      uint                 `json:"word_id"`
		Words       string               `json:"words"`
		Kana        string               `json:"kana"`
		Tone        string               `json:"tone"`
		Description string               `json:"description"`
		Meanings    []models.EditMeaning `json:"meanings"`
		Examples    []models.EditExample `json:"examples"`
	}
	if err := c.ShouldBindJSON(&Req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"err": err.Error()})
		return
	}
	DB.Delete(&models.WordEdit{}, "user_id = ? AND word_id = ?", UserId, Req.WordID)
	editData := models.WordEdit{}
	editData.Words = Req.Words
	editData.WordID = Req.WordID
	editData.Kana = Req.Kana
	editData.Tone = Req.Tone
	editData.Description = Req.Description
	editData.Meanings = Req.Meanings
	editData.Examples = Req.Examples
	editData.UserID = UserId.(uint)
	DB.Create(&editData)
	c.JSON(http.StatusOK, gin.H{})
}

type WordEditHistory struct {
	ID        uint      `json:"id"`
	UserID    uint      `json:"user_id"`
	Status    int8      `gorm:"type:tinyint;default:0" json:"status"`
	Comment   string    `gorm:"type:text" json:"comment"`
	CreatedAt time.Time `json:"created_at"`
	User      userInfo  `gorm:"foreignKey:UserID" json:"user"`
}

func (h WordHandler) getEditHistory(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, err := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	if err != nil || page < 1 {
		page = 1
	}
	if err != nil || pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}
	editHistorys := make([]WordEditHistory, 0)
	var total int64 = 0
	db := DB.Model(&models.WordEdit{}).Where("word_id = ?", c.Query("word_id"))
	db.Count(&total)
	db.Preload("User").Limit(pageSize).Offset((page - 1) * pageSize).Order("id desc").Find(&editHistorys)
	c.JSON(http.StatusOK, gin.H{
		"data":  editHistorys,
		"total": total,
	})
}

type ListenAnwser struct {
	Anwser   bool     `json:"anwser"`
	WordInfo WordInfo `json:"word"`
	Class    string   `json:"class"`
}

func (h *WordHandler) getListenOptions(c *gin.Context) {
	wordId := c.Query("wordId")
	var others []models.JapaneseDict
	err := DB.Where("id >= FLOOR(RAND() * (SELECT MAX(id) FROM japanese_dict))").
		Where("id NOT IN (?)", wordId).
		Limit(3).
		Find(&others).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"err": err.Error()})
		return
	}
	var word models.JapaneseDict
	DB.First(&word, wordId)
	results := make([]ListenAnwser, 0)
	for _, v := range others {
		results = append(results, ListenAnwser{
			WordInfo: WordInfo{
				ID:          v.ID,
				Words:       v.Words,
				Tone:        v.Tone,
				Rome:        v.Rome,
				Kana:        v.Kana,
				Description: v.Description,
				Voice:       fmt.Sprintf("https://jpx2ink.oss-cn-shanghai.aliyuncs.com/audio/dict/jc/%d/word.wav", v.ID),
			},
			Anwser: false,
		})
	}
	results = append(results, ListenAnwser{
		WordInfo: WordInfo{
			ID:          word.ID,
			Words:       word.Words,
			Tone:        word.Tone,
			Rome:        word.Rome,
			Kana:        word.Kana,
			Description: word.Description,
			Voice:       fmt.Sprintf("https://jpx2ink.oss-cn-shanghai.aliyuncs.com/audio/dict/jc/%d/word.wav", word.ID),
		},
		Anwser: true,
	})
	rand.Shuffle(len(results), func(i, j int) {
		results[i], results[j] = results[j], results[i]
	})
	c.JSON(http.StatusOK, gin.H{
		"data": results,
	})
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

type TodayWords struct {
	WordInfo
	Write  bool `json:"write"`
	Listen bool `json:"listen"`
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
		db.Where("`write`=?", false)
	case "listen":
		db.Where("listen=?", false)
	}
	db.Preload("Word").Limit(10).Find(&reviewProgress)
	result := make([]TodayWords, 0)
	for _, word := range reviewProgress {
		wordinfo := TodayWords{
			WordInfo: WordInfo{
				ID:          word.Word.ID,
				Words:       word.Word.Words,
				Tone:        word.Word.Tone,
				Rome:        word.Word.Rome,
				Browse:      word.Word.Browse,
				Detail:      word.Word.Detail,
				Kana:        word.Word.Kana,
				Description: word.Word.Description,
				UpdatedAt:   word.Word.UpdatedAt,
				Voice:       fmt.Sprintf("https://jpx2ink.oss-cn-shanghai.aliyuncs.com/audio/dict/jc/%d/word.wav", word.WordID),
			},
			Write:  word.Write,
			Listen: word.Listen,
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
	startOfDay = startOfDay.Truncate(time.Second)
	endOfDay := startOfDay.Add(24*time.Hour - time.Second)
	startOfDayStr := startOfDay.Format("2006-01-02 15:04:05")
	endOfDayStr := endOfDay.Format("2006-01-02 15:04:05")
	subQuery := DB.
		Table("review_progress").
		Select("word_id, MAX(next_review_date) as max_next_review_date").
		Where("user_id = ?", UserId).
		Group("word_id")
	DB.Debug().
		Preload("Word").
		Joins("JOIN (?) AS latest ON review_progress.word_id = latest.word_id AND review_progress.next_review_date = latest.max_next_review_date", subQuery).
		Where("review_progress.next_review_date BETWEEN ? AND ?", startOfDayStr, endOfDayStr).
		Limit(10).
		Find(&reviewProgress)
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
			Voice:       fmt.Sprintf("https://jpx2ink.oss-cn-shanghai.aliyuncs.com/audio/dict/jc/%d/word.wav", word.WordID),
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
		reviewProgress.Type = "learn"
		DB.Create(&reviewProgress)
	} else {
		fmt.Print(reviewProgress.ID)
		info := utils.Review(Req.Quality, reviewProgress.Easiness, reviewProgress.Interval, reviewProgress.Repetitions, reviewProgress.NextReviewDate)
		updateReviewProgress(&reviewProgress, info, Req.Quality)
		reviewProgress.ID = 0
		reviewProgress.Type = "review"
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

	// 查询该用户的所有学习/复习记录
	var learnRecords []models.ReviewProgress
	DB.Order("created_at desc").Where("user_id = ?", UserId).Find(&learnRecords)

	now := time.Now()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	todayEnd := todayStart.Add(24*time.Hour - time.Second)

	var (
		dayReview int64
		dayLearn  int64
	)

	// -------- 统计今日学习 & 复习数量，同时收集日期 --------
	dateSet := make(map[string]bool)
	for _, v := range learnRecords {
		if v.Type == "learn" && v.CreatedAt.After(todayStart) && v.CreatedAt.Before(todayEnd) {
			dayLearn++
		}
		if v.Type == "review" && v.UpdatedAt.After(todayStart) && v.UpdatedAt.Before(todayEnd) {
			dayReview++
		}

		// 收集所有学习日期
		if v.Type == "learn" || v.Type == "review" {
			dayStr := v.CreatedAt.Format("2006-01-02")
			dateSet[dayStr] = true
		}
	}

	// -------- 计算连续学习天数 --------
	day := 0
	if len(dateSet) > 0 {
		dates := make([]time.Time, 0, len(dateSet))
		for d := range dateSet {
			t, _ := time.ParseInLocation("2006-01-02", d, time.Local)
			dates = append(dates, t)
		}
		sort.Slice(dates, func(i, j int) bool { return dates[i].Before(dates[j]) })

		today := todayStart
		for i := len(dates) - 1; i >= 0; i-- {
			diffDays := int(today.Sub(dates[i]).Hours() / 24)
			if diffDays == day {
				day++
			} else {
				break // 不连续
			}
		}
	}

	// -------- 获取用户配置及单词书信息 --------
	var userConfig models.UserConfig
	DB.Where("user_id = ?", UserId).First(&userConfig)

	var bookInfo BookInfo
	DB.Model(models.WordBook{}).First(&bookInfo, userConfig.BookID)

	// 获取当前书的全部单词ID
	var wordIDs []uint
	DB.Model(models.WordBooksRelation{}).Where("book_id = ?", userConfig.BookID).Pluck("word_id", &wordIDs)

	// -------- 统计学习单词数量与等待复习数量 --------
	learnNum := 0
	review := 0
	latestMap := make(map[uint]models.ReviewProgress)

	for _, rp := range learnRecords {
		// 取每个单词最新的复习记录
		if latest, ok := latestMap[rp.WordID]; !ok || rp.NextReviewDate.After(latest.NextReviewDate) {
			latestMap[rp.WordID] = rp
		}
		if slice.Contain(wordIDs, rp.WordID) {
			learnNum++
		}
	}

	for _, rp := range latestMap {
		if !rp.NextReviewDate.Before(todayStart) && rp.NextReviewDate.Before(todayEnd) {
			review++
		}
	}

	// -------- 构造日期数据（本月内的日期） --------
	year, month, _ := now.Date()
	datesForChart := make([]string, 0)
	for d := range dateSet {
		t, _ := time.ParseInLocation("2006-01-02", d, time.Local)
		if t.Year() == year && t.Month() == month {
			datesForChart = append(datesForChart, t.Format("01-02"))
		}
	}

	// -------- 返回结果 --------
	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"learnt":     len(learnRecords), // 一共学习的单词
			"day_review": dayReview,         // 今日复习
			"day_learn":  dayLearn,          // 今日新学
			"wordnum":    len(wordIDs),      // 当前单词书总数
			"review":     review,            // 等待复习数量
			"learnnum":   learnNum,          // 当前书中学习过的数量
			"day":        day,               // 连续学习天数
			"book_info":  bookInfo,
			"dates":      datesForChart, // 本月学习日期
		},
	})
}

func (h *WordHandler) getNewWord(c *gin.Context) {
	UserId, _ := c.Get("UserId")
	var config models.UserConfig
	DB.First(&config, "user_id = ?", UserId)
	wordbooks := make([]models.WordBooksRelation, 0)
	result := make([]WordInfo, 0)
	DB.Debug().Preload("Word").Joins("LEFT JOIN review_progress lp ON lp.word_id = word_books_relation.word_id").
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
			Voice:       fmt.Sprintf("https://jpx2ink.oss-cn-shanghai.aliyuncs.com/audio/dict/jc/%d/word.wav", word.WordID),
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
	Voice       string          `json:"voice"`
}

func (h *WordHandler) jcInfo(c *gin.Context) {
	var dict models.JapaneseDict
	id, err := strconv.Atoi(c.Query("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"err": "The id format is incorrect"})
		return
	}
	if err := DB.First(&dict, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"err": "Word does not exist"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"err": err.Error()})
		return
	}
	wordInfo := WordInfo{
		ID:          dict.ID,
		Words:       dict.Words,
		Kana:        dict.Kana,
		Tone:        dict.Tone,
		Rome:        dict.Rome,
		Detail:      dict.Detail,
		Browse:      dict.Browse,
		Description: dict.Description,
		Voice:       fmt.Sprintf("https://jpx2ink.oss-cn-shanghai.aliyuncs.com/audio/dict/jc/%d/word.wav", dict.ID),
		UpdatedAt:   dict.UpdatedAt,
	}
	DB.Model(&dict).Update("browse", dict.Browse+1)
	c.JSON(http.StatusOK, gin.H{"data": wordInfo})

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
