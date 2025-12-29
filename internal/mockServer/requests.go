package mockServer

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"
)

func RequestsAuth() []Request {
	return []Request{{
		Path:   "/access/ticket",
		Method: POST,
		HandlerFunc: func(w http.ResponseWriter, r *http.Request, t *testing.T) {
			w.Write([]byte(`{"data":{"ticket":"FAKE_TICKET","CSRFPreventionToken":"FAKE_CSRF_TOKEN"}}`))
		}}}
}

func RequestsDelete(urlPath Path, expected any) []Request {
	return []Request{{
		Path:   urlPath,
		Method: DELETE,
		HandlerFunc: func(w http.ResponseWriter, r *http.Request, t *testing.T) {
			requestsParseParams(expected, r, t)
		}}}
}

func RequestsGetJson(urlPath Path, v any) []Request {
	return []Request{{
		Path:   urlPath,
		Method: GET,
		HandlerFunc: func(w http.ResponseWriter, r *http.Request, t *testing.T) {
			tmp, _ := json.Marshal(v)
			w.Write(tmp)
		}}}
}

// RequestsPost creates a request that expects a POST with JSON body matching 'expected'
// all values in 'expected' will be treated as strings or arrays of strings.
func RequestsPost(urlPath Path, expected any) []Request {
	return []Request{{
		Path:   urlPath,
		Method: POST,
		HandlerFunc: func(w http.ResponseWriter, r *http.Request, t *testing.T) {
			requestsParseParams(expected, r, t)
		}}}
}

func RequestsPostResponse(urlPath Path, expected any, response []byte) []Request {
	return []Request{{
		Path:   urlPath,
		Method: POST,
		HandlerFunc: func(w http.ResponseWriter, r *http.Request, t *testing.T) {
			requestsParseParams(expected, r, t)
			w.Write(response)
		}}}
}

// RequestsPut creates a request that expects a PUT with JSON body matching 'expected'
// all values in 'expected' will be treated as strings or arrays of strings.
func RequestsPut(urlPath Path, expected any) []Request {
	return []Request{{
		Path:   urlPath,
		Method: PUT,
		HandlerFunc: func(w http.ResponseWriter, r *http.Request, t *testing.T) {
			requestsParseParams(expected, r, t)
		}}}
}

func RequestsPutHandler(urlPath Path, handler func(t *testing.T, v url.Values)) []Request {
	return []Request{{
		Path:   urlPath,
		Method: PUT,
		HandlerFunc: func(w http.ResponseWriter, r *http.Request, t *testing.T) {
			values := requestsParseParamsPartial(t, r)
			handler(t, values)
		}}}
}

func requestsParseParamsPartial(t *testing.T, r *http.Request) url.Values {
	body, err := io.ReadAll(r.Body)
	require.NoError(t, err)

	values, err := url.ParseQuery(string(body))
	require.NoError(t, err)
	return values
}

func requestsParseParams(expected any, r *http.Request, t *testing.T) {
	values := requestsParseParamsPartial(t, r)
	out := make(map[string]any)
	for k, v := range values {
		if len(v) > 0 {
			out[k] = v[0]
		}
	}
	if expected == nil {
		expected = map[string]any{}
	}
	require.Equal(t, expected, out)
}

func RequestsErrorHandled(url Path, method Method, err HTTPerror) []Request {
	return []Request{{
		Path:   url,
		Method: method,
		HandlerFunc: func(w http.ResponseWriter, r *http.Request, t *testing.T) {
			http.Error(w, err.Message, int(err.Code))
		}}}
}

func RequestsError(url Path, method Method, Code HTTPcode, amount uint) []Request {
	requests := make([]Request, amount)
	for i := range int(amount) {
		requests[i] = Request{
			Path:   url,
			Method: method,
			HandlerFunc: func(w http.ResponseWriter, r *http.Request, t *testing.T) {
				http.Error(w, "", int(Code))
			}}
	}
	return requests
}
