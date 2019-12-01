package statics

import (
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/rakyll/statik/fs"
)

func MustStatics() http.FileSystem {
	statics, err := Statics()
	if err != nil {
		log.Fatalf("unable to load static files: %s", err)
	}
	return statics
}

func Statics() (statics http.FileSystem, err error) {
	if strings.Contains(os.Args[0], "go-build") {
		log.Printf("go run, using files")
		statics = http.Dir("web")
	} else {
		statics, err = fs.New()
	}
	return
}
