package core

import "net/http"

type Middleware func(next http.Handler) http.HandlerFunc

func PassThroughMiddleware(next http.Handler) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
	})
}

func Chain(middlewares ...Middleware) Middleware {
	return func(final http.Handler) http.HandlerFunc {
		for i := len(middlewares) - 1; i >= 0; i-- {
			final = middlewares[i](final)
		}
		return final.ServeHTTP
	}
}
