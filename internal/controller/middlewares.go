package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	_ "github.com/joho/godotenv/autoload"
)

const (
	authorizationHeaderKey  = "authorization"
	authorizationTypeBearer = "bearer"
)

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

func limitMiddleware(rl *rateLimiter, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			http.Error(w, "unable to determine IP", http.StatusInternalServerError)
			return
		}

		limiter := rl.getClient(ip)
		if !limiter.Allow() {
			http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}
