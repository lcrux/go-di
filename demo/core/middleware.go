package core

import "net/http"

// Middleware defines a function type that takes an http.Handler and returns an http.HandlerFunc.
type Middleware func(next http.Handler) http.HandlerFunc

// PassThroughMiddleware is a middleware that simply passes the request to the next handler without any modification.
func PassThroughMiddleware(next http.Handler) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
	})
}

// Chain combines multiple middlewares into a single middleware.
func Chain(middlewares ...Middleware) Middleware {
	return func(final http.Handler) http.HandlerFunc {
		for i := len(middlewares) - 1; i >= 0; i-- {
			final = middlewares[i](final)
		}
		return final.ServeHTTP
	}
}
