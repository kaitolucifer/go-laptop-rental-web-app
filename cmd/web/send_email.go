package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"strings"
	"time"

	"github.com/kaitolucifer/go-laptop-rental-site/internal/models"
	mail "github.com/xhit/go-simple-mail/v2"
)

// listenForMail() listens for app.MailChan and send mail
func listenForMail() {
	go func() {
		for m := range app.MailChan {
			sendMail(m)
		}
	}()
}

// sendMail sends mail
func sendMail(m models.MailData) {
	server := mail.NewSMTPClient()
	server.Host = "localhost" // mailhog smtp test host
	server.Port = 1025        // mailhog smtp test port
	server.KeepAlive = false
	server.ConnectTimeout = 10 * time.Second
	server.SendTimeout = 10 * time.Second
	// set varaible below when using production email service
	// server.Username
	// server.Password
	// server.Encryption

	client, err := server.Connect()
	if err != nil {
		app.ErrorLog.Println(err)
	}

	email := mail.NewMSG()
	email.SetFrom(m.From).AddTo(m.To).SetSubject(m.Subject)
	if m.Template == "" {
		email.SetBody(mail.TextHTML, m.Content)
	} else {
		data, err := ioutil.ReadFile(fmt.Sprintf("./templates/%s", m.Template))
		if err != nil {
			app.ErrorLog.Println(err)
		}

		mailTemplate := string(data)
		msgToSend := strings.Replace(mailTemplate, "[%body%]", m.Content, 1)
		email.SetBody(mail.TextHTML, msgToSend)
	}

	err = email.Send(client)
	if err != nil {
		log.Println(err)
	} else {
		log.Println("Email sent!")
	}
}
