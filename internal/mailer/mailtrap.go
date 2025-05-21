package mailer

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"log"
	"time"

	gomail "gopkg.in/mail.v2"
)

type mailtrapClient struct {
	fromEmail, apiKey string
}

func NewMailTrapClient(apikey, fromEmail string) (mailtrapClient, error) {
	if apikey == "" {
		return mailtrapClient{}, errors.New("api key is required")
	}

	return mailtrapClient{
		fromEmail: fromEmail,
		apiKey:    apikey,
	}, nil
}

func (m mailtrapClient) Send(templateFile, username, email string, data any, isSandbox bool) (int, error) {
	// template parsing and building
	subject, body, err := buildTemplate(templateFile, data)
	if err != nil {
		return -1, err
	}

	resp, err := sendMail(m.fromEmail, email, subject, body, m.apiKey)
	return resp, err
}

func buildTemplate(templateFile string, data any) (string, string, error) {
	tmpl, err := template.ParseFS(FS, "templates/"+templateFile)
	if err != nil {
		return "", "", err
	}

	subject := new(bytes.Buffer)
	if err := tmpl.ExecuteTemplate(subject, "subject", data); err != nil {
		return "", "", err
	}

	body := new(bytes.Buffer)
	if err := tmpl.ExecuteTemplate(body, "body", data); err != nil {
		return "", "", err

	}

	return subject.String(), body.String(), nil
}

func sendMail(fromEmail, toEmail, subject, body, apiKey string) (int, error) {
	message := gomail.NewMessage()
	message.SetHeader("From", fromEmail)
	message.SetHeader("To", toEmail)
	message.SetHeader("Subject", subject)
	message.AddAlternative("text/html", body)
	dialer := gomail.NewDialer("smtp-relay.brevo.com", 587, "8d6a49001@smtp-brevo.com", "dDrN4xg58mSGnyPM")
	for i := 0; i < maxRetries; i++ {
		if err := dialer.DialAndSend(message); err != nil {
			log.Printf("failed to send email to %v, attempt %d of %d", toEmail, i+1, maxRetries)
			log.Printf("Error : %v", err.Error())

			// Exponential backoff
			time.Sleep(time.Second * time.Duration(i+1))
			continue
		}

		log.Printf("Email sent")
		return 200, nil
	}

	ErrorMessage := fmt.Errorf("failed to send email after %d attemps", maxRetries)
	return -1, errors.New(ErrorMessage.Error())
}
