package core

// Controller defines the interface for a controller in the application.
type Controller interface {
	// RegisterRoutes registers the routes for the controller with the given router and middleware.
	RegisterRoutes(router ServerMuxRouter, middleware Middleware) error
}
