package handlers

import (
	. "easyjapanese/db"
	"easyjapanese/internal/middleware"
	"easyjapanese/internal/models"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"math/rand"
	"net/http"
	"regexp"
	"strconv"
	"strings"
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
	router.GET("/wordbook", h.getWordBook)
	router.GET("/todayword", middleware.User(), h.getTodayWord)
	router.GET("/wordbook/:id/:page/:size", h.getWordBookList)
	jc := router.Group("/jc")
	{
		jc.POST("/add", h.jcAddWord)
		jc.GET("/search/:page/:size/:val", h.jcSearch)
		jc.GET("/info/:id", h.jcInfo)
	}
	cj := router.Group("/cj")
	{
		cj.POST("/add", h.jcAddWord)
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
	learnRecords := []models.LearnRecord{}
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
	wordids := []uint{}
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
	now := time.Now()
	midnight := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	endOfDay := midnight.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
	endOfDayTimestamp := endOfDay.Unix()
	words := []models.LearnRecord{}
	DB.Order("id desc").Preload("Word").Where("review_time < ?", endOfDayTimestamp).Find(&words)
	notin := []uint{}
	for _, word := range words {
		notin = append(notin, word.WordID)
	}
	randomWords := []models.WordBookRelation{}
	DB.Order("RAND()").Preload("Word").Where("id NOT IN ?", notin).Limit(len(words) * 4).Find(&randomWords)
	result := []reviewRes{}
	for _, word := range words {
		var progress []bool
		if word.Word.Kana == word.Word.Word {
			progress = []bool{false, false, false}
		} else {
			progress = []bool{false, false, false}
		}
		today := reviewRes{
			Exercise:      false,
			Done:          false,
			Tone:          word.Word.Tone,
			ErrorCount:    0,
			Progress:      progress,
			Meaning:       getMeaning(word.Word.Detail),
			Word:          word.Word.Word,
			Kana:          word.Word.Kana,
			ID:            word.Word.ID,
			Rome:          word.Word.Rome,
			Voice:         word.Word.Voice,
			Detail:        word.Word.Detail,
			MeaningOption: getMeaningOption(word.Word, randomWords),
		}
		result = append(result, today)
	}
	c.JSON(http.StatusOK, gin.H{
		"data": result,
	})
}
func (h *WordHandler) getWriteFromMemory(c *gin.Context) {
	words := []models.LearnRecord{}
	writeFromMemorys := []writeFromMemory{}
	todayStart := time.Now().Truncate(24 * time.Hour)
	todayEnd := todayStart.Add(24*time.Hour - 1*time.Second)
	DB.Order("id desc").
		Preload("Word").
		Where("created_at BETWEEN ? AND ?", todayStart.Format("2006-01-02 15:04:05"), todayEnd.Format("2006-01-02 15:04:05")).
		Find(&words)
	for _, word := range words {
		writeFromMemorys = append(writeFromMemorys, writeFromMemory{
			Word:       word.Word.Word,
			Meaning:    strings.Join(getMeaning(word.Word.Detail), ";"),
			Kana:       word.Word.Kana,
			Tone:       word.Word.Tone,
			ID:         word.WordID,
			ErrorCount: 0,
			Rome:       word.Word.Rome,
			Voice:      word.Word.Voice,
			Done:       false,
		})
	}
	c.JSON(http.StatusOK, gin.H{
		"data": writeFromMemorys,
	})
}

type option struct {
	Text   string `json:"text"`
	Answer bool   `json:"answer"`
}
type todayWordRes struct {
	Done          bool            `json:"done"`
	Tone          string          `json:"tone"`
	ErrorCount    int             `json:"error_count"`
	Progress      []bool          `json:"progress"`
	Meaning       string          `json:"meaning"`
	Word          string          `json:"word"`
	Kana          string          `json:"kana"`
	ID            uint            `json:"id"`
	Rome          string          `json:"rome"`
	Voice         string          `json:"voice"`
	Detail        []models.Detail `json:"detail" gorm:"serializer:json"`
	KanaOption    []option        `json:"kana_option"`
	MeaningOption []option        `json:"meaning_option"`
	WordOption    []option        `json:"word_option"`
	VoiceOption   []option        `json:"voice_option"`
}

func extractKanaFromEnd(input string) string {
	re := regexp.MustCompile(`[ぁ-ゖァ-ヶー]+$`)
	matches := re.FindString(input)
	return matches
}

func isPureKana(s string) bool {
	re := regexp.MustCompile(`^[ぁ-ゖァ-ヶー]+$`)
	return re.MatchString(s)
}

func getKanaOption(word string, kana string) []option {
	options := make([]option, 0)
	if word == kana {
		return options
	}
	jadicts := make([]models.Jadict, 0)
	if isPureKana(word) {
		DB.Preload("Word").Where("LENGTH(kana) <= ?", 7).Where("kana not in ? and id >= (SELECT FLOOR(RAND() * (SELECT MAX(id) FROM jadict)))", kana).Limit(3).Find(&jadicts)
	} else {
		suffix := extractKanaFromEnd(word)
		if suffix == "" {
			DB.Where("LENGTH(kana) <= ?", 7).Where("kana != ? and id >= (SELECT FLOOR(RAND() * (SELECT MAX(id) FROM jadict)))", kana).Limit(3).Find(&jadicts)
		} else {
			searchTerm := fmt.Sprintf("%%%s", suffix)
			DB.Raw("select * from jadict where kana NOT LIKE ? and kana LIKE ? limit 3", kana, searchTerm).Scan(&jadicts)
		}
	}
	if len(jadicts) < 3 {
		random := make([]models.Jadict, 0)
		DB.Where("kana != ? and id >= (SELECT FLOOR(RAND() * (SELECT MAX(id) FROM jadict)))", kana).Limit(3 - len(jadicts)).Find(&random)
		jadicts = append(jadicts, random...)
	}
	for _, jadict := range jadicts {
		options = append(options, option{
			Text:   retainKana(jadict.Kana),
			Answer: false,
		})
	}
	options = append(options, option{
		Text:   retainKana(kana),
		Answer: true,
	})
	return options
}
func getRandomElements(slice []models.WordBookRelation, n int) []models.WordBookRelation {
	if n > len(slice) {
		return nil
	}
	indices := rand.Perm(len(slice))
	result := make([]models.WordBookRelation, n)
	for i := 0; i < n; i++ {
		result[i] = slice[indices[i]]
	}
	return result
}
func getMeaningOption(word models.Jadict, randomWords []models.WordBookRelation) []option {
	options := []option{}
	wordbooks := getRandomElements(randomWords, 3)
	for _, wordbook := range wordbooks {
		options = append(options, option{
			Answer: false,
			Text:   strings.Join(getMeaning(wordbook.Word.Detail), ";"),
		})
	}
	options = append(options, option{
		Answer: true,
		Text:   strings.Join(getMeaning(word.Detail), ";"),
	})
	return options
}
func retainKana(input string) string {
	re := regexp.MustCompile(`[^\x{3040}-\x{309F}\x{30A0}-\x{30FF}]`)
	result := re.ReplaceAllString(input, "")
	return result
}
func getWordOption(word models.Jadict, randomWords []models.WordBookRelation) []option {
	options := []option{}
	wordbooks := getRandomElements(randomWords, 3)
	for _, wordbook := range wordbooks {
		options = append(options, option{
			Answer: false,
			Text:   wordbook.Word.Kana,
		})
	}
	options = append(options, option{
		Answer: true,
		Text:   word.Kana,
	})
	return options
}
func getVoiceOption(word models.Jadict, randomWords []models.WordBookRelation) []option {
	options := []option{}
	wordbooks := getRandomElements(randomWords, 3)
	for _, wordbook := range wordbooks {
		options = append(options, option{
			Answer: false,
			Text:   wordbook.Word.Word,
		})
	}
	options = append(options, option{
		Answer: true,
		Text:   word.Word,
	})
	return options
}
func (h *WordHandler) getTodayWord(c *gin.Context) {
	UserId, _ := c.Get("UserId")
	//获取用户配置
	var config models.UserConfig
	DB.First(&config, "user_id = ?", UserId)
	wordbooks := make([]models.WordBookRelation, 0)
	result := make([]todayWordRes, 0)
	DB.Joins("LEFT JOIN learn_record lr ON lr.word_id = word_book_relation.word_id").
		Where("lr.word_id IS NULL AND word_book_relation.book_id = ?", config.BookID).
		Order("word_book_relation.id DESC").
		Limit(config.LearnGroup).
		Preload("Word").
		Find(&wordbooks)
	//获取随机单词充当选项
	notin := make([]uint, 0)
	for _, word := range wordbooks {
		notin = append(notin, word.WordId)
	}
	randomWords := make([]models.WordBookRelation, 0)
	DB.Preload("Word").Where("id not in ? and id >= (SELECT FLOOR(RAND() * (SELECT MAX(id) FROM word_book_relation)))", notin).Limit(config.LearnGroup * 40).Find(&randomWords)
	for _, word := range wordbooks {
		var progress []bool
		if word.Word.Kana == word.Word.Word {
			progress = []bool{true, false, false, false}
		} else {
			progress = []bool{false, false, false, false}
		}
		today := todayWordRes{
			Done:          false,
			Tone:          word.Word.Tone,
			ErrorCount:    0,
			Progress:      progress,
			Meaning:       strings.Join(getMeaning(word.Word.Detail), ";"),
			Word:          word.Word.Word,
			Kana:          word.Word.Kana,
			ID:            word.Word.ID,
			Rome:          word.Word.Rome,
			Voice:         word.Word.Voice,
			Detail:        word.Word.Detail,
			KanaOption:    getKanaOption(word.Word.Word, word.Word.Kana),
			MeaningOption: getMeaningOption(word.Word, randomWords),
			WordOption:    getWordOption(word.Word, randomWords),
			VoiceOption:   getVoiceOption(word.Word, randomWords),
		}
		result = append(result, today)
	}
	c.JSON(http.StatusOK, gin.H{
		"data": result,
		"msg":  "Successfully obtained",
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
	bookId, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}
	wordbooks := []models.WordBookRelation{}
	result := []Res{}
	var total int64
	DB.Order("id desc").Preload("Word").Where("book_id = ?", bookId).Limit(size).Offset(size * (page - 1)).Find(&wordbooks)
	DB.Model(models.WordBookRelation{}).Where("book_id = ?", bookId).Count(&total)
	for _, book := range wordbooks {
		result = append(result, Res{
			Word:    book.Word.Word,
			Kana:    book.Word.Kana,
			ID:      book.Word.ID,
			Meaning: getMeaning(book.Word.Detail),
		})
	}
	c.JSON(http.StatusOK, gin.H{
		"data":  result,
		"total": total,
		"msg":   "Successfully obtained",
	})
}
func (h *WordHandler) getWordBook(c *gin.Context) {
	res := []wordBookRes{}
	wordbook := []models.Wordbook{}
	DB.Model(&models.Wordbook{}).Preload("Words").Find(&wordbook)
	for _, word := range wordbook {
		res = append(res, wordBookRes{
			ID:       word.ID,
			Name:     word.Name,
			Icon:     word.Icon,
			Category: word.Category,
			Words:    len(word.Words),
			Describe: word.Describe,
		})
	}
	c.JSON(http.StatusOK, gin.H{
		"data": res,
		"msg":  "Successfully obtained",
	})
}

func (h *WordHandler) jcAddWord(c *gin.Context) {
	var Word models.Jadict
	if err := c.ShouldBindJSON(&Word); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"err": err.Error()})
		return
	}
	result := DB.Create(&Word)
	if result.Error != nil {
		c.JSON(http.StatusBadRequest, gin.H{"err": result.Error.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"msg": "Successfully Added",
	})
}

type Res struct {
	Word    string   `json:"word"`
	Kana    string   `json:"kana"`
	ID      uint     `json:"id"`
	Meaning []string `json:"meaning"`
}
type ChRes struct {
	Ch     string `json:"ch"`
	Pinyin string `json:"pinyin"`
	Result []Res  `json:"result"`
}

func (h *WordHandler) cjInfo(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"err": "The id format is incorrect"})
		return
	}
	var res ChRes
	var Word models.Chdict
	err = DB.Select("ch", "ja", "pinyin").Model(models.Chdict{}).First(&Word, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"err": "The word does not exist"})
		return
	}
	res.Ch = Word.Ch
	res.Pinyin = Word.Pinyin
	var JaWords []models.Jadict
	if len(Word.Ja) > 0 {
		DB.Debug().Select("detail", "id", "kana", "word").Model(&models.Jadict{}).Where("word IN ?", Word.Ja).Find(&JaWords)
	}
	var Res1 Res
	for _, v1 := range JaWords {
		Res1.Meaning = getMeaning(v1.Detail)
		Res1.ID = v1.ID
		Res1.Kana = v1.Kana
		Res1.Word = v1.Word
		res.Result = append(res.Result, Res1)
	}
	c.JSON(http.StatusOK, gin.H{
		"msg":  "Successfully obtained",
		"data": res,
	})
}
func (h *WordHandler) jcInfo(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"err": "The id format is incorrect"})
		return
	}
	var Word models.Jadict
	DB.Model(models.Jadict{}).Omit("created_at", "updated_at").First(&Word, id)
	c.JSON(http.StatusOK, gin.H{
		"msg":  "Successfully obtained",
		"data": Word,
	})
}
func (h *WordHandler) jcSearch(c *gin.Context) {
	var Res1 Res
	var Res2 []Res
	var Word []models.Jadict
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
	searchTerm := fmt.Sprintf("'%s*'", val)
	var total int64
	DB.Raw("SELECT word,kana,detail,id FROM jadict WHERE MATCH(word,kana) AGAINST(? IN BOOLEAN MODE) order by LENGTH(word) limit ? offset ?", searchTerm, size, size*(page-1)).Scan(&Word)
	DB.Raw("SELECT word,kana,detail,id FROM jadict WHERE MATCH(word,kana) AGAINST(? IN BOOLEAN MODE)", searchTerm).Count(&total)
	if total > 0 {
		for _, v := range Word {
			Res1.Meaning = getMeaning(v.Detail)
			Res1.ID = v.ID
			Res1.Kana = v.Kana
			Res1.Word = v.Word
			Res2 = append(Res2, Res1)
		}
	}
	c.JSON(http.StatusOK, gin.H{
		"msg":   "Successfully obtained",
		"data":  Res2,
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
