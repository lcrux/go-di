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

// cleanUrlPath trims spaces and removes trailing slashes from the URL path if it is longer than 1 character.
func cleanUrlPath(p string) string {
	p = strings.TrimSpace(p)
	if len(p) > 1 {
		p = strings.TrimSuffix(p, "/")
	}
	return p
}

// joinUrlPath joins the base path with the provided elements and returns the resulting path.
// It trims any trailing slashes from the resulting path if it is longer than 1 character.
func joinUrlPath(base string, elem ...string) (string, error) {
	if len(elem) == 0 {
		return cleanUrlPath(base), nil
	}

	pattern, err := url.JoinPath(base, elem...)
	if err != nil {
		panic(err)
	}

	pattern, err = url.PathUnescape(pattern)
	if err != nil {
		panic(err)
	}

	pattern = cleanUrlPath(pattern)

	return pattern, nil
}

func NewServerMuxRouter() ServerMuxRouter {
	return &serverMuxRouterImpl{
		mux:   http.NewServeMux(),
		group: "",
	}
}

type ServerMuxRouter interface {
	WithGroup(group string) ServerMuxRouter
	Handler() http.Handler
	Get(pattern string, handler http.HandlerFunc)
	Post(pattern string, handler http.HandlerFunc)
	Patch(pattern string, handler http.HandlerFunc)
	Delete(pattern string, handler http.HandlerFunc)
}

type serverMuxRouterImpl struct {
	mux   *http.ServeMux
	group string
}

func (r *serverMuxRouterImpl) Handler() http.Handler {
	return r.mux
}

func (r *serverMuxRouterImpl) addRouteWithMethod(method string, path string, handler http.HandlerFunc) {
	if method == "" {
		panic("method cannot be empty")
	}
	if !slices.Contains(routerHttpMethods, strings.ToUpper(method)) {
		panic(fmt.Sprintf("method is not supported, must be one of [%s], got %s", strings.Join(routerHttpMethods, ", "), method))
	}
	if path == "" {
		panic("path cannot be empty")
	}
	if handler == nil {
		panic("handler cannot be nil")
	}

	routerPath := fmt.Sprintf("%s %s", method, path)

	log.Printf("Adding route: %s\n", routerPath)

	r.mux.HandleFunc(routerPath, handler)
}

func (r *serverMuxRouterImpl) WithGroup(group string) ServerMuxRouter {
	if strings.TrimSpace(group) == "" {
		return r
	}

	newGroup, err := joinUrlPath("/", r.group, "/", group)
	if err != nil {
		panic(err)
	}

	newGroup = strings.TrimSpace(newGroup)
	if len(newGroup) > 1 {
		newGroup = strings.TrimSuffix(newGroup, "/")
	}

	return &serverMuxRouterImpl{
		group: newGroup,
		mux:   r.mux,
	}
}

func (r *serverMuxRouterImpl) Get(pattern string, handler http.HandlerFunc) {
	pattern, err := joinUrlPath(r.group, pattern)
	if err != nil {
		panic(err)
	}
	r.addRouteWithMethod("GET", pattern, handler)
}

func (r *serverMuxRouterImpl) Post(pattern string, handler http.HandlerFunc) {
	pattern, err := joinUrlPath(r.group, pattern)
	if err != nil {
		panic(err)
	}
	r.addRouteWithMethod("POST", pattern, handler)
}

func (r *serverMuxRouterImpl) Patch(pattern string, handler http.HandlerFunc) {
	pattern, err := joinUrlPath(r.group, pattern)
	if err != nil {
		panic(err)
	}
	r.addRouteWithMethod("PATCH", pattern, handler)
}

func (r *serverMuxRouterImpl) Delete(pattern string, handler http.HandlerFunc) {
	pattern, err := joinUrlPath(r.group, pattern)
	if err != nil {
		panic(err)
	}
	r.addRouteWithMethod("DELETE", pattern, handler)
}
