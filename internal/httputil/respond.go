package httputil

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/Jarimnark/sugar-cane-service/internal/domain"
)

func Respond(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if body != nil {
		if err := json.NewEncoder(w).Encode(body); err != nil {
			slog.Error("encode response", "error", err)
		}
	}
}

func RespondError(w http.ResponseWriter, err error) {
	if ae, ok := domain.AsAppError(err); ok {
		if ae.Err != nil {
			slog.Error("app error", "error", ae.Err, "message", ae.Message)
		}
		Respond(w, ae.Code, map[string]string{"error": ae.Message})
		return
	}
	slog.Error("unexpected error", "error", err)
	Respond(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
}
