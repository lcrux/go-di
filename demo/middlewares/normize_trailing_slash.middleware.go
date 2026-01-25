package middlewares

import (
	"demo/core"
	"net/http"
)

// NormalizeTrailingSlashMiddleware is a middleware that removes trailing slashes from the request URL path.
func NormalizeTrailingSlashMiddleware(next http.Handler) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.URL.Path = core.CleanUpTrailingSlash(r.URL.Path)
		next.ServeHTTP(w, r)
	})
}
