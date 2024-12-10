package utils

import (
	"crypto/rand"
	"crypto/sha256"
	"easyjapanese/config"
	"encoding/hex"
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

func GetToken() {

}
