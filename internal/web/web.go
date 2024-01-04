package web

import (
	"context"
	"fmt"
	"net"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/gofrs/uuid/v5"
	"github.com/holedaemon/tod/internal/db"
	"github.com/zikaeroh/ctxlog"
	"github.com/zmb3/spotify/v2"
	spotifyauth "github.com/zmb3/spotify/v2/auth"
	"go.uber.org/zap"
)

const defaultAddr = ":8080"

// Server is a web server.
type Server struct {
	addr string
	db   *db.DB
	auth *spotifyauth.Authenticator

	states map[string]struct{}
}

// New creates a new server.
func New(opts ...Option) (*Server, error) {
	s := &Server{
		states: map[string]struct{}{},
	}

	for _, o := range opts {
		o(s)
	}

	if s.addr == "" {
		s.addr = defaultAddr
	}

	if s.db == nil {
		return nil, fmt.Errorf("web: missing db")
	}

	if s.auth == nil {
		return nil, fmt.Errorf("web: missing auth")
	}

	return s, nil
}

func (s *Server) Run(ctx context.Context) error {
	r := chi.NewMux()

	r.Use(recoverer)

	logger := ctxlog.FromContext(ctx)
	r.Use(requestLogger(logger))

	r.Get("/", s.getSpotify)
	r.Get("/callback", s.getSpotifyCallback)
	r.Get("/np/{id}", s.getNowPlaying)
	r.Get("/overlay/{id}", s.getOverlay)

	srv := &http.Server{
		Addr:        s.addr,
		Handler:     r,
		BaseContext: func(l net.Listener) context.Context { return ctx },
	}

	go func() {
		<-ctx.Done()
		if err := srv.Shutdown(context.Background()); err != nil {
			ctxlog.Error(ctx, "error shutting down server", zap.Error(err))
			return
		}
	}()

	ctxlog.Info(ctx, "server listening...")
	return srv.ListenAndServe()
}

func (s *Server) getSpotify(w http.ResponseWriter, r *http.Request) {
	id := uuid.Must(uuid.NewV4()).String()
	s.states[id] = struct{}{}
	url := s.auth.AuthURL(id)
	http.Redirect(w, r, url, http.StatusSeeOther)
}

func (s *Server) getSpotifyCallback(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	st := r.FormValue("state")
	if st == "" {
		writeError(w, http.StatusBadRequest)
		return
	}

	if _, ok := s.states[st]; !ok {
		writeError(w, http.StatusBadRequest)
		return
	}

	tok, err := s.auth.Token(r.Context(), st, r)
	if err != nil {
		writeError(w, http.StatusBadRequest)
		return
	}

	delete(s.states, st)

	client := spotify.New(s.auth.Client(r.Context(), tok))
	user, err := client.CurrentUser(r.Context())
	if err != nil {
		ctxlog.Error(ctx, "error fetching user from Spotify", zap.Error(err))
		writeError(w, http.StatusInternalServerError)
		return
	}

	dbTok := &db.Token{
		AccessToken:  tok.AccessToken,
		RefreshToken: tok.RefreshToken,
		Expiry:       tok.Expiry,
		TokenType:    tok.TokenType,
	}

	if err := s.db.SetToken(user.ID, dbTok); err != nil {
		ctxlog.Error(ctx, "error setting token", zap.Error(err))
		writeError(w, http.StatusInternalServerError)
		return
	}

	url := fmt.Sprintf("/overlay/%s", user.ID)
	http.Redirect(w, r, url, http.StatusSeeOther)
}

type nowPlaying struct {
	Artist string `json:"artist"`
	Album  string `json:"album"`
	Song   string `json:"song"`
}

func (s *Server) getNowPlaying(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	ctx := r.Context()

	t, err := s.db.Token(id)
	if err != nil {
		ctxlog.Error(ctx, "error fetching token from db", zap.Error(err))
		writeJSONError(w, r, "error fetching token from db", http.StatusInternalServerError)
		return
	}

	if t == nil {
		writeJSONError(w, r, "not found", http.StatusNotFound)
		return
	}

	tok := tokToTok(t)
	client := spotify.New(s.auth.Client(r.Context(), tok))

	playing, err := client.PlayerCurrentlyPlaying(r.Context())
	if err != nil {
		ctxlog.Error(ctx, "error fetching now playing from spotify", zap.Error(err))
		writeJSONError(w, r, "error fetching now playing from spotify", http.StatusInternalServerError)
		return
	}

	if playing.Playing {
		var artist string
		if len(playing.Item.Artists) > 1 {
			artist = "Various Artists"
		} else {
			artist = playing.Item.Artists[0].Name
		}

		np := &nowPlaying{
			Artist: artist,
			Album:  playing.Item.Album.Name,
			Song:   playing.Item.Name,
		}

		render.Status(r, http.StatusOK)
		render.JSON(w, r, np)
	} else {
		render.NoContent(w, r)
	}

	cliTok, err := client.Token()
	if err != nil {
		ctxlog.Error(ctx, "error checking if token has been refreshed", zap.Error(err))
		return
	}

	if cliTok.Expiry.After(tok.Expiry) {
		newTok := &db.Token{
			AccessToken:  cliTok.AccessToken,
			RefreshToken: cliTok.RefreshToken,
			Expiry:       cliTok.Expiry,
			TokenType:    cliTok.TokenType,
		}

		if err := s.db.SetToken(id, newTok); err != nil {
			ctxlog.Error(ctx, "error updating db token", zap.Error(err), zap.String("id", id))
			return
		}
	}
}
