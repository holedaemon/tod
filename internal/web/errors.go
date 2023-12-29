package web

import (
	"net/http"

	"github.com/go-chi/render"
)

type apiError struct {
	Message string `json:"message"`
}

func writeJSONError(w http.ResponseWriter, r *http.Request, err string, status int) {
	render.Status(r, status)
	render.JSON(w, r, &apiError{err})
}
