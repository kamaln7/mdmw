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
	"github.com/kamaln7/mdmw/mdmw/middleware"
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
	Verbose           bool

	mux        *http.ServeMux
	outputTmpl *template.Template
}

type key int

const (
	_ key = iota
	statusCode
	pageTitle
	isRaw
)

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
		s.Serve(w, r, s.fetchListing, s.renderMarkdown, s.prettyHTML)
		return
	}

	s.Serve(w, r, s.trimRaw, s.validateExtension, s.getMarkdown, s.renderMarkdown, s.prettyHTML)
}

// Serve serves a chain of middleware + pretty errors
func (s *Server) Serve(w http.ResponseWriter, r *http.Request, mws ...middleware.Middleware) {
	// first run the chain normally
	ctx := middleware.Chain(middleware.New(r), mws...)

	// then run again with prettyErrors
	// this makes sure prettyErrors is run even if the middleware chain
	postProcess := []middleware.Middleware{s.prettyErrors}

	if s.Verbose {
		postProcess = append(postProcess, middleware.Log)
	}

	ctx = middleware.Chain(ctx, postProcess...)

	// serve
	ctx.Apply(w)
}

func (s *Server) prettyErrors(ctx *middleware.Ctx) error {
	req := ctx.Request()
	err := req.Context().Err()

	// set the default status code for a non-nil and non-context-canceled error to 500
	if err != nil &&
		err != context.Canceled &&
		ctx.StatusCode == 0 {

		ctx.StatusCode = http.StatusInternalServerError
	}

	// some error occurred
	switch ctx.StatusCode {
	case http.StatusInternalServerError:
		fmt.Printf("error in request chain (uri=%s) %s: %v\n", req.RequestURI, strings.Join(ctx.Chain(), " -> "), err)

		ctx.Header().Set("Content-Type", "text/html")
		ctx.Body = []byte(HTMLServerError)
	case http.StatusNotFound:
		ctx.Header().Set("Content-Type", "text/html")
		ctx.Body = []byte(HTMLNotFound)
	case http.StatusForbidden:
		ctx.Header().Set("Content-Type", "text/html")
		ctx.Body = []byte(HTMLForbidden)
	}

	return nil
}

func (s *Server) trimRaw(ctx *middleware.Ctx) error {
	req := ctx.Request()

	path := req.RequestURI
	if strings.HasSuffix(path, "/raw") {
		req.RequestURI = strings.TrimSuffix(path, "/raw")
		ctx.WithValue(isRaw, true)
	}

	return nil
}

func (s *Server) validateExtension(ctx *middleware.Ctx) error {
	req := ctx.Request()
	if !s.ValidateExtension {
		return nil
	}

	extension := filepath.Ext(req.RequestURI)
	for _, ext := range MarkdownExtensions {
		if ext == extension {
			return nil
		}
	}

	ctx.StatusCode = http.StatusNotFound
	return fmt.Errorf("invalid extension %s", extension)
}

func (s *Server) getMarkdown(ctx *middleware.Ctx) error {
	req := ctx.Request()
	output, err := s.Storage.Read(req.RequestURI)

	if err != nil {
		switch err {
		case storage.ErrNotFound:
			ctx.StatusCode = http.StatusNotFound
			return fmt.Errorf("object not found: %v", err)
		case storage.ErrForbidden:
			ctx.StatusCode = http.StatusForbidden
			return fmt.Errorf("couldn't read from storage: %v", err)
		default:
			return err
		}
	}

	ctx.Body = output

	if ctx.Context().Value(isRaw) != nil {
		// raw markdown
		ctx.Header().Set("Content-Type", "text/markdown")
		ctx.Cancel()
	}

	return nil
}

func (s *Server) prettyHTML(ctx *middleware.Ctx) error {
	req := ctx.Request()
	ctx.Header().Set("Content-Type", "text/html")

	var html bytes.Buffer

	title := filepath.Base(req.RequestURI)
	err := s.outputTmpl.Execute(&html, outputTemplateData{
		Title: title,
		Body:  template.HTML(string(ctx.Body)),
	})

	if err != nil {
		return fmt.Errorf("couldn't execute output template: %v", err)
	}

	ctx.WithValue(pageTitle, title)
	ctx.Body = html.Bytes()
	return nil
}

func (s *Server) fetchListing(ctx *middleware.Ctx) error {
	req := ctx.Request()
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

	ctx.WithValue(pageTitle, title)
	ctx.Body = md.Bytes()
	return nil
}

func (s *Server) renderMarkdown(ctx *middleware.Ctx) error {
	ctx.Body = blackfriday.Run(ctx.Body)
	ctx.Header().Set("Content-Type", "text/markdown")

	return nil
}
