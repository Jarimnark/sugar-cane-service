package listhandler

import (
	listsvc "github.com/Jarimnark/sugar-cane-service/internal/service/list"
)

type Handler struct {
	svc listsvc.Service
}

func New(svc listsvc.Service) *Handler {
	return &Handler{svc: svc}
}
