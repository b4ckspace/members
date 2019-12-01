//go:generate mockgen -source=$GOFILE -destination=$PWD/mocks/${GOFILE} -package=mocks
package core

type (
	Mailer interface {
		SendPassword(to, nickname, token string) error
	}
)
