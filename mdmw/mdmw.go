package mdmw

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/kamaln7/mdmw/mdmw/storage"
	blackfriday "gopkg.in/russross/blackfriday.v2"
)

var MarkdownExtensions = []string{".md", ".markdown", ".mdown", ".mkdn", ".mkd"}

// Server is a mdmw HTTP server
type Server struct {
	StorageDriver     storage.Driver
	ListenAddress     string
	ValidateExtension bool

	mux *http.ServeMux
}

// Listen starts the actual HTTP server
func (s *Server) Listen() {
	s.mux = http.NewServeMux()
	s.mux.HandleFunc("/", s.httpHandler)

	fmt.Printf("mdmw listening on %s\n", s.ListenAddress)
	if err := http.ListenAndServe(s.ListenAddress, s.mux); err != nil {
		fmt.Fprintf(os.Stderr, "couldn't start HTTP server: %v\n", err)
		os.Exit(1)
	}
}

func (s *Server) httpHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	var (
		path = r.RequestURI
		raw  = false
	)

	if strings.HasSuffix(path, "/raw") {
		raw = true
		path = strings.TrimSuffix(path, "/raw")
	}

	if s.ValidateExtension {
		extension := filepath.Ext(path)
		valid := false
		for _, ext := range MarkdownExtensions {
			if ext == extension {
				valid = true
				break
			}
		}

		if !valid {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(HTMLNotFound))
			return
		}
	}

	output, err := s.StorageDriver.Read(path)

	if err != nil {
		if err == storage.ErrNotFound {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(HTMLNotFound))
			return
		}

		fmt.Fprintf(os.Stderr, "couldn't serve %s: %v\n", path, err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(HTMLServerError))
		return
	}

	if raw {
		// raw markdown
		w.Header().Set("Content-Type", "text/markdown")
	} else {
		// render markdown as html
		w.Header().Set("Content-Type", "text/html")
		output = blackfriday.Run(output)

		// poor man's templating
		html := strings.Replace(HTMLOutput, "$body", string(output), 1)
		html = strings.Replace(html, "$title", filepath.Base(path), -1)
		output = []byte(html)
	}
	w.Write(output)
}
