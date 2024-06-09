package controller

import (
	"net/http"

	"github.com/zde37/Jusgo/internal/service"
)

type HandlerProvider interface {
	Mux() *http.ServeMux
	CreateJoke(w http.ResponseWriter, r *http.Request) error
	GetJoke(w http.ResponseWriter, r *http.Request) error
	GetAllJokes(w http.ResponseWriter, r *http.Request) error
	UpdateJoke(w http.ResponseWriter, r *http.Request) error
	DeleteJoke(w http.ResponseWriter, r *http.Request) error
}

type Handler struct {
	Hndl HandlerProvider
}

func NewHandler(s service.ServiceProvider) *Handler  {
	return &Handler{
		Hndl: newHandlerImpl(s),
	}
}