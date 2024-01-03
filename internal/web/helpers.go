package web

import (
	"net/http"

	"github.com/holedaemon/tod/internal/db"
	"golang.org/x/oauth2"
)

func writeError(w http.ResponseWriter, status int) {
	http.Error(w, http.StatusText(status), status)
}

func tokToTok(tok *db.Token) *oauth2.Token {
	return &oauth2.Token{
		AccessToken:  tok.AccessToken,
		TokenType:    tok.TokenType,
		RefreshToken: tok.RefreshToken,
		Expiry:       tok.Expiry,
	}
}
