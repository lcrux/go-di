package core

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"slices"
	"strings"
)

// DO NOT confuse with CORS allowed methods
var routerHttpMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE"}

// CleanUpTrailingSlash trims spaces and removes trailing slashes from the URL path if it is longer than 1 character.
func CleanUpTrailingSlash(p string) string {
	p = strings.TrimSpace(p)
	if len(p) > 1 && strings.HasSuffix(p, "/") {
		p = strings.TrimSuffix(p, "/")
	}
	return p
}

// JoinUrlPath joins the base path with the provided elements and returns the resulting path.
// It trims any trailing slashes from the resulting path if it is longer than 1 character.
func JoinUrlPath(base string, elem ...string) (string, error) {
	if len(elem) == 0 {
		return CleanUpTrailingSlash(base), nil
	}

	pattern, err := url.JoinPath(base, elem...)
	if err != nil {
		panic(err)
	}

	pattern, err = url.PathUnescape(pattern)
	if err != nil {
		panic(err)
	}

	pattern = CleanUpTrailingSlash(pattern)

	return pattern, nil
}

// NewServerMuxRouter creates a new instance of ServerMuxRouter with an empty group and a new http.ServeMux.
func NewServerMuxRouter() ServerMuxRouter {
	return &serverMuxRouterImpl{
		mux:   http.NewServeMux(),
		group: "",
	}
}

// ServerMuxRouter defines the interface for a router that can group routes, handle HTTP requests, and add controllers and routes with specific HTTP methods.
type ServerMuxRouter interface {
	Group(string) ServerMuxRouter
	Handler() http.Handler
	AddController(Controller, Middleware) error
	AddGet(string, http.HandlerFunc) error
	AddPost(string, http.HandlerFunc) error
	AddPatch(string, http.HandlerFunc) error
	AddDelete(string, http.HandlerFunc) error
}

// serverMuxRouterImpl is the concrete implementation of the ServerMuxRouter interface.
type serverMuxRouterImpl struct {
	mux   *http.ServeMux
	group string
}

// Handler returns the underlying http.Handler for the router.
func (r *serverMuxRouterImpl) Handler() http.Handler {
	return r.mux
}

// Group creates a new sub-router with the given group prefix.
func (r *serverMuxRouterImpl) Group(group string) ServerMuxRouter {
	group = strings.TrimSpace(group)
	if group == "" {
		return r
	}

	newGroup, err := JoinUrlPath("/", r.group, group)
	if err != nil {
		panic(err)
	}

	return &serverMuxRouterImpl{
		group: newGroup,
		mux:   r.mux,
	}
}

// AddController registers the routes for the given controller with the router and middleware.
func (r *serverMuxRouterImpl) AddController(c Controller, m Middleware) error {
	return c.RegisterRoutes(r, m)
}

// AddGet registers a GET route with the given pattern and handler.
func (r *serverMuxRouterImpl) AddGet(pattern string, handler http.HandlerFunc) error {
	pattern, err := JoinUrlPath(r.group, pattern)
	if err != nil {
		return err
	}
	return r.addRouteWithMethod("GET", pattern, handler)
}

// AddPost registers a POST route with the given pattern and handler.
func (r *serverMuxRouterImpl) AddPost(pattern string, handler http.HandlerFunc) error {
	pattern, err := JoinUrlPath(r.group, pattern)
	if err != nil {
		return err
	}
	return r.addRouteWithMethod("POST", pattern, handler)
}

// AddPatch registers a PATCH route with the given pattern and handler.
func (r *serverMuxRouterImpl) AddPatch(pattern string, handler http.HandlerFunc) error {
	pattern, err := JoinUrlPath(r.group, pattern)
	if err != nil {
		return err
	}
	return r.addRouteWithMethod("PATCH", pattern, handler)
}

// AddDelete registers a DELETE route with the given pattern and handler.
func (r *serverMuxRouterImpl) AddDelete(pattern string, handler http.HandlerFunc) error {
	pattern, err := JoinUrlPath(r.group, pattern)
	if err != nil {
		return err
	}
	return r.addRouteWithMethod("DELETE", pattern, handler)
}

// addRouteWithMethod adds a route with the specified HTTP method, path, and handler to the router.
func (r *serverMuxRouterImpl) addRouteWithMethod(method string, path string, handler http.HandlerFunc) error {
	if method == "" {
		return fmt.Errorf("method cannot be empty")
	}
	if !slices.Contains(routerHttpMethods, strings.ToUpper(method)) {
		return fmt.Errorf("method is not supported, must be one of [%s], got %s", strings.Join(routerHttpMethods, ", "), method)
	}
	if strings.TrimSpace(path) == "" {
		return fmt.Errorf("path cannot be empty")
	}
	if !strings.HasPrefix(path, "/") {
		return fmt.Errorf("path must start with a '/', got %s", path)
	}
	if handler == nil {
		return fmt.Errorf("handler cannot be nil")
	}

	routerPath := fmt.Sprintf("%s %s", method, path)
	r.mux.HandleFunc(routerPath, handler)

	log.Printf("Added route: %s\n", routerPath)
	return nil
}
