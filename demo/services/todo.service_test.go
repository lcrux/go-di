package services

import (
	"demo/models"
	"demo/repositories"
	"errors"
	"testing"
	"time"
)

type fakeTodoRepo struct {
	created []*models.CreateTodoRequest
	closed  []uint32
	Todos   []models.Todo
}

func (f *fakeTodoRepo) GetTodos() ([]models.Todo, error) {
	return f.Todos, nil
}

func (f *fakeTodoRepo) CreateTodo(todo *models.CreateTodoRequest) (models.Todo, error) {
	if todo == nil {
		return models.Todo{}, errors.New("nil todo")
	}
	f.created = append(f.created, todo)
	return models.Todo{ID: 1, Title: todo.Title}, nil
}

func (f *fakeTodoRepo) CloseTodo(id uint32) (models.Todo, error) {
	f.closed = append(f.closed, id)
	return models.Todo{ID: id, Done: true}, nil
}

func TestTodoService_CreateTodoValidation(t *testing.T) {
	repo := &fakeTodoRepo{}
	service := NewTodoService(repo)

	if _, err := service.CreateTodo(nil); err == nil {
		t.Fatal("expected error for nil todo")
	}

	if _, err := service.CreateTodo(&models.CreateTodoRequest{Title: ""}); err == nil {
		t.Fatal("expected error for empty title")
	}

	longTitle := make([]byte, 256)
	for i := range longTitle {
		longTitle[i] = 'a'
	}
	if _, err := service.CreateTodo(&models.CreateTodoRequest{Title: string(longTitle)}); err == nil {
		t.Fatal("expected error for long title")
	}

	if _, err := service.CreateTodo(&models.CreateTodoRequest{Title: "Test", DueDate: time.Now().Add(-time.Hour)}); err == nil {
		t.Fatal("expected error for past due date")
	}
}

func TestTodoService_CreateTodoAndCloseTodo(t *testing.T) {
	repo := &fakeTodoRepo{}
	service := NewTodoService(repo)

	todo, err := service.CreateTodo(&models.CreateTodoRequest{Title: "Valid"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if todo.Title != "Valid" {
		t.Fatalf("expected title 'Valid', got %q", todo.Title)
	}
	if len(repo.created) != 1 {
		t.Fatalf("expected repo to be called once, got %d", len(repo.created))
	}

	if _, err := service.CloseTodo(0); err == nil {
		t.Fatal("expected error for invalid id")
	}

	closed, err := service.CloseTodo(10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !closed.Done || closed.ID != 10 {
		t.Fatalf("expected closed todo with id 10, got %+v", closed)
	}
}

var _ repositories.TodoRepository = (*fakeTodoRepo)(nil)
