package esqlo

import (
	"io"
	"net/http"
	"path"
	"strings"
)

// Handler is will wrap fileserver and render any esqlo templates returned. If a file is not html,
// it is served unmodified like a normal filesystem http server. If a file is html, it is parsed
// as a esqlo template and rendered with the sql statements automatically resolved using the configured database connections.
type Handler struct {
	Databases map[string]Database

	fileserver http.Handler // normal fileserver
}

func RenderAll(handler http.Handler) http.Handler {
	return &Handler{
		fileserver: handler,
	}
}

func (d *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fpath := path.Clean(r.URL.Path)
	if strings.HasSuffix(fpath, ".html") || fpath == "/" {
		pr, pw := io.Pipe()
		defer pr.Close()
		go func() {
			d.fileserver.ServeHTTP(&respWrapper{Writer: pw, w: w}, r)
			pw.Close()
		}()
		render := NewRenderer()
		render.Databases = d.Databases
		w.Header().Set("Content-Type", "text/html")
		render.RenderHTML(pr, w)
	} else {
		d.fileserver.ServeHTTP(w, r)
	}
}

// respWrapper takes a straem of bytes representing an html file and returns
type respWrapper struct {
	io.Writer
	w http.ResponseWriter
}

func (rb *respWrapper) Header() http.Header {
	return make(map[string][]string)
}

func (rb *respWrapper) WriteHeader(statusCode int) {
	rb.w.WriteHeader(statusCode)
}
