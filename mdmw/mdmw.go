package mdmw

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	textTemplate "text/template"

	"github.com/igorsobreira/titlecase"
	"github.com/kamaln7/mdmw/mdmw/storage"
	blackfriday "gopkg.in/russross/blackfriday.v2"
)

var MarkdownExtensions = []string{".md", ".markdown", ".mdown", ".mkdn", ".mkd"}

const (
	_ = iota
	ListingOff
	ListingFiles
	ListingTitleCase
)

// Server is a mdmw HTTP server
type Server struct {
	Storage           storage.Driver
	ListenAddress     string
	ValidateExtension bool
	RootListing       int
	RootListingTitle  string

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
	if s.RootListing != ListingOff && r.RequestURI == "/" {
		middlewareChain(w, r, s.fetchListing, s.renderMarkdown, s.prettyHTML)
		return
	}

	middlewareChain(w, r, s.trimRaw, s.validateExtension, s.getMarkdown, s.renderMarkdown, s.prettyHTML)
}

func (s *Server) trimRaw(req *http.Request, res *Response) error {
	path := req.RequestURI
	if strings.HasSuffix(path, "/raw") {
		req.RequestURI = strings.TrimSuffix(path, "/raw")
		res.ctx = context.WithValue(res.Context(), IsRaw{}, true)
	}

	return nil
}

func (s *Server) validateExtension(req *http.Request, res *Response) error {
	if !s.ValidateExtension {
		return nil
	}

	extension := filepath.Ext(req.RequestURI)
	for _, ext := range MarkdownExtensions {
		if ext == extension {
			return nil
		}
	}

	res.StatusCode = http.StatusNotFound
	return fmt.Errorf("invalid extension %s", extension)
}

func (s *Server) getMarkdown(req *http.Request, res *Response) error {
	output, err := s.Storage.Read(req.RequestURI)

	if err != nil {
		switch err {
		case storage.ErrNotFound:
			res.StatusCode = http.StatusNotFound
			return fmt.Errorf("object not found: %v", err)
		case storage.ErrForbidden:
			res.StatusCode = http.StatusForbidden
			return fmt.Errorf("couldn't read from storage: %v", err)
		default:
			return err
		}
	}

	res.Body = output

	if res.Context().Value(IsRaw{}) != nil {
		// raw markdown
		res.Header().Set("Content-Type", "text/markdown")
		res.Cancel()
	}

	return nil
}

func (s *Server) prettyHTML(req *http.Request, res *Response) error {
	res.Header().Set("Content-Type", "text/html")

	var html bytes.Buffer

	title := filepath.Base(req.RequestURI)
	err := s.outputTmpl.Execute(&html, outputTemplateData{
		Title: title,
		Body:  template.HTML(string(res.Body)),
	})

	if err != nil {
		return fmt.Errorf("couldn't execute output template: %v", err)
	}

	res.Title = title
	res.Body = html.Bytes()
	return nil
}

func (s *Server) fetchListing(req *http.Request, res *Response) error {
	path := req.RequestURI

	// get files
	files, err := s.Storage.List(path)
	if err != nil {
		return fmt.Errorf("couldn't list files at %s: %v", path, err)
	}

	// transform filename according to the listing type
	if s.RootListing == ListingTitleCase {
		for i, file := range files {
			name := file.Name
			// remove extension
			name = strings.TrimSuffix(name, filepath.Ext(name))
			// replace dashes with spaces
			name = strings.NewReplacer(
				"_", " ",
				"-", " ",
			).Replace(name)
			// titlecase
			name = titlecase.Title(name)

			files[i].Name = name
		}
	}

	title := s.RootListingTitle
	if title == "" {
		title = fmt.Sprintf("Listing of %s", path)
	}

	md := new(bytes.Buffer)
	tmpl, err := textTemplate.New("").Parse(`
# {{.Title}}

{{if .Files}}
	{{range $file := .Files}}
* [{{$file.Name}}]({{$file.Path}})
	{{end}}
{{else}}
There are no files here
{{end}}
`)
	if err != nil {
		return fmt.Errorf("couldn't parse output template: %v", err)
	}

	err = tmpl.Execute(md, listingTemplateData{
		Title: title,
		Files: files,
	})
	if err != nil {
		return fmt.Errorf("couldn't execute output template: %v", err)
	}

	res.Title = title
	res.Body = md.Bytes()
	return nil
}

func (s *Server) renderMarkdown(req *http.Request, res *Response) error {
	res.Body = blackfriday.Run(res.Body)
	res.Header().Set("Content-Type", "text/markdown")

	return nil
}
