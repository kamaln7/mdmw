package mdmw

import (
	"bytes"
	"fmt"
	"html/template"
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

	mux        *http.ServeMux
	outputTmpl *template.Template
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

		var (
			html      bytes.Buffer
			tmplInput = struct {
				Title, Body string
			}{
				Title: filepath.Base(path),
				Body:  string(output),
			}
		)

		err := s.outputTmpl.Execute(&html, tmplInput)
		if err != nil {
			fmt.Fprintf(os.Stderr, "couldn't execute output template: %v\n", err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(HTMLServerError))
			return
		}

		output = html.Bytes()
	}
	w.Write(output)
}
