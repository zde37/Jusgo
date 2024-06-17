package controller

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/zde37/Jusgo/internal/models"
	"github.com/zde37/Jusgo/internal/service"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
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
	rl := newRateLimiter() 

	h.server.Handle("GET /hello-world", middleware(h.HealthHandler))
	h.server.Handle("POST /jokes", limitMiddleware(rl, ensureAdmin(middleware(h.CreateJoke)))) // admin only
	h.server.Handle("GET /jokes/{id}", limitMiddleware(rl, middleware(h.GetJoke)))
	h.server.Handle("GET /jokes", limitMiddleware(rl, middleware(h.GetAllJokes)))
	h.server.Handle("PATCH /jokes/{id}", ensureAdmin(middleware(h.UpdateJoke)))  // admin only
	h.server.Handle("DELETE /jokes/{id}", ensureAdmin(middleware(h.DeleteJoke))) // admin only

	v1 := http.NewServeMux()
	v1.Handle("/v1/", http.StripPrefix("/v1", h.server))
	h.server = v1
}

func (h *handlerImpl) HealthHandler(w http.ResponseWriter, r *http.Request) error {
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte("Hello world")); err != nil {
		return NewErrorStatus(err, http.StatusInternalServerError)
	}
	return nil
}

func (h *handlerImpl) CreateJoke(w http.ResponseWriter, r *http.Request) error {
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
	id := r.PathValue("id")
	if id == "" {
		return NewErrorStatus(errors.New("id is required"), http.StatusBadRequest)
	}

	joke, err := h.service.GetJoke(r.Context(), id)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return NewErrorStatus(err, http.StatusNotFound)
		}
		return NewErrorStatus(err, http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	return json.NewEncoder(w).Encode(joke)
}

func (h *handlerImpl) GetAllJokes(w http.ResponseWriter, r *http.Request) error {
	page, limit, err := parsePaginationParams(r)
	if err != nil {
		return NewErrorStatus(err, http.StatusBadRequest)
	}

	jokes, err := h.service.GetAllJokes(r.Context(), page, limit)
	if err != nil {
		return NewErrorStatus(err, http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	return json.NewEncoder(w).Encode(jokes)
}

func parsePaginationParams(r *http.Request) (int, int, error) {
	page := r.URL.Query().Get("page")
	limit := r.URL.Query().Get("limit")

	p := 1
	l := 10

	var err error
	if page != "" {
		p, err = strconv.Atoi(page)
		if err != nil || p < 1 {
			return 0, 0, errors.New("invalid page number")
		}
	}

	if limit != "" {
		l, err = strconv.Atoi(limit)
		if err != nil || l < 1 {
			return 0, 0, errors.New("invalid limit number")
		}
	}

	return p, l, nil
}

func (h *handlerImpl) UpdateJoke(w http.ResponseWriter, r *http.Request) error {
	id := r.PathValue("id")
	if id == "" {
		return NewErrorStatus(errors.New("id is required"), http.StatusBadRequest)
	}

	var req models.JokeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		if err == io.EOF {
			return NewErrorStatus(errors.New("request body must not be empty"), http.StatusBadRequest)
		}
		return NewErrorStatus(err, http.StatusBadRequest)
	}

	if err := h.validate.Struct(&req); err != nil {
		return NewErrorStatus(err, http.StatusBadRequest)
	}

	objectID, err := primitive.ObjectIDFromHex(id) // convert id to mongodb primitive ID
	if err != nil {
		return NewErrorStatus(errors.New("failed to create object id"), http.StatusInternalServerError)
	}

	updatedJoke, err := h.service.UpdateJoke(r.Context(), models.Jusgo{
		ID:        objectID,
		Joke:      req.Joke,
		UpdatedAt: time.Now(),
	})
	if err != nil {
		return NewErrorStatus(err, http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	return json.NewEncoder(w).Encode(updatedJoke)
}

func (h *handlerImpl) DeleteJoke(w http.ResponseWriter, r *http.Request) error {
	id := r.PathValue("id")
	if id == "" {
		return NewErrorStatus(errors.New("id is required"), http.StatusBadRequest)
	}

	// check if joke exists first
	if _, err := h.service.GetJoke(r.Context(), id); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return NewErrorStatus(err, http.StatusNotFound)
		}
		return NewErrorStatus(err, http.StatusInternalServerError)
	}

	if err := h.service.DeleteJoke(r.Context(), id); err != nil {
		return NewErrorStatus(err, http.StatusInternalServerError)
	}

	w.WriteHeader(http.StatusOK)
	return nil
}
