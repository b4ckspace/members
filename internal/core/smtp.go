//go:generate mockgen -source=$GOFILE -destination=$PWD/mocks/${GOFILE} -package=mocks
package core

import (
	"crypto/tls"
	"io"
)

type (
	SmtpConn interface {
		Data() (io.WriteCloser, error)
		Mail(string) error
		Rcpt(string) error
		StartTLS(*tls.Config) error
		Close() error
	}
)
