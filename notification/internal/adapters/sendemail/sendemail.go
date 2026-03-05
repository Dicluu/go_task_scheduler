package sendemail

import (
	"crypto/tls"
	"notification/internal/config"

	"gopkg.in/mail.v2"
)

type MessageEmailInterface interface {
	SendEmailNotify(to, subject, body string) error
}

type emailAttribute struct {
	Username string
	Password string
	Host     string
	Port     int
	From     string
	IsTls    bool
}

type Adapter struct {
	message MessageEmailInterface
}

func New(cfg config.SmtpServer) *Adapter {
	return &Adapter{
		message: newMessageEmail(cfg),
	}
}

func (a *Adapter) Send(to, subject, body string) error {
	err := a.message.SendEmailNotify(to, subject, body)

	return err
}

func (e emailAttribute) SendEmailNotify(to, subject, body string) error {
	m := mail.NewMessage()
	m.SetHeader("From", e.From)
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", body)

	d := mail.NewDialer(e.Host, e.Port, e.Username, e.Password)
	d.TLSConfig = &tls.Config{
		InsecureSkipVerify: true, // not for production
	}

	if err := d.DialAndSend(m); err != nil {
		return err
	}

	return nil
}

func newMessageEmail(cfg config.SmtpServer) MessageEmailInterface {
	return &emailAttribute{
		Username: cfg.Username,
		Password: cfg.Password,
		Host:     cfg.Host,
		Port:     cfg.Port,
		From:     cfg.From,
		IsTls:    cfg.IsTls,
	}
}
