package repositories

import (
	"fmt"
	"go-di/demo/models"
	"sync"
)

var usersMutex = sync.RWMutex{}
var users = []models.User{
	{ID: 1, Name: "John Doe"},
	{ID: 2, Name: "Jane Smith"},
}

func nextUserID() int {
	maxID := 0
	for _, user := range users {
		if user.ID > maxID {
			maxID = user.ID
		}
	}
	return maxID + 1
}

type UserRepository interface {
	GetUsers(pagination *models.Pagination) ([]models.User, error)
	GetUser(id int) (models.User, error)
	CreatUser(user models.User) error
	DeleteUser(id int) error
}

func NewUserRepository() UserRepository {
	return &UserRepositoryImpl{}
}

type UserRepositoryImpl struct{}

func (r *UserRepositoryImpl) GetUser(id int) (models.User, error) {
	usersMutex.RLock()
	defer usersMutex.RUnlock()

	for _, user := range users {
		if user.ID == id {
			return user, nil
		}
	}
	return models.User{}, fmt.Errorf("user with ID %d not found", id)
}

func (r *UserRepositoryImpl) CreatUser(user models.User) error {
	usersMutex.Lock()
	defer usersMutex.Unlock()

	user.ID = nextUserID()
	users = append(users, user)
	return nil
}

func (r *UserRepositoryImpl) DeleteUser(id int) error {
	usersMutex.Lock()
	defer usersMutex.Unlock()

	for i, user := range users {
		if user.ID == id {
			users = append(users[:i], users[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("user with ID %d not found", id)
}

func (r *UserRepositoryImpl) GetUsers(pagination *models.Pagination) ([]models.User, error) {
	// For demonstration purposes, return a list of dummy users
	usersMutex.RLock()
	defer usersMutex.RUnlock()

	index := (pagination.Page - 1) * pagination.PageSize
	end := index + pagination.PageSize
	if index > len(users) {
		return []models.User{}, nil
	}
	if end > len(users) {
		end = len(users)
	}

	return users[index:end], nil
}
