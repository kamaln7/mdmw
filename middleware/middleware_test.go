package middleware

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCtx_Apply(t *testing.T) {
	t.Run("apply Ctx to ResponseWriter", func(t *testing.T) {
		body := []byte("test")
		headerKey := "test-header"
		headerValue := "test value"
		statusCode := http.StatusTeapot

		req := httptest.NewRequest("GET", "http://example.com/", nil)
		w := httptest.NewRecorder()

		// create + apply
		ctx := New(req)
		ctx.Body = body
		ctx.Header().Add(headerKey, headerValue)
		ctx.StatusCode = statusCode
		ctx.Apply(w)

		// check
		resp := w.Result()
		gotBody, _ := ioutil.ReadAll(resp.Body)

		assert.Equal(t, body, gotBody)
		assert.Equal(t, resp.StatusCode, statusCode)
		assert.Equal(t, headerValue, resp.Header.Get(headerKey))
	})
}

func Test_getFunctionName(t *testing.T) {
	t.Run("get function name", func(t *testing.T) {
		want := "Test_getFunctionName"
		got := getFunctionName(Test_getFunctionName)

		assert.Equal(t, want, got)
	})
}
