package services

import (
	demoerrors "demo/custom-errors"
	"demo/models"
	"demo/repositories"
	"time"
)

func NewTodoService(repo repositories.TodoRepository) TodoService {
	return &todoServiceImpl{
		repo: repo,
	}
}

type TodoService interface {
	GetTodos() ([]models.Todo, error)
	CreateTodo(todo *models.CreateTodoRequest) (models.Todo, error)
	CloseTodo(id uint32) (models.Todo, error)
}

type todoServiceImpl struct {
	repo repositories.TodoRepository
}

func (s *todoServiceImpl) GetTodos() ([]models.Todo, error) {
	return s.repo.GetTodos()
}

func (s *todoServiceImpl) CreateTodo(todo *models.CreateTodoRequest) (models.Todo, error) {
	var zero models.Todo

	if todo == nil {
		return zero, demoerrors.NewValidationError("todo is empty")
	}
	if todo.Title == "" {
		return zero, demoerrors.NewValidationError("todo title is empty")
	}
	if len(todo.Title) > 255 {
		return zero, demoerrors.NewValidationError("todo title is too long, max length is 255 characters")
	}
	if !todo.DueDate.IsZero() && todo.DueDate.Before(time.Now()) {
		return zero, demoerrors.NewValidationError("todo due date is in the past")
	}

	return s.repo.CreateTodo(todo)
}

func (s *todoServiceImpl) CloseTodo(id uint32) (models.Todo, error) {
	if id == 0 {
		return models.Todo{}, demoerrors.NewValidationError("invalid todo id")
	}

	return s.repo.CloseTodo(id)
}
