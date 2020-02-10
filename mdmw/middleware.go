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

type Middleware func(*Request) error
type responseBody struct {
	StatusCode int
	Title      string

	body   *bytes.Buffer
	header http.Header
}
type Request struct {
	body *responseBody

	ctx        context.Context
	req        *http.Request
	cancelFunc context.CancelFunc
}
type IsRaw struct{}

func (rb *responseBody) Header() http.Header {
	if rb.header == nil {
		rb.header = make(http.Header)
	}

	return rb.header
}

func (r *Request) Context() context.Context {
	if r.ctx == nil {
		r.ctx, r.cancelFunc = context.WithCancel(context.Background())
	}

	return r.ctx
}

func (r *Request) Cancel() {
	_ = r.Context() // make sure ctx & cancelFunc aren't nil

	r.cancelFunc()
}

func (r *Request) Request() *http.Request {
	return r.req
}

func (r *Request) Body() *responseBody {
	if r.body == nil {
		r.body = &responseBody{}
	}
	return r.body
}

func (rb *responseBody) Body() *bytes.Buffer {
	if rb.body == nil {
		rb.body = new(bytes.Buffer)
	}
	return rb.body
}

func middlewareChain(w http.ResponseWriter, req *http.Request, mws ...Middleware) {
	mdmwReq := Request{
		req: req,
	}
	var chain []string

	var err error
	for _, mw := range mws {
		chain = append(chain, getFunctionName(mw))

		err = mw(&mdmwReq)

		if mdmwReq.Context().Err() == context.Canceled || err != nil {
			break
		}
	}

	if err != nil {
		switch mdmwReq.Body().StatusCode {
		case 0:
			mdmwReq.Body().StatusCode = http.StatusInternalServerError
			fallthrough
		case http.StatusInternalServerError:
			fmt.Printf("error in request chain (uri=%s) %s: %v\n", req.RequestURI, strings.Join(chain, " -> "), err)

			mdmwReq.Body().Header().Set("Content-Type", "text/html")
			mdmwReq.Body().body = bytes.NewBufferString(HTMLServerError)
		case http.StatusNotFound:
			mdmwReq.Body().Header().Set("Content-Type", "text/html")
			mdmwReq.Body().body = bytes.NewBufferString(HTMLNotFound)
		case http.StatusForbidden:
			mdmwReq.Body().Header().Set("Content-Type", "text/html")
			mdmwReq.Body().body = bytes.NewBufferString(HTMLForbidden)
		}
	}

	if mdmwReq.Body().StatusCode == 0 {
		mdmwReq.Body().StatusCode = http.StatusOK
	}

	w.WriteHeader(mdmwReq.Body().StatusCode)
	for header, values := range mdmwReq.Body().Header() {
		for _, value := range values {
			w.Header().Add(header, value)
		}
	}
	mdmwReq.Body().Body().WriteTo(w)
}

func getFunctionName(i interface{}) string {
	name := runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name()
	parts := strings.Split(name, ".")
	if len(parts) == 0 {
		return ""
	}
	return parts[len(parts)-1]
}
