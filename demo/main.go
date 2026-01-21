package main

import (
	"demo/controllers"
	"demo/repositories"
	"demo/services"
	"log"
	"net/http"

	"github.com/lcrux/go-di/v0/di"
)

func init() {
	di.Register[controllers.TodoController](controllers.NewTodoController, di.Singleton)
	di.Register[services.TodoService](services.NewTodoService, di.Transient)
	di.Register[repositories.TodoRepository](repositories.NewTodoRepository, di.Singleton)
}

func main() {
	r := http.NewServeMux()

	todoController, err := di.Resolve[controllers.TodoController]()
	if err != nil {
		panic(err)
	}

	r.HandleFunc("PATCH /todos/{id}/done", todoController.CloseTodo)
	r.HandleFunc("GET /todos", todoController.GetTodos)
	r.HandleFunc("POST /todos", todoController.CreateTodo)

	log.Println("Starting server on http://localhost:8080")
	if err := http.ListenAndServe(":8080", LoggerMiddleware(r)); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func LoggerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Request: %s %s (source: %s)", r.Method, r.URL.Path, r.RemoteAddr)
		next.ServeHTTP(w, r)
	})
}
