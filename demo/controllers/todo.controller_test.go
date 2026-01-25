package controllers

import (
	"bytes"
	customErrors "demo/custom-errors"
	"demo/models"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

type fakeTodoService struct {
	Todos  []models.Todo
	Err    error
	Last   *models.CreateTodoRequest
	Closed []uint32
}

func (f *fakeTodoService) GetTodos() ([]models.Todo, error) {
	return f.Todos, f.Err
}

func (f *fakeTodoService) CreateTodo(todo *models.CreateTodoRequest) (models.Todo, error) {
	f.Last = todo
	if f.Err != nil {
		return models.Todo{}, f.Err
	}
	return models.Todo{ID: 1, Title: todo.Title}, nil
}

func (f *fakeTodoService) CloseTodo(id uint32) (models.Todo, error) {
	f.Closed = append(f.Closed, id)
	if f.Err != nil {
		return models.Todo{}, f.Err
	}
	return models.Todo{ID: id, Done: true}, nil
}

func TestTodoController_GetTodos(t *testing.T) {
	svc := &fakeTodoService{Todos: []models.Todo{{ID: 1, Title: "A"}}}
	controller := NewTodoController(svc)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/todos", nil)
	controller.GetTodos(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	var response []models.Todo
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(response) != 1 || response[0].Title != "A" {
		t.Fatalf("unexpected response: %+v", response)
	}
}

func TestTodoController_GetTodos_NotFound(t *testing.T) {
	svc := &fakeTodoService{Err: customErrors.NewNotFoundError("missing")}
	controller := NewTodoController(svc)

	rec := httptest.NewRecorder()
	controller.GetTodos(rec, httptest.NewRequest(http.MethodGet, "/todos", nil))

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", rec.Code)
	}
}

func TestTodoController_CreateTodo(t *testing.T) {
	svc := &fakeTodoService{}
	controller := NewTodoController(svc)

	body, _ := json.Marshal(models.CreateTodoRequest{Title: "New"})
	rec := httptest.NewRecorder()
	controller.CreateTodo(rec, httptest.NewRequest(http.MethodPost, "/todos", bytes.NewReader(body)))

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d", rec.Code)
	}
}

func TestTodoController_CreateTodo_BadRequest(t *testing.T) {
	svc := &fakeTodoService{}
	controller := NewTodoController(svc)

	rec := httptest.NewRecorder()
	controller.CreateTodo(rec, httptest.NewRequest(http.MethodPost, "/todos", bytes.NewBufferString("not-json")))

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rec.Code)
	}
}

func TestTodoController_CloseTodo(t *testing.T) {
	svc := &fakeTodoService{}
	controller := NewTodoController(svc)

	body, _ := json.Marshal(models.CloseTodoRequest{ID: 5})
	req := httptest.NewRequest(http.MethodPatch, "/todos/5/done", bytes.NewReader(body))
	req.SetPathValue("id", "5")

	rec := httptest.NewRecorder()
	controller.CloseTodo(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
}

func TestTodoController_CloseTodo_IdMismatch(t *testing.T) {
	svc := &fakeTodoService{}
	controller := NewTodoController(svc)

	body, _ := json.Marshal(models.CloseTodoRequest{ID: 6})
	req := httptest.NewRequest(http.MethodPatch, "/todos/5/done", bytes.NewReader(body))
	req.SetPathValue("id", "5")

	rec := httptest.NewRecorder()
	controller.CloseTodo(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rec.Code)
	}
}

func TestTodoController_CloseTodo_InvalidId(t *testing.T) {
	svc := &fakeTodoService{}
	controller := NewTodoController(svc)

	body, _ := json.Marshal(models.CloseTodoRequest{ID: 1})
	req := httptest.NewRequest(http.MethodPatch, "/todos/bad/done", bytes.NewReader(body))
	req.SetPathValue("id", "bad")

	rec := httptest.NewRecorder()
	controller.CloseTodo(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rec.Code)
	}
}
