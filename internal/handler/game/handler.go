package gamehandler

import (
	"encoding/json"
	"net/http"

	"github.com/go-playground/validator/v10"

	"github.com/Jarimnark/sugar-cane-service/internal/domain"
	gamesvc "github.com/Jarimnark/sugar-cane-service/internal/service/game"
)

var validate = validator.New()

type Handler struct {
	svc gamesvc.Service
}

func New(svc gamesvc.Service) *Handler {
	return &Handler{svc: svc}
}

func decode[T any](r *http.Request) (T, error) {
	var v T
	if err := json.NewDecoder(r.Body).Decode(&v); err != nil {
		return v, domain.BadRequest("invalid request body", err)
	}
	if err := validate.Struct(v); err != nil {
		return v, domain.BadRequest(err.Error(), err)
	}
	return v, nil
}
