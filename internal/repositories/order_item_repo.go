package repositories

import (
	"fmt"

	"github.com/PharmaKart/order-svc/internal/models"
	"github.com/PharmaKart/order-svc/pkg/errors"
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
	if err := r.db.Create(item).Error; err != nil {
		return errors.NewInternalError(err)
	}
	return nil
}

func (r *orderItemRepository) GetItemsByOrderID(orderID string) ([]models.OrderItem, error) {
	var items []models.OrderItem

	if err := r.db.Where("order_id = ?", orderID).Find(&items).Error; err != nil {
		return nil, errors.NewInternalError(err)
	}

	if len(items) == 0 {
		return nil, errors.NewNotFoundError(fmt.Sprintf("No items found for order ID '%s'", orderID))
	}

	return items, nil
}
