package main

import (
	"context"
	"demo/controllers"
	"demo/core"
	"demo/middlewares"
	"demo/repositories"
	"demo/services"
	"demo/utils"
	"fmt"
	"log"
	"net/http"

	"github.com/joho/godotenv"
	"github.com/lcrux/go-di/di"
)

func main() {
	r := core.NewServerMuxRouter()
	container := di.NewContainer()
	defer container.Shutdown()

	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	// Register dependencies
	if err := registerServices(container); err != nil {
		log.Fatalf("Failed to register services: %v", err)
	}

	// Resolve the TodoController
	todoController := di.Resolve[controllers.TodoController](container, nil)
	if todoController == nil {
		panic("failed to resolve TodoController")
	}

	// Add the TodoController to the API router
	apiRouter := r.Group("api")
	if err := apiRouter.AddController(todoController, nil); err != nil {
		log.Fatalf("Failed to add controller to API router: %v", err)
	}

	// Create a new lifecycle context and resolve the TodoController within that context
	cctx := container.NewContext()
	// Resolve the TodoController within the new lifecycle context
	todoController2 := di.Resolve[controllers.TodoController](container, cctx)
	if todoController2 == nil {
		panic("failed to resolve TodoController with lifecycle context")
	}

	// Add the TodoController to the API2 router
	api2Router := r.Group("api2")
	if err := api2Router.AddController(todoController2, nil); err != nil {
		log.Fatalf("Failed to add controller to API2 router: %v", err)
	}

	// Chain middlewares
	middleware := core.Chain(
		middlewares.CorsMiddleware,
		middlewares.LoggerMiddleware,
		middlewares.NormalizeTrailingSlashMiddleware,
	)

	server := &http.Server{
		Addr:    ":8080",
		Handler: middleware(r.Handler()),
	}

	// Set up a shutdown handler to gracefully shut down the server and DI container when an interrupt signal is received.
	_ = utils.ShutdownHandler(func() {
		log.Println("Shutting down DI lifecycle contexts...")
		if err := container.Shutdown(); err != nil {
			log.Printf("Error during DI shutdown: %v", err)
		}

		log.Println("Shutting down server...")
		if err := server.Shutdown(context.Background()); err != nil {
			log.Printf("Error during server shutdown: %v", err)
		}
	})

	// Start the HTTP server
	log.Println("Starting server on http://localhost:8080")
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Failed to start server: %v", err)
		return
	}
}

func registerServices(container di.Container) error {
	if err := di.Register[controllers.TodoController](container, controllers.NewTodoController, di.Scoped); err != nil {
		return fmt.Errorf("Failed to register TodoController: %v", err)
	}
	if err := di.Register[services.TodoService](container, services.NewTodoService, di.Transient); err != nil {
		return fmt.Errorf("Failed to register TodoService: %v", err)
	}
	if err := di.Register[repositories.TodoRepository](container, repositories.NewTodoRepository, di.Singleton); err != nil {
		return fmt.Errorf("Failed to register TodoRepository: %v", err)
	}
	return nil
}
