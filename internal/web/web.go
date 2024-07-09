package web

import (
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"path/filepath"

	"github.com/b4ckspace/members/internal/core"
	"github.com/b4ckspace/members/internal/statics"
	_ "github.com/b4ckspace/members/statik"
)

const (
	SUCCESS MessageKind = "success"
	WARNING MessageKind = "warning"
	DANGER  MessageKind = "danger"
)

type (
	Web struct {
		mailer     core.Mailer
		mux        http.Handler
		ldapDialer core.LdapDialer
		templates  map[string]*template.Template
		statics    http.FileSystem
	}
	MessageKind string
	Message     struct {
		Kind    MessageKind
		Message string
	}

	RegisterTemplateData struct {
		Form     *RegisterForm
		Messages []Message
	}
	ResetTemplateData struct {
		Form     *ResetForm
		Messages []Message
	}
	PasswordTemplateData struct {
		Form     *PasswordForm
		Messages []Message
	}
)

func New(mailer core.Mailer, ld core.LdapDialer) (web *Web, err error) {
	mux := http.NewServeMux()
	web = &Web{
		mailer:     mailer,
		mux:        web.registerMiddlewares(mux, logMiddleware),
		ldapDialer: ld,
		templates:  map[string]*template.Template{},
		statics:    statics.MustStatics(),
	}
	templates := []string{"index.html", "register.html", "reset.html", "password.html"}
	for _, tplFile := range templates {
		tt, err := web.templateParseFilesFromFs(
			"/templates/base.html",
			fmt.Sprintf("/templates/%s", tplFile),
		)
		if err != nil {
			return nil, fmt.Errorf(
				"unable to load template %s: %s",
				tplFile, err,
			)
		}
		web.templates[tplFile] = tt
	}
	web.registerRoutes(mux)
	return web, nil
}

func (web *Web) GetMux() http.Handler {
	return web.mux
}

func (web *Web) registerMiddlewares(mux http.Handler, middlewares ...func(http.Handler) http.Handler) (handler http.Handler) {
	handler = mux
	for _, middleware := range middlewares {
		handler = middleware(handler)
	}
	return
}

func (web *Web) registerRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		err := web.templates["index.html"].Execute(w, nil)
		if err != nil {
			log.Printf("unable to render template: %s", err)
		}
	})
	mux.HandleFunc("/reset", func(w http.ResponseWriter, r *http.Request) {
		td := web.handleReset(r)
		err := web.templates["reset.html"].Execute(w, td)
		if err != nil {
			log.Printf("unable to render template: %s", err)
		}
	})
	mux.HandleFunc("/register", func(w http.ResponseWriter, r *http.Request) {
		td := web.handleRegister(r)
		err := web.templates["register.html"].Execute(w, td)
		if err != nil {
			log.Printf("unable to render template: %s", err)
		}
	})
	mux.HandleFunc("/password", func(w http.ResponseWriter, r *http.Request) {
		td := web.handlePassword(r)
		err := web.templates["password.html"].Execute(w, td)
		if err != nil {
			log.Printf("unable to render template: %s", err)
		}
	})

	// static files
	mux.Handle("/static/", http.FileServer(web.statics))
}

func (web *Web) templateParseFilesFromFs(files ...string) (t *template.Template, err error) {
	for _, file := range files {
		fp, err := web.statics.Open(file)
		if err != nil {
			return nil, err
		}
		c, err := io.ReadAll(fp)
		if err != nil {
			return nil, err
		}
		fileName := filepath.Base(file)
		if t == nil {
			t = template.New(fileName)
		}
		var tmpl *template.Template
		if t.Name() == fileName {
			tmpl = t
		} else {
			tmpl = t.New(fileName)
		}
		_, err = tmpl.Parse(string(c))
		if err != nil {
			return nil, err
		}
	}
	return
}
