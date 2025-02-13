package router

import (
	"assignment-imesh/http/controllers"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"golang.org/x/time/rate"
)

// RateLimiterMiddleware limits the number of requests per second
func RateLimiterMiddleware(next http.Handler) http.Handler {
	limiter := rate.NewLimiter(rate.Every(time.Second), 10) // Allow 10 requests per second

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !limiter.Allow() {
			http.Error(w, "Too many requests", http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func Init() *mux.Router {
	// New mux router
	router := mux.NewRouter()

	// Set up routes with rate limiting middleware
	router.HandleFunc("/convert", controllers.ConvertCurrencyHandler).Methods("GET")
	router.Use(RateLimiterMiddleware)

	return router
}
