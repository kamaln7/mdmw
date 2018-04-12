package mdmw

import (
	"fmt"
	"net/http"
	"os"

	"github.com/kamaln7/mdmw/mdmw/storage"
	blackfriday "gopkg.in/russross/blackfriday.v2"
)

// Server is a mdmw HTTP server
type Server struct {
	StorageDriver storage.Driver
	ListenAddress string

	mux *http.ServeMux
}

// Listen starts the actual HTTP server
func (s *Server) Listen() {
	s.mux = http.NewServeMux()
	s.mux.HandleFunc("/", s.httpHandler)

	fmt.Printf("mdmw listening on %s\n", s.ListenAddress)
	if err := http.ListenAndServe(s.ListenAddress, s.mux); err != nil {
		fmt.Fprintf(os.Stderr, "couldn't start HTTP server: %v\n", err)
	}
}

func (s *Server) httpHandler(w http.ResponseWriter, r *http.Request) {
	path := r.RequestURI
	source, err := s.StorageDriver.Read(path)

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

	output := blackfriday.Run(source)
	w.Write(output)
}
