package web

import (
	"errors"
	"fmt"
	"net/http"
	"regexp"
)

type (
	RegisterForm struct {
		Nickname string
		EMail    string
		MlAddr   string
		Error    string
		ErrorMsg string
	}
	ResetForm struct {
		Nickname string
		Error    string
		ErrorMsg string
	}
	PasswordForm struct {
		Password  string
		Password2 string
		Doorpass  string
		Doorpass2 string
		Error     string
		ErrorMsg  string
	}
)

var nickValid = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_-]*[a-zA-Z0-9]$`)
var mailValid = regexp.MustCompile("^[a-zA-Z0-9.!#$%&’*+/=?^_`{|}~-]+@[a-zA-Z0-9-]+(?:\\.[a-zA-Z0-9-]+)*$")

func parseRegisterForm(r *http.Request) (f *RegisterForm, posted bool, err error) {
	if r.Method != "POST" {
		return &RegisterForm{}, false, nil
	}
	posted = true

	f = &RegisterForm{
		Nickname: r.PostFormValue("nickname"),
		EMail:    r.PostFormValue("email"),
		MlAddr:   r.PostFormValue("mladdr"),
	}
	if len(f.Nickname) < 2 || f.Nickname == "penis" {
		err = fmt.Errorf("%s is to short", f.Nickname)
		f.Error = "nickname"
		f.ErrorMsg = err.Error()
		return
	}
	if !nickValid.MatchString(f.Nickname) {
		err = errors.New("invalid nickname")
		f.Error = "nickname"
		f.ErrorMsg = err.Error()
		return
	}
	if !mailValid.MatchString(f.EMail) {
		err = errors.New("invalid email address")
		f.Error = "email"
		f.ErrorMsg = err.Error()
		return
	}

	switch f.MlAddr {
	case "own":
		f.MlAddr = f.EMail
	case "space":
		f.MlAddr = fmt.Sprintf("%s@hackerspace-bamberg.de", f.Nickname)
	default:
		err = errors.New("invalid ml address")
		f.Error = "mladdr"
		f.ErrorMsg = err.Error()
		return
	}
	return
}

func parseResetForm(r *http.Request) (f *ResetForm, posted bool, err error) {
	if r.Method != "POST" {
		return &ResetForm{}, false, nil
	}
	posted = true

	f = &ResetForm{
		Nickname: r.PostFormValue("nickname"),
	}
	if len(f.Nickname) < 2 || f.Nickname == "penis" {
		err = fmt.Errorf("%s is to short", f.Nickname)
		f.Error = "nickname"
		f.ErrorMsg = err.Error()
		return
	}
	return
}

func parsePasswordForm(r *http.Request) (f *PasswordForm, posted bool, err error) {
	if r.Method != "POST" {
		return &PasswordForm{}, false, nil
	}
	posted = true

	f = &PasswordForm{
		Password:  r.PostFormValue("password"),
		Password2: r.PostFormValue("password2"),
		Doorpass:  r.PostFormValue("doorpass"),
		Doorpass2: r.PostFormValue("doorpass2"),
	}
	if len(f.Password) < 8 {
		err = errors.New("password to short, needs more than eight characters")
		f.Error = "password"
		f.ErrorMsg = err.Error()
		return
	}
	if f.Password != f.Password2 {
		err = errors.New("passwords do not match")
		f.Error = "password"
		f.ErrorMsg = err.Error()
		return
	}

	if len(f.Doorpass) < 8 {
		err = errors.New("door password to short, needs more than eight characters")
		f.Error = "doorpass"
		f.ErrorMsg = err.Error()
		return
	}
	if f.Doorpass != f.Doorpass2 {
		err = errors.New("door passwords do not match")
		f.Error = "doorpass"
		f.ErrorMsg = err.Error()
		return
	}
	return
}
