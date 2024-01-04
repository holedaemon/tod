package main

import (
	"context"
	"fmt"
	"os"

	"github.com/caarlos0/env/v7"
	"github.com/holedaemon/tod/internal/db"
	"github.com/holedaemon/tod/internal/web"
	"github.com/zikaeroh/ctxlog"
	spotifyauth "github.com/zmb3/spotify/v2/auth"
	"go.uber.org/zap"
)

type options struct {
	Addr                string   `env:"TOD_ADDR"`
	SpotifyClientID     string   `env:"TOD_SPOTIFY_CLIENT_ID"`
	SpotifyClientSecret string   `env:"TOD_SPOTIFY_CLIENT_SECRET"`
	SpotifyScopes       []string `env:"TOD_SPOTIFY_SCOPES"`
	SpotifyRedirectURL  string   `env:"TOD_SPOTIFY_REDIRECT_URL"`
}

func main() {
	opts := &options{}
	eo := env.Options{
		RequiredIfNoDef: true,
	}

	if err := env.Parse(opts, eo); err != nil {
		fmt.Fprintf(os.Stderr, "error parsing env variables into struct: %s\n", err.Error())
		return
	}

	logger := ctxlog.New(false)
	ctx := ctxlog.WithLogger(context.Background(), logger)

	file, err := db.Open()
	if err != nil {
		logger.Fatal("error opening db file", zap.Error(err))
	}

	defer file.Close()

	auth := spotifyauth.New(
		spotifyauth.WithClientID(opts.SpotifyClientID),
		spotifyauth.WithClientSecret(opts.SpotifyClientSecret),
		spotifyauth.WithScopes(opts.SpotifyScopes...),
		spotifyauth.WithRedirectURL(opts.SpotifyRedirectURL),
	)

	srv, err := web.New(
		web.WithAddr(opts.Addr),
		web.WithAuth(auth),
		web.WithDB(file),
	)
	if err != nil {
		logger.Fatal("error creating server", zap.Error(err))
	}

	if err := srv.Run(ctx); err != nil {
		logger.Fatal("error starting server", zap.Error(err))
	}
}
