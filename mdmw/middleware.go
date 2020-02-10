package mdmw

import (
	"context"
	"fmt"
	"net/http"
	"reflect"
	"runtime"
	"strings"
)

type Middleware func(*http.Request, *Response) error
type Response struct {
	StatusCode int
	Title      string
	Body       []byte

	header http.Header
	ctx    context.Context
	cancel context.CancelFunc
}
type IsRaw struct{}

func (r *Response) Header() http.Header {
	if r.header == nil {
		r.header = make(http.Header)
	}

	return r.header
}

func (r *Response) Context() context.Context {
	if r.ctx == nil {
		r.ctx, r.cancel = context.WithCancel(context.Background())
	}

	return r.ctx
}

func (r *Response) Cancel() {
	_ = r.Context() // make sure ctx & cancelFunc aren't nil

	r.cancel()
}

func middlewareChain(w http.ResponseWriter, req *http.Request, mws ...Middleware) {
	var (
		res   = &Response{}
		chain []string
		err   error
	)

	for _, mw := range mws {
		chain = append(chain, getFunctionName(mw))

		err = mw(req, res)

		if res.Context().Err() == context.Canceled || err != nil {
			break
		}
	}

	if err != nil {
		switch res.StatusCode {
		case 0:
			res.StatusCode = http.StatusInternalServerError
			fallthrough
		case http.StatusInternalServerError:
			fmt.Printf("error in request chain (uri=%s) %s: %v\n", req.RequestURI, strings.Join(chain, " -> "), err)

			res.Header().Set("Content-Type", "text/html")
			res.Body = []byte(HTMLServerError)
		case http.StatusNotFound:
			res.Header().Set("Content-Type", "text/html")
			res.Body = []byte(HTMLNotFound)
		case http.StatusForbidden:
			res.Header().Set("Content-Type", "text/html")
			res.Body = []byte(HTMLForbidden)
		}
	}

	if res.StatusCode == 0 {
		res.StatusCode = http.StatusOK
	}

	w.WriteHeader(res.StatusCode)
	for header, values := range res.Header() {
		for _, value := range values {
			w.Header().Add(header, value)
		}
	}
	w.Write(res.Body)
}

func getFunctionName(i interface{}) string {
	name := runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name()
	parts := strings.Split(name, ".")
	if len(parts) == 0 {
		return ""
	}
	return parts[len(parts)-1]
}
