package repositories

import (
	"demo/models"
	"testing"
)

func resetTodos() {
	todosMutex.Lock()
	defer todosMutex.Unlock()
	for k := range todos {
		delete(todos, k)
	}
}

func TestTodoRepository_CreateAndGet(t *testing.T) {
	resetTodos()
	repo := NewTodoRepository()

	created, err := repo.CreateTodo(&models.CreateTodoRequest{Title: "Test"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if created.Title != "Test" {
		t.Fatalf("expected title 'Test', got %q", created.Title)
	}

	list, err := repo.GetTodos()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("expected 1 todo, got %d", len(list))
	}
}

func TestTodoRepository_CloseTodo(t *testing.T) {
	resetTodos()
	repo := NewTodoRepository()

	created, err := repo.CreateTodo(&models.CreateTodoRequest{Title: "Close Me"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	closed, err := repo.CloseTodo(created.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !closed.Done {
		t.Fatal("expected todo to be marked done")
	}

	if _, err := repo.CloseTodo(99999); err == nil {
		t.Fatal("expected error for missing todo")
	}
}
