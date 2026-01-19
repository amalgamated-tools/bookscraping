package server

import (
	"github.com/amalgamated-tools/bookscraping/pkg/booklore"
	"github.com/amalgamated-tools/bookscraping/pkg/db"
	"github.com/amalgamated-tools/bookscraping/pkg/goodreads"
)

type ServerOption func(*Server)

func WithQuerier(querier db.Querier) ServerOption {
	return func(s *Server) {
		s.queries = querier
	}
}

func WithGoodreadsClient(client *goodreads.Client) ServerOption {
	return func(s *Server) {
		s.grClient = client
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
