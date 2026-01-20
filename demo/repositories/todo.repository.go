package repositories

import (
	"demo/models"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

var todosMutex = sync.RWMutex{}
var todos map[uint32]*models.Todo = make(map[uint32]*models.Todo)

func init() {
	var todoId uint32
	todoId = uuid.New().ID()
	todos[todoId] = &models.Todo{ID: todoId, Title: "Demo Todo " + fmt.Sprint(todoId), Done: false, CreatedAt: time.Now(), UpdatedAt: time.Now()}

	todoId = uuid.New().ID()
	todos[todoId] = &models.Todo{ID: todoId, Title: "Demo Todo " + fmt.Sprint(todoId), Done: false, CreatedAt: time.Now(), UpdatedAt: time.Now()}
}

func NewTodoRepository() TodoRepository {
	return &TodoRepositoryImpl{}
}

type TodoRepository interface {
	GetTodos() ([]models.Todo, error)
	CreateTodo(todo *models.CreateTodoRequest) (models.Todo, error)
	CloseTodo(id uint32) (models.Todo, error)
}

type TodoRepositoryImpl struct{}

func (r *TodoRepositoryImpl) GetTodos() ([]models.Todo, error) {
	todosMutex.RLock()
	defer todosMutex.RUnlock()

	todoList := make([]models.Todo, 0, len(todos))
	for _, todo := range todos {
		// Append a copy of the todo to the list
		todoList = append(todoList, *todo)
	}
	return todoList, nil
}

func (r *TodoRepositoryImpl) CreateTodo(todo *models.CreateTodoRequest) (models.Todo, error) {
	todosMutex.Lock()
	defer todosMutex.Unlock()

	if todo == nil {
		return models.Todo{}, fmt.Errorf("todo is nil")
	}

	newTodo := &models.Todo{
		ID:        uuid.New().ID(),
		Title:     todo.Title,
		Done:      false,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	todos[newTodo.ID] = newTodo

	// Return a copy of the newly created todo
	return *newTodo, nil
}

func (r *TodoRepositoryImpl) CloseTodo(id uint32) (models.Todo, error) {
	todosMutex.Lock()
	defer todosMutex.Unlock()

	todo, ok := todos[id]
	if !ok {
		return models.Todo{}, fmt.Errorf("todo not found")
	}

	// Mark the todo as done
	todo.Done = true
	todo.UpdatedAt = time.Now()

	// Return a copy of the updated todo
	return *todo, nil
}
