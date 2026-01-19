package services

import (
	"go-di/demo/models"
	"go-di/demo/repositories"
)

type UserService interface {
	GetUser(id int) (models.User, error)
	GetUsers(pagination *models.Pagination) ([]models.User, error)
}

func NewUserService(userRepository repositories.UserRepository) UserService {
	return &UserServiceImpl{
		userRepository: userRepository,
	}
}

type UserServiceImpl struct {
	userRepository repositories.UserRepository
}

// GetUsers implements UserService.
func (s *UserServiceImpl) GetUsers(pagination *models.Pagination) ([]models.User, error) {
	return s.userRepository.GetUsers(pagination)
}

func (s *UserServiceImpl) GetUser(id int) (models.User, error) {
	return s.userRepository.GetUser(id)
}
