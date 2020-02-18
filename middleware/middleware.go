package middleware

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"runtime"
	"strings"
)

// Middleware is an HTTP middleware
type Middleware func(*Ctx) error

// Ctx represents an incoming Request as a middleware context
type Ctx struct {
	Body       []byte
	StatusCode int

	header http.Header
	req    *http.Request
	cancel context.CancelFunc
	chain  []string
}

// New creates a new middleware context using an HTTP request
func New(req *http.Request) *Ctx {
	reqCtx, cancelFunc := context.WithCancel(req.Context())

	return &Ctx{
		req:    req.WithContext(reqCtx),
		cancel: cancelFunc,
	}
}

// Request returns the incoming request
func (c *Ctx) Request() *http.Request {
	return c.req
}

// Header returns the response's HTTP Headers map
func (c *Ctx) Header() http.Header {
	if c.header == nil {
		c.header = make(http.Header)
	}

	return c.header
}

// Cancel cancels the embedded context
func (c *Ctx) Cancel() {
	c.cancel()
}

// Chain returns the chain of middleware that was run so far
func (c *Ctx) Chain() []string {
	return c.chain
}

// Context returns the embedded Request's context
func (c *Ctx) Context() context.Context {
	return c.Request().Context()
}

// WithValue adds a value to the embedded context
func (c *Ctx) WithValue(key, value interface{}) {
	newCtx := context.WithValue(c.Context(), key, value)
	c.req = c.Request().WithContext(newCtx)
}

// Chain runs a slice of Middlewares and returns an error of interrupted
// along with the chain of middleware that was run up until that point
func Chain(ctx *Ctx, mws ...Middleware) *Ctx {
	var (
		chain = ctx.Chain()
		err   error
	)

	for _, mw := range mws {
		chain = append(chain, getFunctionName(mw))
		ctx.chain = chain

		err = mw(ctx)

		// stop if the context was canceled or if an error was returned
		if ctx.Context().Err() == context.Canceled || err != nil {
			break
		}
	}

	return ctx
}

// Serve runs a chain of middleware and writes the final response to an http.ResponseWriter
func Serve(w http.ResponseWriter, req *http.Request, mws ...Middleware) {
	Chain(New(req), mws...).Apply(w)
}

// Apply applies the context to an http.ResponseWriter
func (c *Ctx) Apply(w http.ResponseWriter) {
	for header, values := range c.Header() {
		for _, value := range values {
			w.Header().Add(header, value)
		}
	}

	if c.StatusCode != 0 {
		w.WriteHeader(c.StatusCode)
	}

	w.Write(c.Body)
}

// Log logs the request chain for debugging purposes
func Log(w io.Writer) Middleware {
	return func(ctx *Ctx) error {
		fmt.Fprintf(w, "uri [%s], chain [%s], status [%d]\n",
			ctx.Request().RequestURI,
			ctx.Chain(),
			ctx.StatusCode,
		)

		return nil
	}
}

func getFunctionName(i interface{}) string {
	name := runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name()
	parts := strings.Split(name, ".")
	if len(parts) == 0 {
		return ""
	}
	return parts[len(parts)-1]
}
