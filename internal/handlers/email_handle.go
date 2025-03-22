package handlers

import (
	"crypto/tls"
	"easyjapanese/config"
	"log"

	"gopkg.in/gomail.v2"
)

func SendEmail(title, content, to string) bool {
	m := gomail.NewMessage()
	m.SetHeader("From", config.EmailUser)
	m.SetHeader("To", to)
	m.SetHeader("Subject", title)
	m.SetBody("text/html", content)
	d := gomail.NewDialer(
		config.EmailHost,
		config.EmailPort,
		config.EmailUser,
		config.EmailPasswd,
	)
	d.TLSConfig = &tls.Config{InsecureSkipVerify: true}
	if err := d.DialAndSend(m); err != nil {
		log.Println("邮箱发送失败", err)
		return false
	}
	return true
}
