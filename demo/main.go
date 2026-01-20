package main

import (
	"demo/controllers"
	"demo/repositories"
	"demo/services"
	"log"
	"net/http"

	godi "github.com/lcrux/go-di"
)

func init() {
	godi.Register[controllers.TodoController](controllers.NewTodoController, godi.Singleton)
	godi.Register[services.TodoService](services.NewTodoService, godi.Transient)
	godi.Register[repositories.TodoRepository](repositories.NewTodoRepository, godi.Singleton)
}

func main() {
	r := http.NewServeMux()

	todoController, err := godi.Resolve[controllers.TodoController]()
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
