package server

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"

	"andboson/mock-server/internal/services/expectations"
	"andboson/mock-server/internal/templates"
)

const (
	ServerAddrHTTP    = "SERVER_ADDR_HTTP"
	DefaultServerAddr = ":8081"
)

type Server struct {
	address string
	server  *http.Server
	store   *expectations.Store

	tpls *templates.Templates
}

// NewServer returns instance of a service and sets up a Server
func NewServer(addr string, tpls *templates.Templates, store *expectations.Store) *Server {
	mux := http.NewServeMux()

	if addr == "" {
		addr = DefaultServerAddr
	}

	s := &Server{
		tpls:    tpls,
		address: addr,
		store:   store,
		server: &http.Server{
			Handler: mux,
		},
	}

	mux.Handle("/", s.createHTTPHandler())

	return s
}

// Start starts a httpserver
func (s *Server) Start() error {
	ln, err := net.Listen("tcp", s.address)
	if err != nil {
		return fmt.Errorf("could not listen on %s: %w", s.address, err)
	}

	log.Printf("HTTP CLI LOGGER Server started: %s", s.address)

	for _, exp := range s.store.DumpAvailableExpectations() {
		log.Println(exp.String())
	}
	if err := s.server.Serve(ln); err != nil {
		return fmt.Errorf("can't start Server: %w", err)
	}

	return nil
}

func (s *Server) Stop(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}

func (s *Server) createHTTPHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// serve index page
		if r.RequestURI == "/" && r.Method == http.MethodGet {
			w.Header().Add("Content-Type", "text/html; charset=utf-8")
			if err := s.tpls.Tpls.Execute(w, s.store.GetHistory(true)); err != nil {
				_, _ = fmt.Fprintf(w, "%+v", err)
			}

			return
		}

		// serve mocks
		s.ServeMocks(w, r)
	})
}
