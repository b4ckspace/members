package ldapwrap

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"sync"
	"time"

	"gopkg.in/ldap.v3"

	"github.com/b4ckspace/members/internal/core"
	"github.com/b4ckspace/members/internal/ssha"
)

type (
	LdapDialer struct {
		connFactory LdapConnFactory
	}
	LdapWrap struct {
		conn core.LdapConn
		m    sync.Mutex
	}

	Token struct {
		ValidUntil time.Time
		Nickname   string
		Random     []byte
	}
)

func EscapeFilter(f string) string {
	return ldap.EscapeFilter(f)
}

func New(cf LdapConnFactory) (l core.LdapDialer, err error) {
	return &LdapDialer{
		connFactory: cf,
	}, err
}

func (ld *LdapDialer) Dial(ctx context.Context) (core.LdapWrap, error) {
	c, err := ld.connFactory()
	if err != nil {
		return nil, fmt.Errorf("unable to connect: %s", err)
	}

	go func() {
		<-ctx.Done()
		c.Close()
	}()

	return &LdapWrap{
		conn: c,
	}, nil
}

func (l *LdapWrap) NextUidNumber() (nextUidNumber int, err error) {
	l.m.Lock()
	defer l.m.Unlock()
	res, err := l.SearchActiveAndInactive("(objectClass=backspaceMember)", []string{"uidNumber"})
	if err != nil {
		return 0, fmt.Errorf("unable to query for new uid: %s", err)
	}
	for _, member := range res.Entries {
		for _, attr := range member.Attributes {
			if attr.Name == "uidNumber" {
				if len(attr.Values) == 1 {
					uidNumber, err := strconv.Atoi(attr.Values[0])
					if err != nil {
						continue
					}
					if uidNumber > nextUidNumber {
						nextUidNumber = uidNumber
					}
				}
			}
		}
	}
	nextUidNumber++
	return
}

func (l *LdapWrap) PasswordReset(nickname string) (token, email string, err error) {
	nickname = ldap.EscapeFilter(nickname)
	search := fmt.Sprintf("(&(objectClass=backspaceMember)(uid=%s))", nickname)
	sr, err := l.SearchActive(search, []string{"alternateEmail"})
	if err != nil {
		return "", "", fmt.Errorf("unable to find member: %s", err)
	}
	if len(sr.Entries) != 1 {
		return "", "", fmt.Errorf("unable to find user with nickname: %s", nickname)
	}

	token, err = GenerateToken(nickname)
	if err != nil {
		return "", "", fmt.Errorf("unable to generate token: %s", err)
	}

	member := sr.Entries[0]
	email = member.GetAttributeValue("alternateEmail")
	req := ldap.NewModifyRequest(member.DN, []ldap.Control{})
	req.Replace("token", []string{token})

	err = l.conn.Modify(req)
	if err != nil {
		return "", "", fmt.Errorf("unable to set token: %s", err)
	}
	return token, email, nil
}

func (l *LdapWrap) RegisterMember(user, email, mlEmail string) (token string, err error) {
	exists, err := l.MemberExists(user)
	if err != nil {
		return "", fmt.Errorf("unable to check if user exists: %s", err)
	}
	if exists {
		return "", fmt.Errorf("unable to add user: user %s exists", user)
	}
	uidNumber, err := l.NextUidNumber()
	if err != nil {
		return "", err
	}
	token, err = GenerateToken(user)
	if err != nil {
		return "", fmt.Errorf("unable to generate token: %s", err)
	}
	intEmail := fmt.Sprintf("%s@hackerspace-bamberg.de", user)

	dn := fmt.Sprintf("uid=%s,ou=inactiveMember,dc=backspace", user)
	req := ldap.NewAddRequest(dn, []ldap.Control{})
	req.Attribute("objectClass", []string{"backspaceMember"})
	req.Attribute("uid", []string{user})
	req.Attribute("uidNumber", []string{fmt.Sprintf("%d", uidNumber)})
	req.Attribute("gidNumber", []string{"1212"})
	req.Attribute("email", []string{intEmail})
	req.Attribute("alternateEmail", []string{email})
	req.Attribute("mlAddress", []string{mlEmail})
	req.Attribute("serviceEnabled", []string{"htaccess", "mail", "redmine"})
	req.Attribute("token", []string{token})
	req.Attribute("userPassword", []string{"-"})
	req.Attribute("doorPassword", []string{"-"})

	err = l.conn.Add(req)
	if err != nil {
		return "", fmt.Errorf("unable to add user: %s", err)
	}

	return token, nil
}

func (l *LdapWrap) SetPassword(token, password, doorpass string) (err error) {
	passwordHash, err := ssha.Hash(password, ssha.SSHA)
	if err != nil {
		return fmt.Errorf("unable to hash password: %s", err)
	}
	doorpassHash, err := ssha.Hash(doorpass, ssha.SSHA512)
	if err != nil {
		return fmt.Errorf("unable to hash door password: %s", err)
	}

	token = ldap.EscapeFilter(token)
	search := fmt.Sprintf("(&(objectClass=backspaceMember)(token=%s))", token)
	sr, err := l.SearchActiveAndInactive(search, []string{})
	if err != nil {
		return fmt.Errorf("unable to search: %s", err)
	}
	if len(sr.Entries) != 1 {
		fmt.Println(search)
		return errors.New("no user with that token found")
	}
	member := sr.Entries[0]

	ok, err := ValidateToken(token, member.GetAttributeValue("uid"))
	if !ok || err != nil {
		return fmt.Errorf("invalid token: %s", err)
	}
	req := ldap.NewModifyRequest(member.DN, []ldap.Control{})
	req.Replace("userPassword", []string{passwordHash})
	req.Replace("doorPassword", []string{doorpassHash})
	req.Replace("token", []string{"**invalidated**"})

	err = l.conn.Modify(req)
	if err != nil {
		return fmt.Errorf("unable to set password: %s", err)
	}

	return nil
}

func (l *LdapWrap) MemberExists(uid string) (exists bool, err error) {
	filter := fmt.Sprintf("(&(objectClass=backspaceMember)(uid=%s))", EscapeFilter(uid))
	res, err := l.SearchActiveAndInactive(filter, []string{})
	if err != nil {
		return false, fmt.Errorf("unable to search: %s", err)
	}
	if len(res.Entries) > 0 {
		return true, nil
	}
	return
}

func (l *LdapWrap) SearchActive(filter string, attrs []string) (sr *ldap.SearchResult, err error) {
	r := ldap.NewSearchRequest(
		"ou=member,dc=backspace",
		ldap.ScopeWholeSubtree,
		ldap.NeverDerefAliases,
		0, 0, false,
		filter,
		attrs,
		[]ldap.Control{},
	)
	return l.conn.Search(r)
}

func (l *LdapWrap) SearchInactive(filter string, attrs []string) (sr *ldap.SearchResult, err error) {
	r := ldap.NewSearchRequest(
		"ou=inactiveMember,dc=backspace",
		ldap.ScopeWholeSubtree,
		ldap.NeverDerefAliases,
		0, 0, false,
		filter,
		attrs,
		[]ldap.Control{},
	)
	return l.conn.Search(r)
}

func (l *LdapWrap) SearchActiveAndInactive(filter string, attrs []string) (sr *ldap.SearchResult, err error) {
	sr1, err := l.SearchActive(filter, attrs)
	if err != nil {
		return
	}
	sr2, err := l.SearchInactive(filter, attrs)
	if err != nil {
		return
	}
	return &ldap.SearchResult{
		Entries:   append(sr1.Entries, sr2.Entries...),
		Referrals: append(sr1.Referrals, sr2.Referrals...),
		Controls:  append(sr1.Controls, sr2.Controls...),
	}, err
}
