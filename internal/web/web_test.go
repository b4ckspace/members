package web

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/golang/mock/gomock"

	"github.com/b4ckspace/members/mocks"
)

func TestWeb(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockMailer := mocks.NewMockMailer(mockCtrl)
	mockLdapDailer := mocks.NewMockLdapDialer(mockCtrl)
	mockLdapWrap := mocks.NewMockLdapWrap(mockCtrl)

	_ = os.Chdir("../../")

	web, err := New(mockMailer, mockLdapDailer)
	if err != nil {
		t.Fatalf("unable to create web: %s", err)
	}

	// newuser test
	registerMemberOpts := []struct {
		testName     string
		nickname     string
		email        string
		mlMailOption string
		mlMail       string
		token        string
		err          error
		mailerErr    error
		want         string
	}{{
		"valid",
		"member", "member@email.local", "space", "member@hackerspace-bamberg.de",
		"token", nil,
		nil,
		"Willkommen im Backspace!",
	}, {
		"ldap error",
		"member", "member@email.local", "space", "member@hackerspace-bamberg.de",
		"token", fmt.Errorf("error 1337"),
		nil,
		"Member konnte nicht angelegt werden",
	}, {
		"mailer err",
		"member", "member@email.local", "space", "member@hackerspace-bamberg.de",
		"token", nil,
		fmt.Errorf("unable to send mail"),
		"Registrierungs-Mail konnte nicht gesendet werden",
	}}
	for _, o := range registerMemberOpts {
		t.Logf("running %s", o.testName)
		mockLdapDailer.EXPECT().Dial(context.Background()).Return(mockLdapWrap, nil)
		mockLdapWrap.EXPECT().MemberExists(o.nickname).Return(false, nil)
		mockLdapWrap.EXPECT().
			RegisterMember(o.nickname, o.email, o.mlMail).
			Return(o.token, o.err)
		if o.err == nil {
			mockMailer.EXPECT().
				SendPassword(o.email, o.nickname, o.token).
				Return(o.mailerErr)
		}

		r := bytes.NewBufferString(fmt.Sprintf(
			"nickname=%s&email=%s&mladdr=%s",
			o.nickname,
			o.email,
			o.mlMailOption,
		))
		ok, err := postOk(web, "/register", r, o.want)
		if err != nil {
			t.Logf("test: %s", o.testName)
			t.Fatalf("unable to post: %s\n%s", err, o.want)
		}
		if !ok {
			t.Logf("test: %s", o.testName)
			t.Fatalf("invalid response, missing: '%s'", o.want)
		}
	}

	changePasswordOpts := []struct {
		testName string
		token    string
		password string
		doorpass string
		want     string
	}{{
		"update password",
		"t0k3n",
		"p4ssw0rd",
		"p4ssw0rd",
		"Passwort wurde aktualisiert",
	}}
	for _, o := range changePasswordOpts {
		mockLdapDailer.EXPECT().Dial(context.Background()).Return(mockLdapWrap, nil)
		mockLdapWrap.EXPECT().SetPassword(o.token, o.password, o.doorpass)
		url := fmt.Sprintf("/password?t=%s", o.token)
		r := bytes.NewBufferString(fmt.Sprintf(
			"password=%s&password2=%s&doorpass=%s&doorpass2=%s",
			o.password, o.password, o.doorpass, o.doorpass,
		))
		ok, err := postOk(web, url, r, o.want)
		if err != nil {
			t.Fatalf("unable to post: %s", err)
		}
		if !ok {
			t.Fatalf("invalid response, missing: '%s'", o.want)
		}
	}
}

func postOk(web *Web, url string, r io.Reader, want string) (ok bool, err error) {
	req, err := http.NewRequestWithContext(
		context.Background(),
		"POST",
		url,
		r,
	)
	if err != nil {
		return false, fmt.Errorf("unable to post /register: %s", err)
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()
	web.GetMux().ServeHTTP(rr, req)
	body, _ := ioutil.ReadAll(rr.Result().Body)
	ok = bytes.Contains(body, []byte(want))
	if !ok {
		return ok, fmt.Errorf(string(body))
	}
	return
}
