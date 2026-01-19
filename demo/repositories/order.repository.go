package repositories

import (
	"go-di/demo/models"
	"sync"
	"time"
)

var ordersMutex = sync.RWMutex{}
var orders = []models.Order{
	{ID: 1, UserID: 1, ProductName: "Product 1", Amount: 100.0, CreatedAt: time.Now()},
	{ID: 2, UserID: 2, ProductName: "Product 2", Amount: 200.0, CreatedAt: time.Now()},
}

func nextOrderID() int {
	maxID := 0
	for _, order := range orders {
		if order.ID > maxID {
			maxID = order.ID
		}
	}
	return maxID + 1
}

func NewOrderRepository() OrderRepository {
	return &OrderRepositoryImpl{}
}

type OrderRepository interface {
	GetOrders(pagination *models.Pagination) ([]models.Order, error)
	GetOrdersByUser(userID int, pagination *models.Pagination) ([]models.Order, error)
	CreateOrder(order *models.Order) (*models.Order, error)
}

type OrderRepositoryImpl struct{}

func (r *OrderRepositoryImpl) GetOrders(pagination *models.Pagination) ([]models.Order, error) {
	// For demonstration purposes, return a list of dummy orders
	index := (pagination.Page - 1) * pagination.PageSize
	end := index + pagination.PageSize
	if index > len(orders) {
		return []models.Order{}, nil
	}
	if end > len(orders) {
		end = len(orders)
	}
	return orders[index:end], nil
}

func (r *OrderRepositoryImpl) CreateOrder(order *models.Order) (*models.Order, error) {
	ordersMutex.Lock()
	defer ordersMutex.Unlock()

	order.ID = nextOrderID()
	order.CreatedAt = time.Now()

	orders = append(orders, *order)
	return order, nil
}

func (r *OrderRepositoryImpl) GetOrdersByUser(userID int, pagination *models.Pagination) ([]models.Order, error) {
	// For demonstration purposes, return a list of dummy orders for the given user
	ordersMutex.RLock()
	defer ordersMutex.RUnlock()

	var userOrders []models.Order
	for _, order := range orders {
		if order.UserID == userID {
			userOrders = append(userOrders, order)
		}
	}
	return userOrders, nil
}
