package mockServer

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

type Request struct {
	Method      Method
	Path        Path
	HandlerFunc func(w http.ResponseWriter, r *http.Request, t *testing.T)
}

func (rule Request) handle(w http.ResponseWriter, r *http.Request, t *testing.T) {
	if r.Method != rule.Method.String() {
		assert.FailNow(t, "Unexpected method. Expected: "+rule.Method.String()+" Got: "+r.Method)
	}
	if r.RequestURI != rule.Path.String() {
		assert.FailNow(t, "Unexpected path. Expected: "+rule.Path.String()+" Got: "+r.RequestURI)
	}
	rule.HandlerFunc(w, r, t)
}

func Append(r ...[]Request) []Request {
	new := make([]Request, 0, len(r)*2)
	for i := range r {
		new = append(new, r[i]...)
	}
	return new
}
