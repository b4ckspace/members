package web

import (
	"log"
	"net/http"
	"fmt"
)

func (web *Web) handlePassword(r *http.Request) (td *PasswordTemplateData) {
	qs := r.URL.Query()
	token := qs.Get("t")

	f, posted, err := parsePasswordForm(r)
	td = &PasswordTemplateData{
		Form:     f,
		Messages: []Message{},
	}
	if !posted {
		return
	}
	if err != nil {
		td.Messages = append(td.Messages, Message{DANGER, err.Error()})
		return
	}

	ldap, err := web.ldapDialer.Dial(r.Context())
	if err != nil {
		log.Printf("ldap error: %s", err)
		td.Messages = append(td.Messages, Message{
			DANGER,
			"Verbindung zum LDAP Server nicht möglich",
		})
		return
	}

	err = ldap.SetPassword(token, f.Password, f.Doorpass)
	if err != nil {
		log.Printf("ldap error: %s", err)
		td.Messages = append(td.Messages, Message{
			DANGER,
			"Passwort konnte nicht gesetzt werden",
		})
		return
	}

	td.Messages = append(
		td.Messages,
		Message{SUCCESS, "Passwort wurde aktualisiert"},
	)
	td.Form = &PasswordForm{}
	return
}

func (web *Web) handleRegister(r *http.Request) (td *RegisterTemplateData) {
	f, posted, err := parseRegisterForm(r)
	td = &RegisterTemplateData{
		Form:     f,
		Messages: []Message{},
	}
	if !posted {
		return
	}
	if err != nil {
		td.Messages = append(td.Messages, Message{DANGER, err.Error()})
		return
	}

	ldap, err := web.ldapDialer.Dial(r.Context())
	if err != nil {
		log.Printf("ldap error: %s", err)
		td.Messages = append(td.Messages, Message{
			DANGER,
			"Verbindung zum LDAP Server nicht möglich",
		})
		return
	}

	exists, err := ldap.MemberExists(td.Form.Nickname)
	if err != nil {
		log.Printf("ldap error: %s", err)
		td.Messages = append(td.Messages, Message{
			DANGER,
			"Nickname check fehlgeschlagen",
		})
		return
	}
	if exists {
		td.Messages = append(td.Messages, Message{
			DANGER,
			fmt.Sprintf(
				"Der Nickname \"%s\" ist bereits vergeben",
				td.Form.Nickname,
			),
		})
		return
	}

	token, err := ldap.RegisterMember(td.Form.Nickname, td.Form.EMail, td.Form.MlAddr)
	if err != nil {
		log.Printf("ldap error: %s", err)
		td.Messages = append(td.Messages, Message{
			DANGER,
			"Member konnte nicht angelegt werden",
		})
		return
	}
	td.Messages = append(td.Messages, Message{
		SUCCESS,
		"Registrierung erfolgreich. " +
			"Bitte klicke auf den Passwort Link in der soeben gesendeten Mail",
	},
	)

	err = web.mailer.SendPassword(td.Form.EMail, td.Form.Nickname, token)
	if err != nil {
		log.Printf("mail error: %s", err.Error())
		td.Messages = append(td.Messages, Message{
			WARNING,
			"Registrierungs-Mail konnte nicht gesendet werden",
		})
		return
	}
	td.Form = &RegisterForm{}
	return
}

func (web *Web) handleReset(r *http.Request) (td *ResetTemplateData) {
	f, posted, err := parseResetForm(r)
	td = &ResetTemplateData{
		Form: f,
	}
	if !posted {
		return
	}
	if err != nil {
		td.Messages = append(td.Messages, Message{DANGER, err.Error()})
		return
	}

	ldap, err := web.ldapDialer.Dial(r.Context())
	if err != nil {
		log.Printf("ldap error: %s", err)
		td.Messages = append(td.Messages, Message{
			DANGER,
			"Verbindung zum LDAP Server nicht möglich",
		})
		return
	}

	token, email, err := ldap.PasswordReset(f.Nickname)
	if err != nil {
		log.Printf("ldap error: %s", err)
		td.Messages = append(td.Messages, Message{
			DANGER,
			"Token zum Passwort zurücksetzen konnte nicht gesetzt werden",
		})
		return
	}

	err = web.mailer.SendPassword(email, f.Nickname, token)
	if err != nil {
		log.Printf("email error: %s", err)
		td.Messages = append(td.Messages, Message{
			DANGER,
			"Passwort Mail konnte nicht gesendet werden",
		})
		return
	}
	td.Messages = append(td.Messages, Message{SUCCESS, "Passwort Mail wurde gesendet"})
	td.Form = &ResetForm{}
	return
}
