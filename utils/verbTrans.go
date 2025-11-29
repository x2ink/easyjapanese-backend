package utils

import (
	"errors"
	"strings"

	"github.com/ikawaha/kagome-dict/ipa"
	"github.com/ikawaha/kagome/v2/tokenizer"
)

// --- 全局映射表 (提升性能) ---

// uToA: 未然形映射 (用于否定、被动、使役)
// 注意: 五段动词以「う」结尾时，未然形是「わ」(wa)，而不是「あ」(a)
var uToA = map[string]string{
	"う": "わ", "く": "か", "す": "さ", "つ": "た", "ぬ": "な",
	"ふ": "は", "む": "ま", "ゆ": "や", "る": "ら", "ぐ": "が",
	"ず": "ざ", "ぶ": "ば", "ぷ": "ぱ", "づ": "だ",
}

// uToI: 连用形映射 (用于ます形)
var uToI = map[string]string{
	"う": "い", "く": "き", "す": "し", "つ": "ち", "ぬ": "に",
	"ふ": "ひ", "む": "み", "ゆ": "い", "る": "り", "ぐ": "ぎ",
	"ず": "じ", "ぶ": "び", "ぷ": "ぴ", "づ": "ぢ",
}

// uToE: 假定形/命令形/可能形映射
var uToE = map[string]string{
	"う": "え", "く": "け", "す": "せ", "つ": "て", "ぬ": "ね",
	"ふ": "へ", "む": "め", "ゆ": "え", "る": "れ", "ぐ": "げ",
	"ず": "ぜ", "ぶ": "べ", "ぷ": "ぺ", "づ": "で",
}

// uToO: 意向形映射
var uToO = map[string]string{
	"う": "お", "く": "こ", "す": "そ", "つ": "と", "ぬ": "の",
	"ふ": "ほ", "む": "も", "ゆ": "よ", "る": "ろ", "ぐ": "ご",
	"ず": "ぞ", "ぶ": "ぼ", "ぷ": "ぽ", "づ": "ど",
}

// uToTe: て形映射 (五段动词音便)
var uToTe = map[string]string{
	"う": "って", "つ": "って", "る": "って", // 促音便
	"ぬ": "んで", "む": "んで", "ぶ": "んで", // 拨音便
	"く": "いて", // イ音便 (例外: 行く -> 行って)
	"ぐ": "いで",
	"す": "して",
}

// uToTa: た形映射 (同て形)
var uToTa = map[string]string{
	"う": "った", "つ": "った", "る": "った",
	"ぬ": "んだ", "む": "んだ", "ぶ": "んだ",
	"く": "いた",
	"ぐ": "いだ",
	"す": "した",
}

// --- 辅助函数 ---

func isVerb(str string) (string, error) {
	// 优先级: 优先匹配变格动词，防止误判
	// 例如 "勉強する" 包含 "する" (サ変)，但也可能被误判
	if strings.Contains(str, "サ変") {
		return "サ変", nil
	}
	if strings.Contains(str, "カ変") {
		return "カ変", nil
	}
	if strings.Contains(str, "一段") {
		return "一段", nil
	}
	if strings.Contains(str, "五段") {
		return "五段", nil
	}
	return "", errors.New("not a verb")
}

func replaceLast(original, old, new string) string {
	lastIndex := strings.LastIndex(original, old)
	if lastIndex == -1 {
		return original
	}
	return original[:lastIndex] + new + original[lastIndex+len(old):]
}

func getLastChar(word string) string {
	runes := []rune(word)
	if len(runes) == 0 {
		return ""
	}
	return string(runes[len(runes)-1])
}

// --- 变形函数 ---

// MasuForm 动词ます形 (连用形)
func MasuForm(word string, verbType string) string {
	if verbType == "一段" {
		return replaceLast(word, "る", "ます")
	} else if verbType == "五段" {
		lastChar := getLastChar(word)
		return replaceLast(word, lastChar, uToI[lastChar]) + "ます"
	} else if verbType == "カ変" {
		// 来る(kuru) -> 来ます(kimasu)
		return replaceLast(word, "来る", "来ます")
	} else if verbType == "サ変" {
		return replaceLast(word, "する", "します")
	}
	return word
}

// NegationForm 动词ない形 (未然形)
func NegationForm(word string, verbType string) string {
	// 特例: ある -> ない
	if word == "ある" {
		return "ない"
	}

	if verbType == "一段" {
		return replaceLast(word, "る", "ない")
	} else if verbType == "五段" {
		lastChar := getLastChar(word)
		return replaceLast(word, lastChar, uToA[lastChar]) + "ない"
	} else if verbType == "カ変" {
		// 来る(kuru) -> 来ない(konai)
		return replaceLast(word, "来る", "来ない")
	} else if verbType == "サ変" {
		return replaceLast(word, "する", "しない")
	}
	return word
}

// TeJointForm 动词て形
func TeJointForm(word string, verbType string) string {
	// 特例: 行く -> 行って
	if word == "行く" {
		return "行って"
	}

	if verbType == "一段" {
		return replaceLast(word, "る", "て")
	} else if verbType == "五段" {
		lastChar := getLastChar(word)
		return replaceLast(word, lastChar, uToTe[lastChar])
	} else if verbType == "カ変" {
		return replaceLast(word, "来る", "来て")
	} else if verbType == "サ変" {
		return replaceLast(word, "する", "して")
	}
	return word
}

// TaJointForm 动词た形
func TaJointForm(word string, verbType string) string {
	// 特例: 行く -> 行った
	if word == "行く" {
		return "行った"
	}

	if verbType == "一段" {
		return replaceLast(word, "る", "た")
	} else if verbType == "五段" {
		lastChar := getLastChar(word)
		return replaceLast(word, lastChar, uToTa[lastChar])
	} else if verbType == "カ変" {
		return replaceLast(word, "来る", "来た")
	} else if verbType == "サ変" {
		return replaceLast(word, "する", "した")
	}
	return word
}

// CommandForm 动词命令形
func CommandForm(word string, verbType string) string {
	if verbType == "一段" {
		return replaceLast(word, "る", "ろ")
	} else if verbType == "五段" {
		lastChar := getLastChar(word)
		return replaceLast(word, lastChar, uToE[lastChar])
	} else if verbType == "カ変" {
		return replaceLast(word, "来る", "来い")
	} else if verbType == "サ変" {
		return replaceLast(word, "する", "しろ")
	}
	return word
}

// PossibleForm 动词可能形
func PossibleForm(word string, verbType string) string {
	if verbType == "一段" {
		return replaceLast(word, "る", "られる")
	} else if verbType == "五段" {
		lastChar := getLastChar(word)
		// 五段: e段 + る (書ける)
		return replaceLast(word, lastChar, uToE[lastChar]) + "る"
	} else if verbType == "カ変" {
		return replaceLast(word, "来る", "来られる")
	} else if verbType == "サ変" {
		return replaceLast(word, "する", "できる")
	}
	return word
}

// AssumingForm 假定形 (ば形)
func AssumingForm(word string, verbType string) string {
	if verbType == "一段" {
		return replaceLast(word, "る", "れば")
	} else if verbType == "五段" {
		lastChar := getLastChar(word)
		return replaceLast(word, lastChar, uToE[lastChar]) + "ば"
	} else if verbType == "カ変" {
		return replaceLast(word, "来る", "来れば")
	} else if verbType == "サ変" {
		return replaceLast(word, "する", "すれば")
	}
	return word
}

// IntentionalForm 意向形 (よう形)
func IntentionalForm(word string, verbType string) string {
	if verbType == "一段" {
		return replaceLast(word, "る", "よう")
	} else if verbType == "五段" {
		lastChar := getLastChar(word)
		// 五段: o段 + う (書こう)
		return replaceLast(word, lastChar, uToO[lastChar]) + "う"
	} else if verbType == "カ変" {
		return replaceLast(word, "来る", "来よう")
	} else if verbType == "サ変" {
		return replaceLast(word, "する", "しよう")
	}
	return word
}

// PassiveForm 被动形 (受身形) - 修正: 使用未然形(a段)
func PassiveForm(word string, verbType string) string {
	if verbType == "一段" {
		return replaceLast(word, "る", "られる")
	} else if verbType == "五段" {
		lastChar := getLastChar(word)
		// 五段: a段 + れる (書かれる)
		return replaceLast(word, lastChar, uToA[lastChar]) + "れる"
	} else if verbType == "カ変" {
		return replaceLast(word, "来る", "来られる")
	} else if verbType == "サ変" {
		return replaceLast(word, "する", "される")
	}
	return word
}

// CausativeForm 使役形 - 修正: 使用未然形(a段)
func CausativeForm(word string, verbType string) string {
	if verbType == "一段" {
		return replaceLast(word, "る", "させる")
	} else if verbType == "五段" {
		lastChar := getLastChar(word)
		// 五段: a段 + せる (書かせる)
		return replaceLast(word, lastChar, uToA[lastChar]) + "せる"
	} else if verbType == "カ変" {
		return replaceLast(word, "来る", "来させる")
	} else if verbType == "サ変" {
		return replaceLast(word, "する", "させる")
	}
	return word
}

// CausativePassiveForm 使役被动形 - 修正: 使用未然形(a段)
func CausativePassiveForm(word string, verbType string) string {
	if verbType == "一段" {
		return replaceLast(word, "る", "させられる")
	} else if verbType == "五段" {
		lastChar := getLastChar(word)
		// 五段: a段 + せられる (书面/完整) 或 a段 + される (口语缩约，su结尾除外)
		// 这里采用完整形式 "せられる" 以保持规则统一，或者更常见的 "させられる" (一段) / "される" (五段缩约)
		// 标准语法: 五段 a段 + せる + られる -> a段 + せられる
		// 现代日语常用缩约: a段 + される (行かされる)
		// 注意: 以「す」结尾的动词(話す)不能缩约，只能用「話させられる」
		if lastChar == "す" {
			return replaceLast(word, lastChar, uToA[lastChar]) + "せられる"
		}
		return replaceLast(word, lastChar, uToA[lastChar]) + "される"
	} else if verbType == "カ変" {
		return replaceLast(word, "来る", "来させられる")
	} else if verbType == "サ変" {
		return replaceLast(word, "する", "させられる")
	}
	return word
}

// --- 结构体定义 ---

type wordForm struct {
	BaseForm   string
	OriginForm string
}

type TransRes struct {
	Category string `json:"category"`
	Result   string `json:"result"`
}

// VerbTransfiguration 主入口函数
func VerbTransfiguration(str string) []TransRes {
	// 初始化分词器 (注意: 生产环境中建议将 tokenizer 作为全局单例，避免重复 Load 字典)
	t, err := tokenizer.New(ipa.Dict(), tokenizer.OmitBosEos())
	if err != nil {
		// 实际项目中建议返回 error 而不是 panic
		panic(err)
	}
	tokens := t.Tokenize(str)
	words := []wordForm{}
	var verbType string

	for _, v := range tokens {
		if v.Class == tokenizer.DUMMY {
			continue
		}
		r := tokenizer.NewTokenData(v)
		// 简单的动词判断逻辑
		verb, err := isVerb(strings.Join(r.Features, " "))
		if err != nil {
			// 如果句子中包含非动词部分，这里只处理识别到的动词
			continue
		}
		verbType = verb
		wordform := wordForm{
			BaseForm:   r.BaseForm,
			OriginForm: r.Surface,
		}
		words = append(words, wordform)
	}

	if len(words) == 0 {
		return nil
	}

	// 拼接单词 (针对复合词或带前缀的情况)
	word := ""
	for i, v := range words {
		if len(words)-1 == i {
			word += v.BaseForm
		} else {
			word += v.OriginForm
		}
	}

	// 针对 "する" 动词的特殊修正
	if verbType == "サ変" {
		if !strings.HasSuffix(word, "する") {
			// 尝试修复 basic form，有些分词结果可能是名词+する分开
			word = word + "する"
		}
	}

	res := make([]TransRes, 0)

	// 辅助添加结果的闭包
	add := func(cat, result string) {
		res = append(res, TransRes{Category: cat, Result: result})
	}

	add("基本形", word)
	add("ます形", MasuForm(word, verbType))
	add("ない形", NegationForm(word, verbType))
	add("て形", TeJointForm(word, verbType))
	add("た形", TaJointForm(word, verbType))
	add("命令形", CommandForm(word, verbType))
	add("可能形", PossibleForm(word, verbType))
	add("假定形", AssumingForm(word, verbType))
	add("意向形", IntentionalForm(word, verbType))
	add("被动形", PassiveForm(word, verbType))
	add("使役形", CausativeForm(word, verbType))
	add("禁止形", word+"な")
	add("使役被动形", CausativePassiveForm(word, verbType))

	return res
}
