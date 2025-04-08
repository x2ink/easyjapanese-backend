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
	"regexp"
	"strconv"
	"time"
)

type WordHandler struct{}
type wordBookRes struct {
	Name     string `json:"name"`
	Category string `json:"category"`
	Icon     string `json:"icon"`
	Words    int    `json:"words"`
	ID       uint   `json:"id"`
	Describe string `json:"describe"`
}

func (h *WordHandler) WordRoutes(router *gin.Engine) {
	router.POST("/followread", middleware.User(), h.followRead)
	router.GET("/followread/:id/:page/:size", h.getFollowRead)
	router.POST("/followread/:type", middleware.User(), h.likeFollowRead)
	router.POST("/editword", middleware.User(), h.editWord)
	router.GET("/editword/:id", h.getEditWord)
	jc := router.Group("/jc")
	{
		jc.GET("/search/:page/:size/:val", h.jcSearch)
		jc.GET("/list/:page/:size", h.jcList)
		jc.GET("/info/:id", h.jcInfo)
		jc.PUT("/:id", h.setInfo)
	}
	//推荐单词
	router.GET("recommend", h.recommendWord)

	router.GET("/wordbook", h.getWordBook)
	router.GET("/wordbook/:id/:page/:size", h.getWordBookList)
	router.GET("/todayword", middleware.User(), h.getNewWord)

	cj := router.Group("/cj")
	{
		cj.GET("/search/:page/:size/:val", h.cjSearch)
		cj.GET("/info/:id", h.cjInfo)
	}
	learn := router.Group("/learn").Use(middleware.User())
	{
		learn.POST("record/add", h.addLearnRecord)
		learn.POST("record/update", h.updateLearnRecord)
		learn.GET("writefrommemory", h.getWriteFromMemory)
		learn.GET("review", h.getReview)
		learn.GET("info", h.getInfo)
	}

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
func (h *WordHandler) setInfo(c *gin.Context) {
	var Req struct {
		Rome     string   `json:"rome"`
		Kana     string   `json:"kana"`
		WordType string   `json:"wordtype"`
		Meaning  []string `json:"meanings"`
	}
	if err := c.ShouldBindJSON(&Req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"err": err.Error()})
		return
	}
	wordId, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"err": "The size format is incorrect"})
		return
	}
	if wordId == 0 {
		c.JSON(http.StatusOK, gin.H{
			"msg": "Submitted successfully",
		})
		return
	}
	var word models.Jcdict
	DB.First(&word, wordId)
	word.Rome = Req.Rome
	word.Kana = Req.Kana
	word.Wordtype = Req.WordType
	DB.Save(&word)
	DB.Where("word_id=?", wordId).Delete(&models.JcdictMeaning{})
	means := make([]models.JcdictMeaning, 0)
	for _, v := range Req.Meaning {
		means = append(means, models.JcdictMeaning{
			WordID:  uint(wordId),
			Meaning: v,
		})
	}
	DB.Create(&means)
	c.JSON(http.StatusOK, gin.H{
		"msg": "Submitted successfully",
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

func (h *WordHandler) likeFollowRead(c *gin.Context) {
	var Req struct {
		ID uint `json:"id"`
	}
	UserId, _ := c.Get("UserId")
	if err := c.ShouldBindJSON(&Req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"err": err.Error()})
		return
	}
	var Res models.WordRead
	Res.ID = Req.ID
	result := DB.Where("user_id = ?", UserId.(uint)).First(&Res).Error
	if errors.Is(result, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"err": result})
		return
	} else {
		if c.Param("type") == "like" {
			Res.Like += 1
		} else {
			Res.Like -= 1
		}
		DB.Save(&Res)
		c.JSON(http.StatusOK, gin.H{"msg": "Liked successfully"})
		return
	}
}
func (h *WordHandler) getFollowRead(c *gin.Context) {
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
	wordId, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"err": "The id format is incorrect"})
		return
	}
	result := make([]FollowReadRes, 0)
	var total int64 = 0
	DB.Model(&models.WordRead{}).Order("id desc").Preload("User").Where("word_id = ?", wordId).Limit(size).Offset(size * (page - 1)).Find(&result)
	DB.Model(&models.WordRead{}).Where("word_id = ?", wordId).Count(&total)
	c.JSON(http.StatusOK, gin.H{
		"data":  result,
		"total": total,
		"msg":   "Successfully obtained",
	})
}
func (h *WordHandler) followRead(c *gin.Context) {
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
func getErrorCount(words []struct {
	Id         uint `json:"id"`
	ErrorCount int  `json:"error_count"`
}, id uint) int {
	for _, word := range words {
		if word.Id == id {
			return word.ErrorCount
		}
	}
	return 0
}
func (h *WordHandler) updateLearnRecord(c *gin.Context) {
	UserId, _ := c.Get("UserId")
	var Req struct {
		Words []struct {
			Id         uint `json:"id"`
			ErrorCount int  `json:"error_count"`
		}
	}
	cycle := []int{1, 3, 7, 14, 30}
	if err := c.ShouldBindJSON(&Req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"err": err.Error()})
		return
	}
	ids := []uint{}
	for _, word := range Req.Words {
		ids = append(ids, word.Id)
	}
	now := time.Now()
	midnight := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	dayTimestamp := midnight.Unix()
	records := []models.LearnRecord{}
	DB.Preload("Word").Where("word_id IN ? AND user_id = ?", ids, UserId).Find(&records)
	base := 86400
	for _, record := range records {
		errorCount := getErrorCount(Req.Words, record.WordID)
		reviewCount := record.ReviewCount
		if errorCount == 0 {
			reviewTime := cycle[reviewCount] * base
			record.ReviewCount = 1 + reviewCount
			record.ReviewTime = dayTimestamp + int64(reviewTime)
		} else {
			reviewTime := 0
			if errorCount >= 6 {
				reviewTime = 1 * base
			} else if errorCount >= 3 {
				reviewTime = 2 * base
			} else {
				reviewTime = 3 * base
			}
			record.ReviewTime = dayTimestamp + int64(reviewTime)
		}
		DB.Save(&record)
	}

	c.JSON(http.StatusOK, gin.H{"msg": "Record successful"})
}
func contains[T comparable](slice []T, target T) bool {
	for _, v := range slice {
		if v == target {
			return true
		}
	}
	return false
}
func (h *WordHandler) getInfo(c *gin.Context) {
	UserId, _ := c.Get("UserId")
	learnRecords := make([]models.LearnRecord, 0)
	DB.Order("created_at desc").Where("user_id = ?", UserId).Find(&learnRecords)
	day := 0
	now := time.Now()
	midnight := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	dayTimestamp := midnight.Unix()
	for k, v := range learnRecords {
		timestamp := time.Date(v.CreatedAt.Year(), v.CreatedAt.Month(), v.CreatedAt.Day(), 0, 0, 0, 0, v.CreatedAt.Location()).Unix()
		diffDays := int((dayTimestamp - timestamp) / 86400)
		// 判断是否连续
		if k == 0 && diffDays == 0 {
			day = 1
		} else if diffDays == k+1 {
			day++
		} else {
			break
		}
	}
	var userConfig models.UserConfig
	DB.Preload("Book").Where("user_id = ?", UserId).First(&userConfig)
	wordids := make([]uint, 0)
	DB.Model(models.WordBookRelation{}).Where("book_id = ?", userConfig.BookID).Pluck("word_id", &wordids)
	learnnum := 0
	review := 0
	learn := 0
	endOfDay := midnight.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
	todayStart := time.Now().Truncate(24 * time.Hour)
	todayEnd := todayStart.Add(24*time.Hour - 1*time.Second)
	for _, v := range learnRecords {
		if v.ReviewTime < endOfDay.Unix() {
			review++
		}
		if v.CreatedAt.After(todayStart) && v.CreatedAt.Before(todayEnd) {
			learn++
		}
		if contains(wordids, v.WordID) {
			learnnum++
		}
	}
	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"bookname": userConfig.Book.Name,
			"review":   review,
			"learn":    learn,
			"wordnum":  len(wordids),
			"learnnum": learnnum,
			"day":      day,
		},
	})
}

type writeFromMemory struct {
	Word       string `json:"word"`
	Meaning    string `json:"meaning"`
	Kana       string `json:"kana"`
	Tone       string `json:"tone"`
	ID         uint   `json:"id"`
	ErrorCount uint   `json:"error_count"`
	Rome       string `json:"rome"`
	Voice      string `json:"voice"`
	Done       bool   `json:"done"`
}
type reviewRes struct {
	Exercise      bool            `json:"exercise"`
	Done          bool            `json:"done"`
	Tone          string          `json:"tone"`
	ErrorCount    int             `json:"error_count"`
	Progress      []bool          `json:"progress"`
	Meaning       []string        `json:"meaning"`
	Word          string          `json:"word"`
	Kana          string          `json:"kana"`
	ID            uint            `json:"id"`
	Rome          string          `json:"rome"`
	Voice         string          `json:"voice"`
	Detail        []models.Detail `json:"detail" gorm:"serializer:json"`
	MeaningOption []option        `json:"meaning_option"`
}

func (h *WordHandler) getReview(c *gin.Context) {
	//now := time.Now()
	//midnight := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	//endOfDay := midnight.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
	//endOfDayTimestamp := endOfDay.Unix()
	//words := []models.LearnRecord{}
	//DB.Order("id desc").Preload("Word").Where("review_time < ?", endOfDayTimestamp).Find(&words)
	//notin := []uint{}
	//for _, word := range words {
	//	notin = append(notin, word.WordID)
	//}
	//randomWords := []models.WordBookRelation{}
	//DB.Order("RAND()").Preload("Word").Where("id NOT IN ?", notin).Limit(len(words) * 4).Find(&randomWords)
	//result := []reviewRes{}
	//for _, word := range words {
	//	var progress []bool
	//	if word.Word.Kana == word.Word.Word {
	//		progress = []bool{false, false, false}
	//	} else {
	//		progress = []bool{false, false, false}
	//	}
	//	today := reviewRes{
	//		Exercise:      false,
	//		Done:          false,
	//		Tone:          word.Word.Tone,
	//		ErrorCount:    0,
	//		Progress:      progress,
	//		Meaning:       getMeaning(word.Word.Detail),
	//		Word:          word.Word.Word,
	//		Kana:          word.Word.Kana,
	//		ID:            word.Word.ID,
	//		Rome:          word.Word.Rome,
	//		Voice:         word.Word.Voice,
	//		Detail:        word.Word.Detail,
	//		MeaningOption: getMeaningOption(word.Word, randomWords),
	//	}
	//	result = append(result, today)
	//}
	//c.JSON(http.StatusOK, gin.H{
	//	"data": result,
	//})
}
func (h *WordHandler) getWriteFromMemory(c *gin.Context) {
	//words := []models.LearnRecord{}
	//writeFromMemorys := []writeFromMemory{}
	//todayStart := time.Now().Truncate(24 * time.Hour)
	//todayEnd := todayStart.Add(24*time.Hour - 1*time.Second)
	//DB.Order("id desc").
	//	Preload("Word").
	//	Where("created_at BETWEEN ? AND ?", todayStart.Format("2006-01-02 15:04:05"), todayEnd.Format("2006-01-02 15:04:05")).
	//	Find(&words)
	//for _, word := range words {
	//	writeFromMemorys = append(writeFromMemorys, writeFromMemory{
	//		Word:       word.Word.Word,
	//		Meaning:    strings.Join(getMeaning(word.Word.Detail), ";"),
	//		Kana:       word.Word.Kana,
	//		Tone:       word.Word.Tone,
	//		ID:         word.WordID,
	//		ErrorCount: 0,
	//		Rome:       word.Word.Rome,
	//		Voice:      word.Word.Voice,
	//		Done:       false,
	//	})
	//}
	//c.JSON(http.StatusOK, gin.H{
	//	"data": writeFromMemorys,
	//})
}

type option struct {
	Text   []string `json:"text"`
	Answer bool     `json:"answer"`
}

func (h *WordHandler) getNewWord(c *gin.Context) {
	UserId, _ := c.Get("UserId")
	var config models.UserConfig
	DB.First(&config, "user_id = ?", UserId)
	wordbooks := make([]models.WordBookRelation, 0)
	result := make([]WordInfo, 0)
	DB.Debug().
		Preload("Word.Meaning").Preload("Word.Example").Joins("LEFT JOIN learn_record lr ON lr.word_id = word_book_relation.word_id").
		Where("lr.word_id IS NULL AND word_book_relation.book_id = ?", config.BookID).
		Order("word_book_relation.id DESC").
		Limit(config.LearnGroup).
		Find(&wordbooks)
	for _, word := range wordbooks {
		wordinfo := WordInfo{
			ID:       word.Word.ID,
			Word:     word.Word.Word,
			Tone:     word.Word.Tone,
			Rome:     word.Word.Rome,
			Browse:   word.Word.Browse,
			Voice:    word.Word.Voice,
			Kana:     word.Word.Kana,
			Wordtype: word.Word.Wordtype,
			Detail:   word.Word.Detail,
			Meaning:  word.Word.Meaning,
			Example:  word.Word.Example,
		}
		result = append(result, wordinfo)
	}
	c.JSON(http.StatusOK, gin.H{
		"data":  result,
		"total": config.LearnGroup,
		"msg":   "Successfully obtained",
	})
}
func (h *WordHandler) getWordBookList(c *gin.Context) {
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
	bookId, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}
	wordbooks := make([]models.WordBookRelation, 0)
	result := make([]JcdictRes, 0)
	var total int64
	DB.Order("id desc").Preload("Word.Meaning").Where("book_id = ?", bookId).Limit(size).Offset(size * (page - 1)).Find(&wordbooks)
	DB.Model(models.WordBookRelation{}).Where("book_id = ?", bookId).Count(&total)
	for _, book := range wordbooks {
		meanings := make([]string, 0)
		for _, meaning := range book.Word.Meaning {
			meanings = append(meanings, meaning.Meaning)
		}
		result = append(result, JcdictRes{
			Word:    book.Word.Word,
			Kana:    book.Word.Kana,
			ID:      book.Word.ID,
			Meaning: meanings,
		})
	}
	c.JSON(http.StatusOK, gin.H{
		"data":  result,
		"total": total,
		"msg":   "Successfully obtained",
	})
}
func (h *WordHandler) getWordBook(c *gin.Context) {
	//res := make([]wordBookRes, 0)
	//wordbook := make([]models.WordBook, 0)
	//DB.Model(&models.WordBook{}).Preload("Words").Find(&wordbook)
	//for _, word := range wordbook {
	//	res = append(res, wordBookRes{
	//		ID:       word.ID,
	//		Name:     word.Name,
	//		Icon:     word.Icon,
	//		Category: word.Category,
	//		Words:    len(word.Words),
	//		Describe: word.Describe,
	//	})
	//}
	//c.JSON(http.StatusOK, gin.H{
	//	"data": res,
	//	"msg":  "Successfully obtained",
	//})
}

type JcdictRes struct {
	Word    string   `json:"word"`
	Kana    string   `json:"kana"`
	ID      uint     `json:"id"`
	Browse  int      `json:"browse"`
	Meaning []string `json:"meaning"`
	Book    []string `json:"book"`
}

//type ChRes struct {
//	Ch     string `json:"ch"`
//	Pinyin string `json:"pinyin"`
//	Result []Res  `json:"result"`
//}

func (h *WordHandler) cjInfo(c *gin.Context) {
	//id, err := strconv.Atoi(c.Param("id"))
	//if err != nil {
	//	c.JSON(http.StatusBadRequest, gin.H{"err": "The id format is incorrect"})
	//	return
	//}
	//var res ChRes
	//var Word models.Chdict
	//err = DB.Select("ch", "ja", "pinyin").Model(models.Chdict{}).First(&Word, id).Error
	//if errors.Is(err, gorm.ErrRecordNotFound) {
	//	c.JSON(http.StatusNotFound, gin.H{"err": "The word does not exist"})
	//	return
	//}
	//res.Ch = Word.Ch
	//res.Pinyin = Word.Pinyin
	//Word.Browse += 1
	//DB.Save(&Word)
	//var JaWords []models.Jadict
	//if len(Word.Ja) > 0 {
	//	DB.Select("detail", "id", "kana", "word").Model(&models.Jadict{}).Where("word IN ?", Word.Ja).Find(&JaWords)
	//}
	//var Res1 Res
	//for _, v1 := range JaWords {
	//	Res1.Meaning = getMeaning(v1.Detail)
	//	Res1.ID = v1.ID
	//	Res1.Kana = v1.Kana
	//	Res1.Word = v1.Word
	//	res.Result = append(res.Result, Res1)
	//}
	//c.JSON(http.StatusOK, gin.H{
	//	"msg":  "Successfully obtained",
	//	"data": res,
	//})
}

type WordInfo struct {
	ID       uint                   `json:"id"`
	Word     string                 `json:"word"`
	Tone     string                 `json:"tone"`
	Rome     string                 `json:"rome"`
	Browse   int                    `json:"browse"`
	Voice    string                 `json:"voice"`
	Kana     string                 `json:"kana"`
	Wordtype string                 `json:"wordtype"`
	Detail   string                 `json:"detail"`
	Meaning  []models.JcdictMeaning `gorm:"foreignKey:WordID;references:ID" json:"meaning"`
	Example  []models.JcdictExample `gorm:"foreignKey:WordID;references:ID" json:"example"`
}

func (h *WordHandler) jcInfo(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"err": "The id format is incorrect"})
		return
	}
	Word := WordInfo{}
	result := DB.Preload("Meaning").Preload("Example").Model(models.Jcdict{}).First(&Word, id).Error
	if errors.Is(result, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusNotFound, gin.H{
			"msg": "Word does not exist",
		})
		return
	}
	Word.Browse += 1
	DB.Model(models.Jcdict{}).Where("id = ?", Word.ID).Update("browse", Word.Browse)
	c.JSON(http.StatusOK, gin.H{
		"msg":  "Successfully obtained",
		"data": Word,
	})
}
func (h *WordHandler) jcList(c *gin.Context) {
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
	Word := make([]models.Jcdict, 0)
	DB.Preload("Meaning").Select("rome", "id", "word", "kana", "word_type").
		Limit(size).
		Offset(size * (page - 1)).
		Find(&Word)
	var total int64 = 0
	DB.Model(&models.Jcdict{}).Select("id").Count(&total)
	c.JSON(http.StatusOK, gin.H{
		"msg":   "Successfully obtained",
		"data":  Word,
		"total": total,
	})
}
func (h *WordHandler) jcSearch(c *gin.Context) {
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
	Word := make([]models.Jcdict, 0)
	val := c.Param("val")
	searchTerm := fmt.Sprintf("'%s*'", val)
	DB.Preload("Book.Book").Preload("Meaning").Select("browse", "id", "word", "kana").
		Where("MATCH(word,kana) AGAINST(? IN BOOLEAN MODE)", searchTerm).
		Order("LENGTH(word)").
		Limit(size).
		Offset(size * (page - 1)).
		Find(&Word)
	var total int64 = 0
	DB.Model(&models.Jcdict{}).Select("word", "kana").Where("MATCH(word,kana) AGAINST(? IN BOOLEAN MODE)", searchTerm).
		Count(&total)
	var Result []JcdictRes
	for _, v := range Word {
		meanings := make([]string, 0)
		for _, meaning := range v.Meaning {
			meanings = append(meanings, meaning.Meaning)
		}
		books := make([]string, 0)
		for _, book := range v.Book {
			books = append(books, book.Book.Tag)
		}
		item := JcdictRes{
			Word:    v.Word,
			Kana:    v.Kana,
			ID:      v.ID,
			Browse:  v.Browse,
			Meaning: meanings,
			Book:    books,
		}
		Result = append(Result, item)
	}
	c.JSON(http.StatusOK, gin.H{
		"msg":   "Successfully obtained",
		"data":  Result,
		"total": total,
	})
}
func (h *WordHandler) cjSearch(c *gin.Context) {
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
	var Word []models.Chdict
	val := c.Param("val")
	searchTerm := fmt.Sprintf("'%s*'", val)
	var total int64
	DB.Raw("SELECT ch,id,pinyin FROM chdict WHERE MATCH(ch) AGAINST(? IN BOOLEAN MODE) order by LENGTH(ch) limit ? offset ?", searchTerm, size, size*(page-1)).Scan(&Word)
	DB.Raw("SELECT ch,id,pinyin FROM chdict WHERE MATCH(ch) AGAINST(? IN BOOLEAN MODE)", searchTerm).Count(&total)
	c.JSON(http.StatusOK, gin.H{
		"msg":   "Successfully obtained",
		"data":  Word,
		"total": total,
	})
}
func removeParenthesesContent(input string) string {
	// 正则表达式，匹配所有括号和括号内的内容
	re := regexp.MustCompile(`[（(\[【〔［「].*?[」］）)\]】〕]`)

	// 替换匹配到的部分为空字符串
	result := re.ReplaceAllString(input, "")

	// 去除多余的空格
	result = regexp.MustCompile(`\s+`).ReplaceAllString(result, " ")

	return result
}
func getMeaning(detail []models.Detail) []string {
	var res []string
	for _, v := range detail {
		for _, v1 := range v.Detail {
			res = append(res, removeParenthesesContent(v1.Meaning))
		}
	}
	return res
}

// 新增学习记录
func (h *WordHandler) addLearnRecord(c *gin.Context) {
	var learnRecord []models.LearnRecord
	var Req struct {
		Words []uint `json:"words"`
	}
	if err := c.ShouldBindJSON(&Req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"err": err.Error()})
		return
	}
	timestamp := time.Now().Unix() + 86400
	UserId, _ := c.Get("UserId")
	for _, item := range Req.Words {
		learnRecord = append(learnRecord, models.LearnRecord{
			WordID:     item,
			UserID:     UserId.(uint),
			ReviewTime: timestamp,
		})
	}
	DB.Create(&learnRecord)
	c.JSON(http.StatusOK, gin.H{
		"msg": "Record successful",
	})
}

// 推荐单词
func (h *WordHandler) recommendWord(c *gin.Context) {
	recommendWords := make([]models.Jcdict, 0)
	DB.Order("browse desc").Limit(10).Find(&recommendWords)
	var Result []string
	for _, v := range recommendWords {
		Result = append(Result, v.Word)
	}
	c.JSON(http.StatusOK, gin.H{
		"msg":  "Successfully obtained",
		"data": Result,
	})
}
