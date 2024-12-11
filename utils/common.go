package utils

import (
	"crypto/rand"
	"crypto/sha256"
	"easyjapanese/config"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
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
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
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
