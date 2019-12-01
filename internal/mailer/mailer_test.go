package mailer

import (
	"bytes"
	"crypto/tls"
	"io"
	"os"
	"testing"

	"github.com/golang/mock/gomock"

	"github.com/b4ckspace/members/internal/core"
	"github.com/b4ckspace/members/mocks"
)

func TestSendPassword(t *testing.T) {
	_ = os.Chdir("../../")

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	c := mocks.NewMockSmtpConn(mockCtrl)
	m := New(func() (core.SmtpConn, error) { return c, nil })

	r, w := io.Pipe()
	mailBody := bytes.NewBuffer([]byte{})
	go func() {
		_, _ = io.Copy(mailBody, r)
	}()

	c.EXPECT().StartTLS(&tls.Config{
		ServerName: "mail.hackerspace-bamberg.de",
	})
	c.EXPECT().Mail("register@hackerspace-bamberg.de")
	c.EXPECT().Data().Return(w, nil)
	c.EXPECT().Rcpt("member@example.com")
	c.EXPECT().Close()

	err := m.SendPassword("member@example.com", "member", "t0k3n")
	if err != nil {
		t.Fatalf("unable to test mail: %s", err)
	}
	if !bytes.Contains(mailBody.Bytes(), []byte("member")) {
		t.Fatalf("nickname not in mail body")
	}
	if !bytes.Contains(mailBody.Bytes(), []byte("t0k3n")) {
		t.Fatalf("nickname not in mail body")
	}
}
