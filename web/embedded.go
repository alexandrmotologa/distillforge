package web

import (
	"embed"
	"io/fs"
	"net/http"
)

//go:embed all:dist
var distFS embed.FS

// StaticFS returns an http.FileSystem for the compiled React studio bundle.
func StaticFS() http.FileSystem {
	sub, err := fs.Sub(distFS, "dist")
	if err != nil {
		return nil
	}
	return http.FS(sub)
}
