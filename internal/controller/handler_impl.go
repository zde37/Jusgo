package controller

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	_ "github.com/joho/godotenv/autoload"
	"github.com/zde37/Jusgo/internal/models"
	"github.com/zde37/Jusgo/internal/service"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

const (
	authorizationHeaderKey  = "authorization"
	authorizationTypeBearer = "bearer"
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
	h.server.Handle("GET /hello-world", middleware(h.HealthHandler))
	h.server.Handle("POST /jokes", ensureAdmin(middleware(h.CreateJoke))) // admin only
	h.server.Handle("GET /jokes/{id}", middleware(h.GetJoke))
	h.server.Handle("GET /jokes", middleware(h.GetAllJokes))
	h.server.Handle("PATCH /jokes/{id}", ensureAdmin(middleware(h.UpdateJoke)))  // admin only
	h.server.Handle("DELETE /jokes/{id}", ensureAdmin(middleware(h.DeleteJoke))) // admin only

	v1 := http.NewServeMux()
	v1.Handle("/v1/", http.StripPrefix("/v1", h.server))
	h.server = v1
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

func ensureAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authorizationHeader := r.Header.Get(authorizationHeaderKey)
		if len(authorizationHeader) == 0 {
			err := fmt.Errorf("authorization header is not provided")
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		fields := strings.Fields(authorizationHeader)
		if len(fields) < 2 {
			err := fmt.Errorf("invalid authorization header format")
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		authorizationType := strings.ToLower(fields[0])
		if authorizationType != authorizationTypeBearer {
			err := fmt.Errorf("unsupported authorization type %s", authorizationType)
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		accessToken := fields[1]
		if accessToken != os.Getenv("TOKEN") {
			err := fmt.Errorf("invalid token")
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
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
