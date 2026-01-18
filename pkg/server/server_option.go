package server

import (
	"github.com/amalgamated-tools/bookscraping/pkg/booklore"
	"github.com/amalgamated-tools/bookscraping/pkg/config"
	"github.com/amalgamated-tools/bookscraping/pkg/db"
	"github.com/amalgamated-tools/bookscraping/pkg/goodreads"
)

type ServerOption func(*Server)

func WithConfig(cfg *config.Config) ServerOption {
	return func(s *Server) {
		s.cfg = cfg
	}
}

func WithQueries(queries *db.Queries) ServerOption {
	return func(s *Server) {
		s.queries = queries
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
