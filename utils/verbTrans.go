package utils

import (
	"errors"
	"github.com/ikawaha/kagome-dict/ipa"
	"github.com/ikawaha/kagome/v2/tokenizer"
	"strings"
)

// CommandForm 动词命令形
func CommandForm(word string, verbType string) string {
	var uToE = map[string]string{
		"う": "え",
		"く": "け",
		"す": "せ",
		"つ": "て",
		"ぬ": "ね",
		"ふ": "へ",
		"む": "め",
		"ゆ": "え",
		"る": "れ",
		"ぐ": "げ",
		"ず": "ぜ",
		"ぶ": "べ",
		"ぷ": "ぺ",
		"づ": "で",
	}
	var res string
	if verbType == "一段" {
		res = replaceLast(word, "る", "ろ")
	} else if verbType == "五段" {
		lastChar := string([]rune(word)[len([]rune(word))-1])
		res = replaceLast(word, lastChar, uToE[lastChar])
	} else if verbType == "カ変" {
		res = replaceLast(word, "来る", "来い")
	} else if verbType == "サ変" {
		res = replaceLast(word, "する", "しろ")
	}
	return res
}

// 动词た形
func TaJointForm(word string, verbType string) string {
	var uToE = map[string]string{
		"う": "った",
		"く": "いた",
		"す": "した",
		"つ": "った",
		"ぬ": "んだ",
		"む": "んだ",
		"る": "った",
		"ぐ": "いだ",
		"ぶ": "んだ",
	}
	var res string
	if verbType == "一段" {
		res = replaceLast(word, "る", "た")
	} else if verbType == "五段" {
		lastChar := string([]rune(word)[len([]rune(word))-1])
		res = replaceLast(word, lastChar, uToE[lastChar])
	} else if verbType == "カ変" {
		res = replaceLast(word, "来る", "来た")
	} else if verbType == "サ変" {
		res = replaceLast(word, "する", "した")
	}
	return res
}

// 动词て形
func TeJointForm(word string, verbType string) string {
	var uToE = map[string]string{
		"う": "って",
		"く": "いて",
		"す": "して",
		"つ": "って",
		"ぬ": "んで",
		"む": "んで",
		"る": "って",
		"ぐ": "いで",
		"ぶ": "んで",
	}
	var res string
	if verbType == "一段" {
		res = replaceLast(word, "る", "て")
	} else if verbType == "五段" {
		lastChar := string([]rune(word)[len([]rune(word))-1])
		res = replaceLast(word, lastChar, uToE[lastChar])
	} else if verbType == "カ変" {
		res = replaceLast(word, "来る", "来て")
	} else if verbType == "サ変" {
		res = replaceLast(word, "する", "して")
	}
	return res
}

// 动词ます形
func JointForm(word string, verbType string) string {
	var uToI = map[string]string{
		"う": "い",
		"く": "き",
		"す": "し",
		"つ": "ち",
		"ぬ": "に",
		"ふ": "ひ",
		"む": "み",
		"ゆ": "い",
		"る": "り",
		"ぐ": "ぎ",
		"ず": "じ",
		"ぶ": "び",
		"ぷ": "ぴ",
		"づ": "ぢ",
	}
	var res string
	if verbType == "一段" {
		res = replaceLast(word, "る", "ます")
	} else if verbType == "五段" {
		lastChar := string([]rune(word)[len([]rune(word))-1])
		res = replaceLast(word, lastChar, uToI[lastChar]) + "ます"
	} else if verbType == "カ変" {
		res = replaceLast(word, "来る", "来ます")
	} else if verbType == "サ変" {
		res = replaceLast(word, "する", "します")
	}
	return res
}
func isVerb(str string) (string, error) {
	verbType := []string{"サ変", "カ変", "一段", "五段"}
	for _, verb := range verbType {
		if strings.Contains(str, verb) {
			return verb, nil
		}
	}
	return "", errors.New("words are not verbs")
}
func replaceLast(original, old, new string) string {
	lastIndex := strings.LastIndex(original, old)
	if lastIndex == -1 {
		return original
	}
	return original[:lastIndex] + new + original[lastIndex+len(old):]
}

// 动词ない形
func NegationForm(word string, verbType string) string {
	var uToA = map[string]string{
		"う": "わ",
		"く": "か",
		"す": "さ",
		"つ": "た",
		"ぬ": "な",
		"ふ": "は",
		"む": "ま",
		"ゆ": "や",
		"る": "ら",
		"ぐ": "が",
		"ず": "ざ",
		"ぶ": "ば",
		"ぷ": "ぱ",
		"づ": "だ",
	}
	var res string
	if verbType == "一段" {
		res = replaceLast(word, "る", "ない")
	} else if verbType == "五段" {
		lastChar := string([]rune(word)[len([]rune(word))-1])
		res = replaceLast(word, lastChar, uToA[lastChar]) + "ない"
	} else if verbType == "カ変" {
		res = replaceLast(word, "来る", "来ない")
	} else if verbType == "サ変" {
		res = replaceLast(word, "する", "しない")
	}
	return res
}

type wordForm struct {
	BaseForm   string
	OriginForm string
}
type TransRes struct {
	Category string `json:"category"`
	Result   string `json:"result"`
}

func VerbTransfiguration(str string) []TransRes {
	t, err := tokenizer.New(ipa.Dict(), tokenizer.OmitBosEos())
	if err != nil {
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
		verb, err := isVerb(strings.Join(r.Features, " "))
		if err != nil {
			return nil
		}
		verbType = verb
		wordform := wordForm{
			BaseForm:   r.BaseForm,
			OriginForm: r.Surface,
		}
		words = append(words, wordform)
	}
	word := ""
	for i, v := range words {
		if len(words)-1 == i {
			word += v.BaseForm
		} else {
			word += v.OriginForm
		}
	}
	if verbType == "サ変" {
		if !strings.Contains(word, "する") {
			word = word + "する"
		}
	}
	res := make([]TransRes, 0)
	res = append(res, TransRes{
		Category: "基本形",
		Result:   word,
	})
	res = append(res, TransRes{
		Category: "ます形",
		Result:   JointForm(word, verbType),
	})
	res = append(res, TransRes{
		Category: "ない形",
		Result:   NegationForm(word, verbType),
	})
	res = append(res, TransRes{
		Category: "て形",
		Result:   TeJointForm(word, verbType),
	})
	res = append(res, TransRes{
		Category: "た形",
		Result:   TaJointForm(word, verbType),
	})
	res = append(res, TransRes{
		Category: "命令形",
		Result:   CommandForm(word, verbType),
	})
	res = append(res, TransRes{
		Category: "可能形",
		Result:   PossibleForm(word, verbType),
	})
	res = append(res, TransRes{
		Category: "假定形",
		Result:   AssumingForm(word, verbType),
	})
	res = append(res, TransRes{
		Category: "意向形",
		Result:   IntentionalForm(word, verbType),
	})
	res = append(res, TransRes{
		Category: "被动形",
		Result:   PassiveForm(word, verbType),
	})
	res = append(res, TransRes{
		Category: "使役形",
		Result:   ServiceForm(word, verbType),
	})
	res = append(res, TransRes{
		Category: "禁止形",
		Result:   word + "な",
	})
	res = append(res, TransRes{
		Category: "使役被动形",
		Result:   PassiveServiceForm(word, verbType),
	})
	return res
}

// 动词可能形
func PossibleForm(word string, verbType string) string {
	var uToE = map[string]string{
		"う": "え",
		"く": "け",
		"す": "せ",
		"つ": "て",
		"ぬ": "ね",
		"ふ": "へ",
		"む": "め",
		"ゆ": "え",
		"る": "れ",
		"ぐ": "げ",
		"ず": "ぜ",
		"ぶ": "べ",
		"ぷ": "ぺ",
		"づ": "で",
	}
	var res string
	if verbType == "一段" {
		res = replaceLast(word, "る", "られる")
	} else if verbType == "五段" {
		lastChar := string([]rune(word)[len([]rune(word))-1])
		res = replaceLast(word, lastChar, uToE[lastChar]) + "る"
	} else if verbType == "カ変" {
		res = replaceLast(word, "来る", "来られる")
	} else if verbType == "サ変" {
		res = replaceLast(word, "する", "できる")
	}
	return res
}

// 假定形
func AssumingForm(word string, verbType string) string {
	var uToE = map[string]string{
		"う": "え",
		"く": "け",
		"す": "せ",
		"つ": "て",
		"ぬ": "ね",
		"ふ": "へ",
		"む": "め",
		"ゆ": "え",
		"る": "れ",
		"ぐ": "げ",
		"ず": "ぜ",
		"ぶ": "べ",
		"ぷ": "ぺ",
		"づ": "で",
	}
	var res string
	if verbType == "一段" {
		res = replaceLast(word, "る", "れば")
	} else if verbType == "五段" {
		lastChar := string([]rune(word)[len([]rune(word))-1])
		res = replaceLast(word, lastChar, uToE[lastChar]) + "ば"
	} else if verbType == "カ変" {
		res = replaceLast(word, "来る", "来れば")
	} else if verbType == "サ変" {
		res = replaceLast(word, "する", "すれば")
	}
	return res
}

// 意向形
func IntentionalForm(word string, verbType string) string {
	var uToO = map[string]string{
		"う": "お",
		"く": "こ",
		"す": "そ",
		"つ": "と",
		"ぬ": "の",
		"ふ": "ほ",
		"む": "も",
		"ゆ": "よ",
		"る": "ろ",
		"ぐ": "ご",
		"ず": "ぞ",
		"ぶ": "ぼ",
		"ぷ": "ぽ",
		"づ": "ど",
	}
	var res string
	if verbType == "一段" {
		res = replaceLast(word, "る", "よう")
	} else if verbType == "五段" {
		lastChar := string([]rune(word)[len([]rune(word))-1])
		res = replaceLast(word, lastChar, uToO[lastChar]) + "う"
	} else if verbType == "カ変" {
		res = replaceLast(word, "来る", "来よう")
	} else if verbType == "サ変" {
		res = replaceLast(word, "する", "しよう")
	}
	return res
}

// 被动形
func PassiveForm(word string, verbType string) string {
	var uToO = map[string]string{
		"う": "お",
		"く": "こ",
		"す": "そ",
		"つ": "と",
		"ぬ": "の",
		"ふ": "ほ",
		"む": "も",
		"ゆ": "よ",
		"る": "ろ",
		"ぐ": "ご",
		"ず": "ぞ",
		"ぶ": "ぼ",
		"ぷ": "ぽ",
		"づ": "ど",
	}
	var res string
	if verbType == "一段" {
		res = replaceLast(word, "る", "られる")
	} else if verbType == "五段" {
		lastChar := string([]rune(word)[len([]rune(word))-1])
		res = replaceLast(word, lastChar, uToO[lastChar]) + "れる"
	} else if verbType == "カ変" {
		res = replaceLast(word, "来る", "来られる")
	} else if verbType == "サ変" {
		res = replaceLast(word, "する", "される")
	}
	return res
}

// 使役形
func ServiceForm(word string, verbType string) string {
	var uToO = map[string]string{
		"う": "お",
		"く": "こ",
		"す": "そ",
		"つ": "と",
		"ぬ": "の",
		"ふ": "ほ",
		"む": "も",
		"ゆ": "よ",
		"る": "ろ",
		"ぐ": "ご",
		"ず": "ぞ",
		"ぶ": "ぼ",
		"ぷ": "ぽ",
		"づ": "ど",
	}
	var res string
	if verbType == "一段" {
		res = replaceLast(word, "る", "させる")
	} else if verbType == "五段" {
		lastChar := string([]rune(word)[len([]rune(word))-1])
		res = replaceLast(word, lastChar, uToO[lastChar]) + "せる"
	} else if verbType == "カ変" {
		res = replaceLast(word, "来る", "来させる")
	} else if verbType == "サ変" {
		res = replaceLast(word, "する", "させる")
	}
	return res
}

// 使役+被动形
func PassiveServiceForm(word string, verbType string) string {
	var uToO = map[string]string{
		"う": "お",
		"く": "こ",
		"す": "そ",
		"つ": "と",
		"ぬ": "の",
		"ふ": "ほ",
		"む": "も",
		"ゆ": "よ",
		"る": "ろ",
		"ぐ": "ご",
		"ず": "ぞ",
		"ぶ": "ぼ",
		"ぷ": "ぽ",
		"づ": "ど",
	}
	var res string
	if verbType == "一段" {
		res = replaceLast(word, "る", "させられる")
	} else if verbType == "五段" {
		lastChar := string([]rune(word)[len([]rune(word))-1])
		res = replaceLast(word, lastChar, uToO[lastChar]) + "される"
	} else if verbType == "カ変" {
		res = replaceLast(word, "来る", "来させられる")
	} else if verbType == "サ変" {
		res = replaceLast(word, "する", "させられる")
	}
	return res
}
func main() {
	VerbTransfiguration("連れる")
	//log.Println(string(bytes.Join(array, []byte(",\n"))))
}
