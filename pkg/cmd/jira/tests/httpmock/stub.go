package httpmock

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"
)

type Matcher func(req *http.Request) bool
type Responder func(req *http.Request) (*http.Response, error)

type Stub struct {
	matched   bool
	Matcher   Matcher
	Responder Responder
}

func MatchAny(*http.Request) bool {
	return true
}

func REST(method, p string) Matcher {
	return func(req *http.Request) bool {
		if !strings.EqualFold(req.Method, method) {
			return false
		}
		return req.URL.EscapedPath() == "/"+p
	}
}

func QueryMatcher(method string, path string, query url.Values) Matcher {
	return func(req *http.Request) bool {
		if !REST(method, path)(req) {
			return false
		}

		actualQuery := req.URL.Query()

		for param := range query {
			if !(actualQuery.Get(param) == query.Get(param)) {
				return false
			}
		}

		return true
	}
}

func StringResponse(body string) Responder {
	return func(req *http.Request) (*http.Response, error) {
		return httpResponse(200, req, bytes.NewBufferString(body)), nil
	}
}

func WithHeader(responder Responder, header string, value string) Responder {
	return func(req *http.Request) (*http.Response, error) {
		resp, _ := responder(req)
		if resp.Header == nil {
			resp.Header = make(http.Header)
		}
		resp.Header.Set(header, value)
		return resp, nil
	}
}

func StatusStringResponse(status int, body string) Responder {
	return func(req *http.Request) (*http.Response, error) {
		return httpResponse(status, req, bytes.NewBufferString(body)), nil
	}
}

func JSONResponse(body interface{}) Responder {
	return func(req *http.Request) (*http.Response, error) {
		b, _ := json.Marshal(body)
		return httpResponse(200, req, bytes.NewBuffer(b)), nil
	}
}

func httpResponse(status int, req *http.Request, body io.Reader) *http.Response {
	return &http.Response{
		StatusCode: status,
		Request:    req,
		Body:       io.NopCloser(body),
		Header:     http.Header{},
	}
}
