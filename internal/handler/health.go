package handler

import (
	"net/http"

	"github.com/Jarimnark/sugar-cane-service/internal/httputil"
)

func handleHealthz(w http.ResponseWriter, r *http.Request) {
	httputil.Respond(w, http.StatusOK, map[string]string{"status": "ok"})
}
