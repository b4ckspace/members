//go:generate mockgen -source=$GOFILE -destination=$PWD/mocks/${GOFILE} -package=mocks
package core

import (
	"context"

	"github.com/go-ldap/ldap/v3"
)

type (
	LdapConn interface {
		Add(*ldap.AddRequest) error
		Modify(*ldap.ModifyRequest) error
		Search(*ldap.SearchRequest) (*ldap.SearchResult, error)
		Close() error
	}
	LdapDialer interface {
		Dial(context.Context) (LdapWrap, error)
	}
	LdapWrap interface {
		RegisterMember(user, email, mlEmail string) (token string, err error)
		SetPassword(token, password, doorpass string) error
		MemberExists(uid string) (exists bool, err error)
		PasswordReset(nickname string) (token, email string, err error)
	}
)
