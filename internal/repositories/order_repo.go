package repositories

import (
	"github.com/PharmaKart/order-svc/internal/models"
	"gorm.io/gorm"
)

type OrderRepository interface {
	CreateOrder(order *models.Order) error
	GetOrderByID(orderID string) (*models.Order, *[]models.OrderItem, error)
	ListCustomersOrders(customerID string) (*[]models.Order, error)
	ListAllOrders() (*[]models.Order, error)
	UpdateOrderStatus(orderID string, status string) error
}

type orderRepository struct {
	db *gorm.DB
}

func NewOrderRepository(db *gorm.DB) OrderRepository {
	return &orderRepository{db}
}

func (r *orderRepository) CreateOrder(order *models.Order) error {
	return r.db.Create(order).Error
}

func (r *orderRepository) GetOrderByID(orderID string) (*models.Order, *[]models.OrderItem, error) {
	var order models.Order
	var items []models.OrderItem
	err := r.db.Where("id = ?", orderID).First(&order).Error
	if err != nil {
		return nil, nil, err
	}

	err = r.db.Where("order_id = ?", orderID).Find(&items).Error
	if err != nil {
		return nil, nil, err
	}

	return &order, &items, nil
}

func (r *orderRepository) ListCustomersOrders(customerID string) (*[]models.Order, error) {
	var orders []models.Order
	err := r.db.Where("customer_id = ?", customerID).Find(&orders).Error
	if err != nil {
		return nil, err
	}

	return &orders, nil
}

func (r *orderRepository) ListAllOrders() (*[]models.Order, error) {
	var orders []models.Order
	err := r.db.Find(&orders).Error
	if err != nil {
		return nil, err
	}

	return &orders, nil
}

func (r *orderRepository) UpdateOrderStatus(orderID string, status string) error {
	return r.db.Model(&models.Order{}).Where("id = ?", orderID).Update("status", status).Error
}
