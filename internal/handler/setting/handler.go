package settinghandler

import (
	"encoding/json"
	"net/http"

	"github.com/go-playground/validator/v10"

	"github.com/Jarimnark/sugar-cane-service/internal/domain"
	settingsvc "github.com/Jarimnark/sugar-cane-service/internal/service/setting"
)

var validate = validator.New()

type Handler struct {
	svc settingsvc.Service
}

func New(svc settingsvc.Service) *Handler {
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
