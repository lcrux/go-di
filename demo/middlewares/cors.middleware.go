package middlewares

import (
	"log"
	"net/http"
	"os"
	"strings"
)

var allowedOrigins = os.Getenv("CORS_ALLOWED_ORIGINS")
var allowedMethods = os.Getenv("CORS_ALLOWED_METHODS")
var allowedHeaders = os.Getenv("CORS_ALLOWED_HEADERS")

func isItAllowed(key string, allowed string) bool {
	if allowed == "*" || allowed == "" {
		return true
	}
	if key == "" {
		return true
	}
	items := strings.Split(allowed, ",")
	for _, item := range items {
		if strings.TrimSpace(item) == key {
			return true
		}
	}
	return false
}

func isOriginAllowed(origin string) bool {
	if allowedOrigins == "*" {
		return true
	}

	items := strings.Split(allowedOrigins, ",")
	for _, item := range items {
		if strings.HasPrefix(origin, strings.TrimSpace(item)) {
			return true
		}
	}
	return false
}

func isMethodAllowed(method string) bool {
	return isItAllowed(method, allowedMethods)
}

func isHeaderAllowed(header string) bool {
	return isItAllowed(header, allowedHeaders)
}

func CorsMiddleware(next http.Handler) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Validate allowed origins
		origin := r.Header.Get("Origin")
		if origin == "" {
			origin = r.RemoteAddr
		}
		if !isOriginAllowed(origin) {
			log.Printf("[CORS] Blocked request from origin: %s", origin)
			http.Error(w, "Origin not allowed", http.StatusForbidden)
			return
		}
		w.Header().Set("Access-Control-Allow-Origin", origin)

		// Allow credentials if needed
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		// Restrict allowed methods
		method := r.Header.Get("Access-Control-Request-Method")
		if method == "" {
			method = r.Method
		}
		if !isMethodAllowed(method) {
			log.Printf("[CORS] Blocked request with method: %s from origin: %s", method, origin)
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Access-Control-Allow-Methods", allowedMethods)

		// Restrict allowed headers
		requestedHeaders := r.Header.Get("Access-Control-Request-Headers")
		for _, header := range strings.Split(requestedHeaders, ",") {
			if !isHeaderAllowed(strings.TrimSpace(header)) {
				log.Printf("[CORS] Blocked request with header: %s from origin: %s", header, origin)
				http.Error(w, "Header not allowed", http.StatusForbidden)
				return
			}
		}
		w.Header().Set("Access-Control-Allow-Headers", allowedHeaders)

		// Expose specific headers to the client
		w.Header().Set("Access-Control-Expose-Headers", "Authorization")

		// Cache preflight responses
		w.Header().Set("Access-Control-Max-Age", "86400") // Cache for 1 day

		// Handle preflight requests
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		// Pass the request to the next handler
		next.ServeHTTP(w, r)
	})
}
