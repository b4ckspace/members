package statics

import (
	"embed"
	"net/http"
)

//go:embed static templates
var EmbeddedStaticFiles embed.FS

func Statics() (statics http.FileSystem) {
	statics = http.FS(EmbeddedStaticFiles)
	return
}
