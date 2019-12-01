package web

import (
	"bytes"
	"errors"
	"fmt"
	"net/http/httptest"
	"testing"
)

func TestParseRegisterForm(t *testing.T) {
	registerFormData := []struct {
		method string

		nickname  string
		email     string
		mlAddr    string
		resMlAddr string

		posted bool
		err    error
	}{
		{"GET", "", "", "", "", false, nil},
		{"POST", "m", "member@space.local", "space", "member@hackerspace-bamberg.de", true, errors.New("m is to short")},
		{"POST", "-_-", "member@space.local", "space", "member@hackerspace-bamberg.de", true, errors.New("invalid nickname")},
		{"POST", "member", "member-space.local", "space", "member@hackerspace-bamberg.de", true, errors.New("invalid email address")},
		{"POST", "member", "member@space.local", "own", "member@space.local", true, nil},
		{"POST", "member", "member@space.local", "invalid", "", true, errors.New("invalid ml address")},
	}
	for _, d := range registerFormData {
		body := bytes.NewBufferString(fmt.Sprintf(
			"nickname=%s&email=%s&mladdr=%s",
			d.nickname, d.email, d.mlAddr,
		))
		r := httptest.NewRequest(d.method, "/", body)
		r.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		f, posted, err := parseRegisterForm(r)
		if posted != d.posted {
			t.Fatalf("posted is not detected")
		}
		if (err != nil || d.err != nil) && (err != nil && d.err == nil ||
			d.err != nil && err == nil ||
			err.Error() != d.err.Error()) {
			t.Fatalf("mismatching error:\n  %s\nvs\n  %s", err, d.err)
		}
		if err != nil {
			continue
		}
		if f.Nickname != d.nickname {
			t.Fatalf("invalid nickname: %s %s", f.Nickname, d.nickname)
		}
		if f.EMail != d.email {
			t.Fatalf("invalid email: %s %s", f.EMail, d.email)
		}
		if f.MlAddr != d.resMlAddr {
			t.Logf("req: %+v", d)
			t.Logf("res: %+v", f)
			t.Fatalf("invalid mladdr: %s %s", f.MlAddr, d.resMlAddr)
		}

	}
}

func TestParsePasswordForm(t *testing.T) {
	passwordFormData := []struct {
		method string

		password  string
		password2 string
		doorpass  string
		doorpass2 string

		posted bool
		err    error
	}{
		{"GET", "", "", "", "", false, nil},
		{"POST", "p4ssw0rd", "p4ssw0rd", "p4ssw1rd", "p4ssw1rd", true, nil},
		{"POST", "p4ss", "p4ss", "p4ssw1rd", "p4ssw1rd", true, errors.New("password to short, needs more than eight characters")},
		{"POST", "p4ssw0rd", "p4ssw0rdx", "p4ssw1rd", "p4ssw1rd", true, errors.New("passwords do not match")},
		{"POST", "p4ssw0rd", "p4ssw0rd", "p4ss", "p4s", true, errors.New("door password to short, needs more than eight characters")},
		{"POST", "p4ssw0rd", "p4ssw0rd", "p4ssw1rd", "p4ssw1rdx", true, errors.New("door passwords do not match")},
	}
	for _, d := range passwordFormData {
		body := bytes.NewBufferString(fmt.Sprintf(
			"password=%s&password2=%s&doorpass=%s&doorpass2=%s",
			d.password, d.password2, d.doorpass, d.doorpass2,
		))
		r := httptest.NewRequest(d.method, "/", body)
		r.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		f, posted, err := parsePasswordForm(r)
		if posted != d.posted {
			t.Fatalf("posted is not detected")
		}
		if (err != nil || d.err != nil) && (err != nil && d.err == nil ||
			d.err != nil && err == nil ||
			err.Error() != d.err.Error()) {
			t.Fatalf("mismatching error:\n  %s\nvs\n  %s", err, d.err)
		}
		if f.Password != d.password {
			t.Fatalf("mismatching password: '%s' '%s'", f.Password, d.password)
		}
		if f.Password2 != d.password2 {
			t.Fatalf("mismatching password2: '%s' '%s'", f.Password2, d.password2)
		}
		if f.Doorpass != d.doorpass {
			t.Fatalf("mismatching doorpass: '%s' '%s'", f.Doorpass, d.doorpass)
		}
		if f.Doorpass2 != d.doorpass2 {
			t.Fatalf("mismatching doorpass2: '%s' '%s'", f.Doorpass2, d.doorpass2)
		}

	}
}
