package main

import (
	"demo/controllers"
	"demo/core"
	"demo/middlewares"
	"demo/repositories"
	"demo/services"
	"log"
	"net/http"

	"github.com/joho/godotenv"
	"github.com/lcrux/go-di/v0/di"
)

func init() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	di.Register[core.ServerMuxRouter](core.NewServerMuxRouter, di.Singleton)
	di.Register[controllers.TodoController](controllers.NewTodoController, di.Singleton)
	di.Register[services.TodoService](services.NewTodoService, di.Transient)
	di.Register[repositories.TodoRepository](repositories.NewTodoRepository, di.Singleton)
}

func main() {
	r := core.NewServerMuxRouter()

	todoController, err := di.Resolve[controllers.TodoController]()
	if err != nil {
		panic(err)
	}

	apiGroup := r.WithGroup("api")
	todoController.RegisterRoutes(apiGroup, nil)

	middleware := core.Chain(middlewares.LoggerMiddleware, middlewares.CorsMiddleware)

	log.Println("Starting server on http://localhost:8080")
	if err := http.ListenAndServe(":8080", middleware(r.Handler())); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
