package repositories

import (
	"fmt"

	"github.com/PharmaKart/order-svc/internal/models"
	"github.com/PharmaKart/order-svc/pkg/errors"
	"gorm.io/gorm"
)

type OrderRepository interface {
	CreateOrder(order *models.Order) (string, error)
	GetOrderByID(orderID string) (*models.Order, *[]models.OrderItem, error)
	ListCustomersOrders(customerID string, page int32, limit int32, sortBy string, sortOrder string, filter string, filterValue string) ([]models.Order, int32, error)
	ListAllOrders(page int32, limit int32, sortBy string, sortOrder string, filter string, filterValue string) ([]models.Order, int32, error)
	UpdateOrderStatus(orderID string, status string) error
}

type orderRepository struct {
	db *gorm.DB
}

func NewOrderRepository(db *gorm.DB) OrderRepository {
	return &orderRepository{db}
}

func (r *orderRepository) CreateOrder(order *models.Order) (string, error) {
	if err := r.db.Create(order).Error; err != nil {
		return "", errors.NewInternalError(err)
	}

	return order.ID.String(), nil
}

func (r *orderRepository) GetOrderByID(orderID string) (*models.Order, *[]models.OrderItem, error) {
	var order models.Order
	var items []models.OrderItem

	err := r.db.Where("id = ?", orderID).First(&order).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil, errors.NewNotFoundError(fmt.Sprintf("Order with ID '%s' not found", orderID))
		}
		return nil, nil, errors.NewInternalError(err)
	}

	err = r.db.Where("order_id = ?", orderID).Find(&items).Error
	if err != nil {
		return nil, nil, errors.NewInternalError(err)
	}

	return &order, &items, nil
}

func (r *orderRepository) ListCustomersOrders(customerID string, page int32, limit int32, sortBy string, sortOrder string, filter string, filterValue string) ([]models.Order, int32, error) {
	var orders []models.Order
	var total int64

	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 10
	}

	query := r.db.Where("customer_id = ?", customerID)
	if filter != "" && filterValue != "" {
		query = query.Where(filter+" = ?", filterValue)
	}

	if sortBy != "" {
		if sortOrder == "" {
			sortOrder = "asc"
		}
		query = query.Order(sortBy + " " + sortOrder)
	}

	err := query.Offset(int((page - 1) * limit)).Limit(int(limit)).Find(&orders).Error
	if err != nil {
		return nil, 0, errors.NewInternalError(err)
	}

	err = query.Model(&models.Order{}).Count(&total).Error
	if err != nil {
		return nil, 0, errors.NewInternalError(err)
	}

	return orders, int32(total), nil
}

func (r *orderRepository) ListAllOrders(page int32, limit int32, sortBy string, sortOrder string, filter string, filterValue string) ([]models.Order, int32, error) {
	var orders []models.Order
	var total int64

	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 10
	}

	query := r.db
	if filter != "" && filterValue != "" {
		query = query.Where(filter+" = ?", filterValue)
	}

	if sortBy != "" {
		if sortOrder == "" {
			sortOrder = "asc"
		}
		query = query.Order(sortBy + " " + sortOrder)
	}

	err := query.Offset(int((page - 1) * limit)).Limit(int(limit)).Find(&orders).Error
	if err != nil {
		return nil, 0, errors.NewInternalError(err)
	}

	err = query.Model(&models.Order{}).Count(&total).Error
	if err != nil {
		return nil, 0, errors.NewInternalError(err)
	}

	return orders, int32(total), nil
}

func (r *orderRepository) UpdateOrderStatus(orderID string, status string) error {
	result := r.db.Model(&models.Order{}).Where("id = ?", orderID).Update("status", status)

	if result.Error != nil {
		return errors.NewInternalError(result.Error)
	}

	if result.RowsAffected == 0 {
		return errors.NewNotFoundError(fmt.Sprintf("Order with ID '%s' not found", orderID))
	}

	return nil
}
