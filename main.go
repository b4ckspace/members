//go:generate statik -f -src web
package main

import (
	"flag"
	"log"
	"net/http"
	"os"

	"github.com/b4ckspace/members/internal/ldapwrap"
	"github.com/b4ckspace/members/internal/mailer"
	"github.com/b4ckspace/members/internal/web"
)

type (
	Args struct {
		LdapServer string
		LdapPort   int
		LdapUser   string
		LdapPass   string
		MailServer string
		WebListen  string
	}
)

func main() {
	args := Args{}
	flag.StringVar(&args.LdapServer, "server", "ldap.example.com", "ldap server")
	flag.StringVar(&args.LdapUser, "user", "uid=user,dc=example", "ldap user")
	flag.IntVar(&args.LdapPort, "port", 389, "ldap port")
	flag.StringVar(&args.MailServer, "mailserver", "localhost:25", "email server")
	flag.StringVar(&args.WebListen, "listen", ":8080", "address to listen on")
	flag.Parse()

	// ldap
	var ok bool
	args.LdapPass, ok = os.LookupEnv("LDAP_PASSWORD")
	if !ok {
		log.Fatalf("unable to load LDAP_PASSWORD from environment")
	}
	ldapConnFactory := ldapwrap.NewLdapConnFactory(
		args.LdapServer,
		args.LdapPort,
		args.LdapUser,
		args.LdapPass,
	)
	l, err := ldapwrap.New(ldapConnFactory)
	if err != nil {
		log.Fatalf("unable to connect to ldap: %s", err)
	}

	// mailer
	mlr := mailer.New(mailer.SmtpConnFactory(args.MailServer))

	// webinterface
	w, err := web.New(mlr, l)
	if err != nil {
		log.Fatalf("unable to start webserver: %s", err)
	}
	err = http.ListenAndServe(args.WebListen, w.GetMux())
	if err != nil {
		log.Fatalf("webserver crashed: %s", err)
	}

}
