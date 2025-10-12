package handlers

import (
	. "easyjapanese/db"
	"easyjapanese/internal/middleware"
	"easyjapanese/internal/models"
	"easyjapanese/utils"
	"errors"
	"net/http"
	"strconv"
	"time"

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

	router.POST("/read", middleware.User(), h.addRead)
	router.GET("/read", h.getFollowRead)
	router.POST("/edit", middleware.User(), h.editWord)
	router.GET("/edit", h.getEditWord)
	router.GET("recommend", h.recommendWord)
	// //学习单词
	// learn := router.Group("/learn").Use(middleware.User())
	// {
	// 	learn.GET("newword", h.getNewWord)
	// 	learn.POST("record/add", h.addLearnRecord)
	// 	learn.POST("record/update", h.updateLearnRecord)
	// 	learn.POST("gettodaywords", h.getTodayWords)
	// 	learn.POST("getoptions", h.getOptions)
	// 	learn.GET("review", h.getReview)
	// 	learn.GET("info", h.getInfo)
	// }
}

type JapaneseDictRes struct {
	ID       uint     `json:"id"`
	Words    []string `json:"words"`
	Kana     string   `json:"kana"`
	Tone     string   `json:"tone"`
	Rome     string   `json:"rome"`
	Meanings string   `json:"meanings"`
}

func (h *WordHandler) getReviewWord(c *gin.Context) {

}
func (h *WordHandler) review(c *gin.Context) {
	var Req struct {
		WordId  int `json:"word_id"`
		Quality int `json:"quality"`
	}
	if err := c.ShouldBindJSON(&Req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"err": err.Error()})
		return
	}

	UserId, _ := c.Get("UserId")
	reviewProgress := models.ReviewProgress{}
	result := DB.Where("user_id = ? and word_id = ?", UserId, Req.WordId).First(&reviewProgress).Error
	if errors.Is(result, gorm.ErrRecordNotFound) {
		t := time.Now()
		todayZero := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
		info := utils.FirstReview(Req.Quality, todayZero)
		updateReviewProgress(&reviewProgress, info, Req.Quality)
		reviewProgress.WordID = Req.WordId
		reviewProgress.UserID = UserId.(uint)
		DB.Create(&reviewProgress)
	} else {
		info := utils.Review(Req.Quality, reviewProgress.Easiness, reviewProgress.Interval, reviewProgress.Repetitions, reviewProgress.NextReviewDate)
		updateReviewProgress(&reviewProgress, info, Req.Quality)
		DB.Save(&reviewProgress)
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
	pageSize, err := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	if err != nil || page < 1 {
		page = 1
	}
	if err != nil || pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}
	var wordList = make([]models.JapaneseDict, 0)
	var total int64 = 0
	db := DB.Model(&models.JapaneseDict{}).Select("words", "tone", "rome", "kana", "id", "detail")
	if val != "" {
		db = db.Where("MATCH(search_text) AGAINST(? IN NATURAL LANGUAGE MODE)", val)
	}
	db.Count(&total)
	db.Limit(pageSize).Offset((page - 1) * pageSize).Order("LENGTH(kana) asc").Find(&wordList)
	var result []JapaneseDictRes
	for _, v := range wordList {
		result = append(result, JapaneseDictRes{
			ID:       v.ID,
			Words:    v.Words,
			Kana:     v.Kana,
			Rome:     v.Rome,
			Tone:     v.Tone,
			Meanings: v.Meanings,
		})
	}
	c.JSON(http.StatusOK, gin.H{
		"data":  result,
		"total": total,
	})
}

type editWordRes struct {
	ID      uint      `json:"id"`
	Comment string    `gorm:"type:text" json:"comment"`
	UserID  uint      `gorm:"column:user_id"`
	User    userInfo  `gorm:"foreignKey:UserID" json:"user"`
	Time    time.Time `json:"time" gorm:"column:created_at"`
}

func (h *WordHandler) getEditWord(c *gin.Context) {
	wordId, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"err": "The id format is incorrect"})
		return
	}
	result := make([]editWordRes, 0)
	DB.Model(&models.WordEdit{}).Order("id desc").Preload("User").Where("word_id = ? and status = 1", wordId).Find(&result)
	c.JSON(http.StatusOK, gin.H{
		"data": result,
		"msg":  "Successfully obtained",
	})
}

func (h *WordHandler) editWord(c *gin.Context) {
	var Req struct {
		WordID  uint   `json:"word_id"`
		Detail  string `json:"detail"`
		Meaning string `json:"meaning"`
		Example string `json:"example"`
	}
	UserId, _ := c.Get("UserId")
	if err := c.ShouldBindJSON(&Req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"err": err.Error()})
		return
	}
	DB.Create(&models.WordEdit{
		UserID:  UserId.(uint),
		Detail:  Req.Detail,
		Meaning: Req.Meaning,
		Example: Req.Example,
		WordID:  Req.WordID,
	})
	c.JSON(http.StatusOK, gin.H{
		"msg": "Submitted successfully",
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

// func (h *WordHandler) updateLearnRecord(c *gin.Context) {
// 	UserId, _ := c.Get("UserId")
// 	var config models.UserConfig
// 	DB.Where("user_id = ?", UserId).First(&config)
// 	type wordStruct struct {
// 		Error  int  ` json:"error"`
// 		WordID uint ` json:"word_id"`
// 	}
// 	var Req struct {
// 		Words []wordStruct `json:"words"`
// 	}
// 	if err := c.ShouldBindJSON(&Req); err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"err": err.Error()})
// 		return
// 	}
// 	ids := make([]uint, 0)
// 	for _, word := range Req.Words {
// 		ids = append(ids, word.WordID)
// 	}
// 	now := time.Now()
// 	midnight := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
// 	dayTimestamp := midnight.Unix()
// 	records := make([]models.LearnRecord, 0)
// 	DB.Debug().Preload("Word").Where("word_id IN ? AND user_id = ?", ids, UserId).Find(&records)
// 	base := 86400
// 	for _, record := range records {
// 		result, _ := slice.FindBy(Req.Words, func(i int, f wordStruct) bool { return f.WordID == record.WordID })
// 		errorCount := result.Error
// 		reviewCount := record.ReviewCount
// 		if errorCount == 0 {
// 			reviewTime := config.CycleConfig.Cycle[reviewCount] * base
// 			record.ReviewCount = 1 + reviewCount
// 			record.ReviewTime = dayTimestamp + int64(reviewTime)
// 			log.Println(reviewTime / base)
// 		} else {
// 			reviewTime := 0
// 			if errorCount >= 6 {
// 				//forgotten
// 				reviewTime = int(math.Ceil((1-float64(config.CycleConfig.Extent.Forgotten)*0.01)*float64(config.CycleConfig.Cycle[reviewCount]))) * base
// 			} else if errorCount >= 3 {
// 				//vague
// 				reviewTime = int(math.Ceil((1-float64(config.CycleConfig.Extent.Vague)*0.01)*float64(config.CycleConfig.Cycle[reviewCount]))) * base
// 			} else {
// 				//partial
// 				reviewTime = int(math.Ceil((1-float64(config.CycleConfig.Extent.Partial)*0.01)*float64(config.CycleConfig.Cycle[reviewCount]))) * base
// 			}
// 			if reviewTime == 0 {
// 				reviewTime = 1 * base
// 			}
// 			record.ReviewTime = dayTimestamp + int64(reviewTime)
// 		}
// 		if reviewCount > config.CycleConfig.Cycle[len(config.CycleConfig.Cycle)-1] {
// 			record.Done = true
// 		}
// 		DB.Save(&record)
// 	}

// 	c.JSON(http.StatusOK, gin.H{"msg": "Record successful"})
// }

// type BookInfo struct {
// 	Name     string      `json:"name"`
// 	ID       uint        `json:"id"`
// 	Describe string      `json:"describe"`
// 	Icon     models.Icon `json:"icon" gorm:"serializer:json"`
// }

// func (h *WordHandler) getInfo(c *gin.Context) {
// 	UserId, _ := c.Get("UserId")
// 	learnRecords := make([]models.LearnRecord, 0)
// 	DB.Order("created_at desc").Where("user_id = ?", UserId).Find(&learnRecords)
// 	// 判断日期是否连续
// 	day := 0
// 	now := time.Now()
// 	midnight := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
// 	dayTimestamp := midnight.Unix()
// 	dayGroups := make([]time.Time, 0)
// 	var dayReview int64 = 0
// 	var dayLearn int64 = 0
// 	todayStart := time.Now().Truncate(24 * time.Hour)
// 	todayEnd := todayStart.Add(24*time.Hour - 1*time.Second)
// 	year, month, _ := now.Date()
// 	dates := make([]string, 0)
// 	for k, v := range learnRecords {
// 		if v.CreatedAt.Year() == year || v.CreatedAt.Month() == month {
// 			dates = slice.AppendIfAbsent(dates, v.CreatedAt.Format("01-02"))
// 		}
// 		if v.UpdatedAt.Before(todayEnd) && v.UpdatedAt.After(todayStart) && !v.UpdatedAt.Equal(v.CreatedAt) {
// 			dayReview++
// 		}
// 		if v.CreatedAt.Before(todayEnd) && v.CreatedAt.After(todayStart) {
// 			dayLearn++
// 		}
// 		if slice.Contain(dayGroups, v.CreatedAt) {
// 			dayGroups = append(dayGroups, v.CreatedAt)
// 			timestamp := time.Date(v.CreatedAt.Year(), v.CreatedAt.Month(), v.CreatedAt.Day(), 0, 0, 0, 0, v.CreatedAt.Location()).Unix()
// 			diffDays := int((dayTimestamp - timestamp) / 86400)
// 			// 判断是否连续
// 			if k == 0 && diffDays == 0 {
// 				day = 1
// 			} else if k == 0 && diffDays == 1 {
// 				day = 1
// 			} else {
// 				//第一个是不是今天
// 				onetimestamp := time.Date(dayGroups[0].Year(), dayGroups[0].Month(), dayGroups[0].Day(), 0, 0, 0, 0, dayGroups[0].Location()).Unix()
// 				if onetimestamp == dayTimestamp {
// 					if diffDays == k {
// 						day++
// 					} else {
// 						break
// 					}
// 				} else {
// 					if diffDays == k+1 {
// 						day++
// 					} else {
// 						break
// 					}
// 				}
// 			}
// 		}
// 	}
// 	var userConfig models.UserConfig
// 	DB.Where("user_id = ?", UserId).First(&userConfig)
// 	var bookInfo BookInfo
// 	DB.Model(models.WordBook{}).First(&bookInfo, userConfig.BookID)
// 	wordids := make([]uint, 0)
// 	DB.Model(models.WordBookRelation{}).Where("book_id = ?", userConfig.BookID).Pluck("word_id", &wordids)
// 	learnnum := 0
// 	review := 0
// 	learn := 0
// 	endOfDay := midnight.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
// 	for _, v := range learnRecords {
// 		if v.ReviewTime < endOfDay.Unix() {
// 			review++
// 		}
// 		if slice.Contain(wordids, v.WordID) {
// 			learnnum++
// 		}
// 	}
// 	c.JSON(http.StatusOK, gin.H{
// 		"data": gin.H{
// 			// 一共学习的单词
// 			"learnt": len(learnRecords),
// 			// 今日复习
// 			"day_review": dayReview,
// 			// 今日新学
// 			"day_learn": dayLearn,
// 			// 选择的单词书数量
// 			"wordnum": len(wordids),
// 			// 等待复习的单词
// 			"review": review,
// 			// 今日学习的单词
// 			"learn": learn,
// 			// 选择的单词书学习的单词数量
// 			"learnnum":  learnnum,
// 			"day":       day,
// 			"book_info": bookInfo,
// 			"dates":     dates,
// 		},
// 	})
// }

// func (h *WordHandler) getReview(c *gin.Context) {
// 	UserId, _ := c.Get("UserId")
// 	var config models.UserConfig
// 	DB.First(&config, "user_id = ?", UserId)
// 	now := time.Now()
// 	midnight := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
// 	endOfDay := midnight.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
// 	endOfDayTimestamp := endOfDay.Unix()
// 	words := make([]models.LearnRecord, 0)
// 	result := make([]WordInfo, 0)
// 	DB.Debug().Order("id desc").Preload("Word.Meaning").Preload("Word.Example").Where("review_time < ?", endOfDayTimestamp).Limit(config.LearnGroup).Find(&words)
// 	for _, word := range words {
// 		wordinfo := WordInfo{
// 			ID:       word.Word.ID,
// 			Word:     word.Word.Word,
// 			Tone:     word.Word.Tone,
// 			Rome:     word.Word.Rome,
// 			Browse:   word.Word.Browse,
// 			Voice:    word.Word.Voice,
// 			Kana:     word.Word.Kana,
// 			Wordtype: word.Word.Wordtype,
// 			Detail:   word.Word.Detail,
// 			Meaning:  word.Word.Meaning,
// 			Example:  word.Word.Example,
// 		}
// 		result = append(result, wordinfo)
// 	}
// 	c.JSON(http.StatusOK, gin.H{
// 		"data":  result,
// 		"total": len(result),
// 		"msg":   "Successfully obtained",
// 	})
// }
// func (h *WordHandler) getOptions(c *gin.Context) {
// 	var Req struct {
// 		Filter []uint `json:"filter"`
// 		Limit  int    `json:"limit"`
// 	}
// 	if err := c.ShouldBindJSON(&Req); err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"err": err.Error()})
// 		return
// 	}
// 	words := make([]models.Jcdict, 0)
// 	query := DB.Order("RAND()").Preload("Meaning")
// 	if len(Req.Filter) > 0 {
// 		query.Where("id not in ?", Req.Filter).Limit(Req.Limit).Find(&words)
// 	} else {
// 		query.Limit(Req.Limit).Find(&words)
// 	}
// 	result := make([]WordInfo, 0)
// 	for _, word := range words {
// 		wordinfo := WordInfo{
// 			ID:       word.ID,
// 			Word:     word.Word,
// 			Tone:     word.Tone,
// 			Rome:     word.Rome,
// 			Browse:   word.Browse,
// 			Voice:    word.Voice,
// 			Kana:     word.Kana,
// 			Wordtype: word.Wordtype,
// 			Detail:   word.Detail,
// 			Meaning:  word.Meaning,
// 			Example:  word.Example,
// 		}
// 		result = append(result, wordinfo)
// 	}
// 	c.JSON(http.StatusOK, gin.H{
// 		"data": result,
// 		"msg":  "Successfully obtained",
// 	})
// }
// func (h *WordHandler) getTodayWords(c *gin.Context) {
// 	var Req struct {
// 		Filter []uint `json:"filter"`
// 		Type   string
// 	}
// 	var limit = 0
// 	if err := c.ShouldBindJSON(&Req); err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"err": err.Error()})
// 		return
// 	}
// 	UserId, _ := c.Get("UserId")
// 	var config models.UserConfig
// 	DB.First(&config, "user_id = ?", UserId)
// 	if Req.Type == "sound" {
// 		limit = config.SoundGroup
// 	} else if Req.Type == "write" {
// 		limit = config.WriteGroup
// 	}
// 	words := make([]models.LearnRecord, 0)
// 	result := make([]WordInfo, 0)
// 	todayStart := time.Now().Truncate(24 * time.Hour)
// 	todayEnd := todayStart.Add(24*time.Hour - 1*time.Second)
// 	query := DB.Debug().Order("id desc").Preload("Word.Meaning").Preload("Word.Example").
// 		Where("created_at BETWEEN ? AND ?", todayStart.Format("2006-01-02 15:04:05"), todayEnd.Format("2006-01-02 15:04:05"))
// 	if len(Req.Filter) > 0 {
// 		query = query.Where("word_id not in ?", Req.Filter)
// 	}
// 	if limit == 0 {
// 		query.Find(&words)
// 	} else {
// 		query.Limit(config.WriteGroup).Find(&words)
// 	}
// 	for _, word := range words {
// 		wordinfo := WordInfo{
// 			ID:       word.Word.ID,
// 			Word:     word.Word.Word,
// 			Tone:     word.Word.Tone,
// 			Rome:     word.Word.Rome,
// 			Browse:   word.Word.Browse,
// 			Voice:    word.Word.Voice,
// 			Kana:     word.Word.Kana,
// 			Wordtype: word.Word.Wordtype,
// 			Detail:   word.Word.Detail,
// 			Meaning:  word.Word.Meaning,
// 			Example:  word.Word.Example,
// 		}
// 		result = append(result, wordinfo)
// 	}
// 	c.JSON(http.StatusOK, gin.H{
// 		"data": result,
// 		"msg":  "Successfully obtained",
// 	})
// }

func (h *WordHandler) getNewWord(c *gin.Context) {
	UserId, _ := c.Get("UserId")
	var config models.UserConfig
	DB.First(&config, "user_id = ?", UserId)
	wordbooks := make([]models.WordBooksRelation, 0)
	result := make([]WordInfo, 0)
	DB.Preload("Word").Joins("LEFT JOIN review_progress lp ON lp.word_id = word_books_relation.word_id").
		Where("lp.word_id IS NULL AND word_books_relation.book_id = ?", config.BookID).
		Order("word_books_relation.id DESC").
		Limit(config.LearnGroup).
		Find(&wordbooks)
	for _, word := range wordbooks {
		wordinfo := WordInfo{
			ID:        word.Word.ID,
			Words:     word.Word.Words,
			Tone:      word.Word.Tone,
			Rome:      word.Word.Rome,
			Browse:    word.Word.Browse,
			Detail:    word.Word.Detail,
			Kana:      word.Word.Kana,
			UpdatedAt: word.Word.UpdatedAt,
		}
		result = append(result, wordinfo)
	}
	c.JSON(http.StatusOK, gin.H{
		"data":  result,
		"total": len(result),
		"msg":   "Successfully obtained",
	})
}

type WordInfo struct {
	ID        uint            `json:"id"`
	Words     []string        `json:"words" gorm:"serializer:json"`
	Kana      string          `json:"kana"`
	Tone      string          `json:"tone"`
	Rome      string          `json:"rome"`
	Detail    []models.Detail `json:"icon" gorm:"serializer:json"`
	UpdatedAt time.Time       `json:"updated_at"`
	Browse    uint            `json:"browse"`
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
			"msg": "Word does not exist",
		})
		return
	}
	DB.Model(&models.JapaneseDict{Model: gorm.Model{ID: wordInfo.ID}}).Update("browse", wordInfo.Browse+1)
	c.JSON(http.StatusOK, gin.H{
		"msg":  "Successfully obtained",
		"data": wordInfo,
	})
}

// // 新增学习记录
//
//	func (h *WordHandler) addLearnRecord(c *gin.Context) {
//		var learnRecord []models.LearnRecord
//		var Req struct {
//			Words []struct {
//				Error  int  ` json:"error"`
//				WordID uint ` json:"word_id"`
//			} `json:"words"`
//		}
//		if err := c.ShouldBindJSON(&Req); err != nil {
//			c.JSON(http.StatusBadRequest, gin.H{"err": err.Error()})
//			return
//		}
//		timestamp := time.Now().Unix() + 86400
//		UserId, _ := c.Get("UserId")
//		for _, item := range Req.Words {
//			var reviewTime int64 = 0
//			if item.Error != 0 {
//				reviewTime = timestamp * 3
//			} else {
//				reviewTime = timestamp
//			}
//			learnRecord = append(learnRecord, models.LearnRecord{
//				WordID:     item.WordID,
//				UserID:     UserId.(uint),
//				ReviewTime: reviewTime,
//			})
//		}
//		DB.Create(&learnRecord)
//		c.JSON(http.StatusOK, gin.H{
//			"msg": "Record successful",
//		})
//	}
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
