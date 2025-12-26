package mockServer

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

type config struct {
	requestNumber int
	request       []Request
	t             *testing.T
}

func (c *config) handle(w http.ResponseWriter, r *http.Request) {
	if c.requestNumber >= len(c.request) {
		assert.FailNow(c.t, "Received more requests than expected")
		return
	}
	c.request[c.requestNumber].handle(w, r, c.t)
	c.requestNumber++
}
