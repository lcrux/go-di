package services

import (
	"go-di/demo/models"
	"go-di/demo/repositories"
)

func NewOrderService(orderRepository repositories.OrderRepository) OrderService {
	return &OrderServiceImpl{
		orderRepository: orderRepository,
	}
}

type OrderService interface {
	GetOrders(pagination *models.Pagination) ([]models.Order, error)
	GetOrdersByUser(userID int, pagination *models.Pagination) ([]models.Order, error)
	CreateOrder(order *models.Order) (*models.Order, error)
}

type OrderServiceImpl struct {
	orderRepository repositories.OrderRepository
}

// GetOrders implements OrderService.
func (s *OrderServiceImpl) GetOrders(pagination *models.Pagination) ([]models.Order, error) {
	return s.orderRepository.GetOrders(pagination)
}

func (s *OrderServiceImpl) CreateOrder(order *models.Order) (*models.Order, error) {
	createdOrder, err := s.orderRepository.CreateOrder(order)
	if err != nil {
		return nil, err
	}
	return createdOrder, nil
}

func (s *OrderServiceImpl) GetOrdersByUser(userID int, pagination *models.Pagination) ([]models.Order, error) {
	return s.orderRepository.GetOrdersByUser(userID, pagination)
}
