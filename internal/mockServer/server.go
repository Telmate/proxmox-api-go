package mockServer

import (
	"net"
	"testing"
)

type Server struct {
	config   *config
	listener net.Listener
}

func (s *Server) Url() string { return "https://" + s.listener.Addr().String() }

func (s *Server) Set(request []Request, t *testing.T) { *s.config = config{request: request, t: t} }
