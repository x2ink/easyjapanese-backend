package utils

import (
	"crypto/rand"
	"crypto/sha256"
	"easyjapanese/config"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

func EncryptionPassword(password string) string {
	buf := make([]byte, 0, len(password)+len(config.Salt))
	buf = append(buf, password...)
	buf = append(buf, config.Salt...)
	b := sha256.Sum256(buf)
	return hex.EncodeToString(b[:])
}

func GenerateRandomString(length int, rtype string) string {
	var charset string
	if rtype == "number" {
		charset = "0123456789"
	} else {
		charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	}
	b := make([]byte, length)
	_, err := rand.Read(b)
	if err != nil {
		panic(err)
	}
	result := make([]byte, length)
	for i := range b {
		result[i] = charset[int(b[i])%len(charset)]
	}
	return string(result)
}

type Token struct {
	RoleId uint
	UserId uint
}

func EncryptToken(data Token) string {
	jsonData, _ := json.Marshal(data)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * config.TokenExpireTime)),
		Issuer:    string(jsonData),
	})
	ss, _ := token.SignedString([]byte(config.Salt))
	return ss
}

func DecryptToken(data string) (Token, error) {
	var tokenData Token
	token, err := jwt.ParseWithClaims(data, jwt.MapClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(config.Salt), nil
	})
	if err != nil {
		return tokenData, err
	}
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		jsonData, _ := claims.GetIssuer()
		err = json.Unmarshal([]byte(jsonData), &tokenData)
		if err != nil {
			return tokenData, err
		}
		return tokenData, nil
	}
	return tokenData, fmt.Errorf("invalid token")
}

func GetIpAddress(ip string) (string, error) {
	type Res struct {
		IP          string `json:"ip"`
		Pro         string `json:"pro"`
		ProCode     string `json:"proCode"`
		City        string `json:"city"`
		CityCode    string `json:"cityCode"`
		Region      string `json:"region"`
		RegionCode  string `json:"regionCode"`
		Addr        string `json:"addr"`
		RegionNames string `json:"regionNames"`
		Err         string `json:"err"`
	}

	// 请求 URL
	url := fmt.Sprintf("https://whois.pconline.com.cn/ipJson.jsp?ip=%s&json=true", ip)
	res, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("Acquisition failed: %v", err)
	}
	defer res.Body.Close()

	// 读取响应体
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", fmt.Errorf("Acquisition failed: %v", err)
	}

	// 如果响应内容使用的是 GBK 编码，可以进行转换
	reader := transform.NewReader(strings.NewReader(string(body)), simplifiedchinese.GBK.NewDecoder())
	convertedBody, err := ioutil.ReadAll(reader)
	if err != nil {
		return "", fmt.Errorf("Failed to convert body: %v", err)
	}

	// 解析 JSON 数据
	var resInfo Res
	if err := json.Unmarshal(convertedBody, &resInfo); err != nil {
		return "", fmt.Errorf("Acquisition failed: %v", err)
	}

	return resInfo.City, nil
}
