package devhandler

import (
	"encoding/json"
	"net/http"

	"github.com/go-playground/validator/v10"

	"github.com/Jarimnark/sugar-cane-service/internal/domain"
	devsvc "github.com/Jarimnark/sugar-cane-service/internal/service/dev"
	"github.com/Jarimnark/sugar-cane-service/internal/storage"
)

var validate = validator.New()

type Handler struct {
	svc     devsvc.Service
	storage *storage.Client
}

func New(svc devsvc.Service, storage *storage.Client) *Handler {
	return &Handler{svc: svc, storage: storage}
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
