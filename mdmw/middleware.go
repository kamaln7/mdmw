package mdmw

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"reflect"
	"runtime"
	"strings"
)

type Middleware func(Response, *http.Request) Response
type Response struct {
	Err        error
	StatusCode int
	Title      string
	Body       bytes.Buffer

	ctx    context.Context
	header http.Header
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
		r.ctx = context.Background()
	}

	return r.ctx
}

func middlewareChain(w http.ResponseWriter, req *http.Request, mws ...Middleware) {
	res := Response{}
	var chain []string

	for _, mw := range mws {
		chain = append(chain, getFunctionName(mw))

		res = mw(res, req)

		if res.Err == nil {
			continue
		}

		if res.StatusCode == 0 {
			res.StatusCode = http.StatusInternalServerError
		}

		switch res.StatusCode {
		case http.StatusInternalServerError:
			fmt.Printf("error in request chain %s: %v\n", strings.Join(chain, " -> "), res.Err)
			fmt.Printf("request uri: %s\n", req.RequestURI)

			res.Header().Set("Content-Type", "text/html")
			res.Body = *bytes.NewBufferString(HTMLServerError)
		case http.StatusNotFound:
			res.Header().Set("Content-Type", "text/html")
			res.Body = *bytes.NewBufferString(HTMLNotFound)
		case http.StatusForbidden:
			res.Header().Set("Content-Type", "text/html")
			res.Body = *bytes.NewBufferString(HTMLForbidden)
		}

		break
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
	res.Body.WriteTo(w)
}

func getFunctionName(i interface{}) string {
	return runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name()
}
