//go:build !dev

package app

import (
	"embed"
	"io/fs"
	"net/http"
	"path"
	"strings"
)

//go:embed all:dist
var dist embed.FS

// registerFrontend mounts the embedded static files, falling back to index.html for
// unknown paths (SPA routing).
func registerFrontend(mux *http.ServeMux) {
	root, err := fs.Sub(dist, "dist")
	if err != nil {
		panic(err)
	}
	files := http.FileServerFS(root)

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		name := strings.TrimPrefix(path.Clean(r.URL.Path), "/")
		if name == "" {
			name = "index.html"
		}
		if info, statErr := fs.Stat(root, name); statErr != nil || info.IsDir() && name != "index.html" {
			if _, indexErr := fs.Stat(root, path.Join(name, "index.html")); indexErr != nil {
				r.URL.Path = "/"
			}
		}
		files.ServeHTTP(w, r)
	})
}
