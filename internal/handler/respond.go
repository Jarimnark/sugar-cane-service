package handler

import (
	"net/http"

	"github.com/Jarimnark/sugar-cane-service/internal/httputil"
)

// Respond and RespondError are thin wrappers kept here for handlers that live
// directly in the handler package (e.g. health).
func Respond(w http.ResponseWriter, status int, body any) {
	httputil.Respond(w, status, body)
}

func RespondError(w http.ResponseWriter, err error) {
	httputil.RespondError(w, err)
}
