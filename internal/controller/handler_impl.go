package controller

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/zde37/Jusgo/internal/models"
	"github.com/zde37/Jusgo/internal/service"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type handlerImpl struct {
	server   *http.ServeMux
	service  service.ServiceProvider
	validate *validator.Validate
}

func newHandlerImpl(s service.ServiceProvider) *handlerImpl {
	mux := http.NewServeMux()
	handlerImpl := &handlerImpl{
		service:  s,
		server:   mux,
		validate: validator.New(),
	}

	handlerImpl.RegisterRoutes()

	return handlerImpl
}

func (h *handlerImpl) Mux() *http.ServeMux {
	return h.server
}

func (h *handlerImpl) RegisterRoutes() {
	h.server.HandleFunc("/v1/jokes", middleware(h.CreateJoke))
	h.server.HandleFunc("/v1/jokes/{id}", middleware(h.GetJoke))
	h.server.HandleFunc("/v1/jokes/all", middleware(h.GetAllJokes))
	h.server.HandleFunc("/v1/jokes/update/{id}", middleware(h.UpdateJoke))
	h.server.HandleFunc("/v1/jokes/delete/{id}", middleware(h.DeleteJoke))
}

func middleware(f func(http.ResponseWriter, *http.Request) error) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		r = r.WithContext(ctx)

		startTime := time.Now()
		if err := f(w, r); err != nil {
			errRes, statusCode := ErrorInfo(err)

			// Log the error with status
			log.Printf("Log => status: failed, error: %s, status_code: %d, method: %s, path: %s, duration: %s", errRes, statusCode, r.Method, r.RequestURI, time.Since(startTime))

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(statusCode)
			if err = json.NewEncoder(w).Encode(errRes); err != nil {
				log.Printf("failed to write response: %v", err)
			}
			return
		}

		log.Printf("Log => status: success, method: %s, path: %s, duration: %s", r.Method, r.RequestURI, time.Since(startTime))
	}
}

func (h *handlerImpl) CreateJoke(w http.ResponseWriter, r *http.Request) error {
	if r.Method != http.MethodPost {
		return NewErrorStatus(errors.New("method not allowed"), http.StatusMethodNotAllowed)
	}
	var req models.JokeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		if err == io.EOF {
			return NewErrorStatus(errors.New("request body must not be empty"), http.StatusBadRequest)
		}
		return NewErrorStatus(err, http.StatusBadRequest)
	}

	if err := h.validate.Struct(req); err != nil {
		return NewErrorStatus(err, http.StatusBadRequest)
	}

	data := models.Jusgo{
		ID:        primitive.NewObjectID(),
		Joke:      req.Joke,
		UpdatedAt: time.Now(),
		CreatedAt: time.Now(),
	}

	joke, err := h.service.CreateJoke(r.Context(), data)
	if err != nil {
		return NewErrorStatus(err, http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	return json.NewEncoder(w).Encode(joke)
}

func (h *handlerImpl) GetJoke(w http.ResponseWriter, r *http.Request) error {
	if r.Method != http.MethodGet {
		return NewErrorStatus(errors.New("method not allowed"), http.StatusMethodNotAllowed)
	}

	id := r.PathValue("id")
	if id == "" {
		return NewErrorStatus(errors.New("missing joke id"), http.StatusBadRequest)
	}

	joke, err := h.service.GetJoke(r.Context(), id)
	if err != nil {
		return NewErrorStatus(err, http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	return json.NewEncoder(w).Encode(joke)
}

func (h *handlerImpl) GetAllJokes(w http.ResponseWriter, r *http.Request) error {
	if r.Method != http.MethodGet {
		return NewErrorStatus(errors.New("method not allowed"), http.StatusMethodNotAllowed)
	}

	return nil
}

func (h *handlerImpl) UpdateJoke(w http.ResponseWriter, r *http.Request) error {
	if r.Method != http.MethodPatch {
		return NewErrorStatus(errors.New("method not allowed"), http.StatusMethodNotAllowed)
	}

	return nil
}

func (h *handlerImpl) DeleteJoke(w http.ResponseWriter, r *http.Request) error {
	if r.Method != http.MethodDelete {
		return NewErrorStatus(errors.New("method not allowed"), http.StatusMethodNotAllowed)
	}

	return nil
}
