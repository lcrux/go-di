package controllers

import (
	"demo/core"
	demoerrors "demo/custom-errors"
	"demo/models"
	"demo/services"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
)

func NewTodoController(ts services.TodoService) TodoController {
	return &todoControllerImpl{
		todoService: ts,
	}
}

type TodoController interface {
	core.Controller
	GetTodos(w http.ResponseWriter, r *http.Request)
	CreateTodo(w http.ResponseWriter, r *http.Request)
	CloseTodo(w http.ResponseWriter, r *http.Request)
}

type todoControllerImpl struct {
	todoService services.TodoService
}

func (c *todoControllerImpl) RegisterRoutes(router core.ServerMuxRouter, middleware core.Middleware) error {
	if middleware == nil {
		middleware = core.PassThroughMiddleware
	}

	todosGroup := router.Group("/todos")
	if err := todosGroup.AddPatch("/{id}/done/", middleware(http.HandlerFunc(c.CloseTodo))); err != nil {
		return err
	}
	if err := todosGroup.AddGet("/", middleware(http.HandlerFunc(c.GetTodos))); err != nil {
		return err
	}
	if err := todosGroup.AddPost("/", middleware(http.HandlerFunc(c.CreateTodo))); err != nil {
		return err
	}
	return nil
}

func (c *todoControllerImpl) GetTodos(w http.ResponseWriter, _ *http.Request) {
	// Implement the method for the TodoController interface
	todos, err := c.todoService.GetTodos()
	if err != nil {
		log.Printf("Error getting todos: %v", err)
		handleCustomError(w, err)
		return
	}

	if err := json.NewEncoder(w).Encode(todos); err != nil {
		log.Printf("Error encoding todos: %v", err)
		handleCustomError(w, err)
		return
	}
}

func (c *todoControllerImpl) CreateTodo(w http.ResponseWriter, r *http.Request) {
	var todo models.CreateTodoRequest
	if err := json.NewDecoder(r.Body).Decode(&todo); err != nil {
		log.Printf("Error decoding todo: %v", err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	newTodo, err := c.todoService.CreateTodo(&todo)
	if err != nil {
		handleCustomError(w, err)
		return
	}

	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(newTodo); err != nil {
		log.Printf("Error encoding todo: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func (c *todoControllerImpl) CloseTodo(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	todoID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		log.Printf("Error converting id to integer: %v", err)
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	var todo models.CloseTodoRequest
	if err := json.NewDecoder(r.Body).Decode(&todo); err != nil {
		log.Printf("Error decoding todo: %v", err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	if todoID != uint64(todo.ID) {
		log.Printf("ID in path does not match ID in request body: %d != %d", todoID, todo.ID)
		http.Error(w, "ID mismatch", http.StatusBadRequest)
		return
	}

	updatedTodo, err := c.todoService.CloseTodo(todo.ID)
	if err != nil {
		handleCustomError(w, err)
		return
	}

	if err := json.NewEncoder(w).Encode(updatedTodo); err != nil {
		log.Printf("Error encoding updated todo: %v", err)
		handleCustomError(w, err)
		return
	}
}

func handleCustomError(w http.ResponseWriter, err error) {
	log.Printf("Error: %v", err)

	if _, ok := err.(*demoerrors.NotFoundError); ok {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	if _, ok := err.(*demoerrors.ValidationError); ok {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	http.Error(w, "Internal Server Error", http.StatusInternalServerError)
}
