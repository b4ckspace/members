package mailer

import (
	"crypto/tls"
	"fmt"
	"net/smtp"
	"text/template"
	"io/ioutil"

	"github.com/b4ckspace/members/internal/core"
	"github.com/b4ckspace/members/internal/statics"
)

type (
	Mailer struct {
		connFactory ConnFactory
	}
	welcomeMail struct {
		Nickname string
		Token    string
	}

	ConnFactory func() (core.SmtpConn, error)
)

func New(connFactory ConnFactory) (m *Mailer) {
	return &Mailer{connFactory: connFactory}
}

func SmtpConnFactory(host string) ConnFactory {
	return func() (core.SmtpConn, error) {
		return smtp.Dial(host)
	}

}

func (m *Mailer) SendPassword(to, nickname, token string) (err error) {
	c, err := m.connFactory()
	if err != nil {
		return fmt.Errorf("unable to open smtp connection: %s", err)
	}
	defer c.Close()
	err = c.StartTLS(&tls.Config{
		ServerName: "mail.hackerspace-bamberg.de",
	})
	if err != nil {
		return fmt.Errorf("unable to upgrade to tls: %s", err)
	}
	err = c.Mail("register@hackerspace-bamberg.de")
	if err != nil {
		return fmt.Errorf("unable to set sender: %s", err)
	}
	err = c.Rcpt(to)
	if err != nil {
		return fmt.Errorf("unable to set rcpt: %s", err)
	}
	body, err := c.Data()
	if err != nil {
		return fmt.Errorf("unable to send mail: %s", err)
	}
	defer body.Close()

	fp , err:= statics.MustStatics().Open("/templates/email.txt")
	if err != nil {
		return fmt.Errorf("unable to open mail template: %s", err)
	}
	templateBody, err := ioutil.ReadAll(fp)
	if err != nil {
		return fmt.Errorf("unable to load mail template: %s", err)
	}
	t, err := template.New("email.txt").Parse(string(templateBody))
	if err != nil {
		return fmt.Errorf("unable to parse mail template: %s", err)
	}

	err = t.Execute(body, welcomeMail{
		Nickname: nickname,
		Token:    token,
	})
	if err != nil {
		return fmt.Errorf("unable to send mail: %s", err)
	}
	return
}
