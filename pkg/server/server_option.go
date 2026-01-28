package server

import (
	"github.com/amalgamated-tools/bookscraping/pkg/booklore"
	"github.com/amalgamated-tools/bookscraping/pkg/db"
)

type ServerOption func(*Server)

func WithQuerier(querier db.Querier) ServerOption {
	return func(s *Server) {
		s.queries = querier
	}
}

func WithBookloreClient(client *booklore.Client) ServerOption {
	return func(s *Server) {
		s.blClient = client
	}
}

func WithAddr(addr string) ServerOption {
	return func(s *Server) {
		s.addr = addr
	}
}
