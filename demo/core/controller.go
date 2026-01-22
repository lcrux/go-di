package core

type Controller interface {
	RegisterRoutes(router ServerMuxRouter, middleware Middleware)
}

type controllerImpl struct {
}

func (c *controllerImpl) RegisterRoutes(router ServerMuxRouter, middleware Middleware) {
	panic("method not implemented")
}
