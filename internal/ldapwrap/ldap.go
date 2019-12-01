package ldapwrap

import (
	"crypto/tls"
	"fmt"

	"gopkg.in/ldap.v3"

	"github.com/b4ckspace/members/internal/core"
)

type LdapConnFactory func() (conn core.LdapConn, err error)

func NewLdapConnFactory(
	host string, port int, username, password string,
) (
	ldapConnFactory LdapConnFactory,
) {
	return func() (conn core.LdapConn, err error) {
		c, err := ldap.Dial("tcp", fmt.Sprintf("%s:%d", host, port))
		if err != nil {
			return nil, fmt.Errorf("unable to connect to ldap: %s", err)
		}
		err = c.StartTLS(&tls.Config{
			ServerName: host,
		})
		if err != nil {
			return nil, fmt.Errorf("unable to switch to tls: %s", err)
		}
		err = c.Bind(username, password)
		if err != nil {
			return nil, fmt.Errorf("unable to login to ldap: %s", err)
		}
		return c, nil
	}
}
