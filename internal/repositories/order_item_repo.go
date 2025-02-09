package repositories

import (
	"github.com/PharmaKart/order-svc/internal/models"
	"gorm.io/gorm"
)

type OrderItemRepository interface {
	AddOrderItem(item *models.OrderItem) error
	GetItemsByOrderID(orderID string) ([]models.OrderItem, error)
}

type orderItemRepository struct {
	db *gorm.DB
}

func NewOrderItemRepository(db *gorm.DB) OrderItemRepository {
	return &orderItemRepository{db}
}

func (r *orderItemRepository) AddOrderItem(item *models.OrderItem) error {
	return r.db.Create(item).Error
}

func (r *orderItemRepository) GetItemsByOrderID(orderID string) ([]models.OrderItem, error) {
	var items []models.OrderItem
	err := r.db.Where("order_id = ?", orderID).Find(&items).Error
	return items, err
}
