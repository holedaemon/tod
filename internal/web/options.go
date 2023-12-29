package web

import (
	"github.com/holedaemon/eva-music/internal/db"
	spotifyauth "github.com/zmb3/spotify/v2/auth"
)

// Option configures a server.
type Option func(*Server)

// WithAddr sets a server's address.
func WithAddr(addr string) Option {
	return func(s *Server) {
		s.addr = addr
	}
}

// WithDB sets a server's DB.
func WithDB(db *db.DB) Option {
	return func(s *Server) {
		s.db = db
	}
}

// WithAuth sets a server's Spotify authenticator.
func WithAuth(auth *spotifyauth.Authenticator) Option {
	return func(s *Server) {
		s.auth = auth
	}
}
