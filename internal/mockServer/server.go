package mockServer

import (
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
)

type Server struct {
	config   *config
	listener net.Listener
}

func (s *Server) Url() string { return "https://" + s.listener.Addr().String() }

func (s *Server) Set(request []Request, t *testing.T) { *s.config = config{request: request, t: t} }

// Clear removes all configured requests and responses.
// Panics if there are unhandled requests remaining.
func (s *Server) Clear(t *testing.T) {
	if s.config.requestNumber < len(s.config.request) {
		assert.FailNow(t, "There are unhandled requests remaining")
	}
}
