package mdmw

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	textTemplate "text/template"

	"github.com/kamaln7/mdmw/mdmw/storage"
	blackfriday "gopkg.in/russross/blackfriday.v2"
)

var MarkdownExtensions = []string{".md", ".markdown", ".mdown", ".mkdn", ".mkd"}

// Server is a mdmw HTTP server
type Server struct {
	Storage                        storage.Driver
	ListenAddress                  string
	ValidateExtension, RootListing bool

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
	if s.RootListing && r.RequestURI == "/" {
		middlewareChain(w, r, s.fetchListing, s.renderMarkdown, s.prettyHTML)
		return
	}

	middlewareChain(w, r, s.trimRaw, s.validateExtension, s.getMarkdown, s.renderMarkdown, s.prettyHTML)
}

func (s *Server) trimRaw(res Response, req *http.Request) Response {
	path := req.RequestURI
	if strings.HasSuffix(path, "/raw") {
		req.RequestURI = strings.TrimSuffix(path, "/raw")
		res.ctx = context.WithValue(res.Context(), IsRaw{}, true)
	}

	return res
}

func (s *Server) validateExtension(res Response, req *http.Request) Response {
	if !s.ValidateExtension {
		return res
	}

	extension := filepath.Ext(req.RequestURI)
	for _, ext := range MarkdownExtensions {
		if ext == extension {
			return res
		}
	}

	return Response{
		StatusCode: http.StatusNotFound,
		Err:        fmt.Errorf("invalid extension %s", extension),
	}
}

func (s *Server) getMarkdown(res Response, req *http.Request) Response {
	output, err := s.Storage.Read(req.RequestURI)

	if err != nil {
		switch err {
		case storage.ErrNotFound:
			return Response{
				StatusCode: http.StatusNotFound,
				Err:        err,
			}
		case storage.ErrForbidden:
			return Response{
				StatusCode: http.StatusForbidden,
				Err:        fmt.Errorf("couldn't read from storage: %v", err),
			}
		default:
			return Response{
				Err: err,
			}
		}
	}

	res.Body = *bytes.NewBuffer(output)

	if res.Context().Value(IsRaw{}) != nil {
		// raw markdown
		res.Header().Set("Content-Type", "text/markdown")
		res.Err = errors.New("")
		res.StatusCode = http.StatusOK
		return res
	}

	return res
}

func (s *Server) prettyHTML(res Response, req *http.Request) Response {
	res.Header().Set("Content-Type", "text/html")

	var html bytes.Buffer

	title := filepath.Base(req.RequestURI)
	err := s.outputTmpl.Execute(&html, outputTemplateData{
		Title: title,
		Body:  template.HTML(res.Body.String()),
	})

	if err != nil {
		return Response{
			Err: fmt.Errorf("couldn't execute output template: %v", err),
		}
	}

	res.Title = title
	res.Body = *bytes.NewBuffer(html.Bytes())
	return res
}

func (s *Server) fetchListing(res Response, req *http.Request) Response {
	path := req.RequestURI

	// get files
	files, err := s.Storage.List(path)
	if err != nil {
		return Response{
			StatusCode: 500,
			Err:        fmt.Errorf("couldn't list files at %s: %v", path, err),
		}
	}

	title := fmt.Sprintf("Listing of %s", path)
	if len(files) == 0 {
		return Response{
			Title: title,
			Body:  *bytes.NewBufferString("There are no files here yet."),
		}
	}

	var md bytes.Buffer
	tmpl, err := textTemplate.New("").Parse(`
# {{.Title}}

{{range $file := .Files}}
* [{{$file.Name}}]({{$file.Path}})
{{end}}
`)
	if err != nil {
		return Response{
			Err: fmt.Errorf("couldn't parse output template: %v", err),
		}
	}

	err = tmpl.Execute(&md, listingTemplateData{
		Title: title,
		Files: files,
	})
	if err != nil {
		return Response{
			Err: fmt.Errorf("couldn't execute output template: %v", err),
		}
	}

	return Response{
		Title: title,
		Body:  md,
	}
}

func (s *Server) renderMarkdown(res Response, req *http.Request) Response {
	body := res.Body.Bytes()
	res.Body = *bytes.NewBuffer(blackfriday.Run(body))
	res.Header().Set("Content-Type", "text/markdown")

	return res
}
