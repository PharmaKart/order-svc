package repositories

import (
	"fmt"

	"github.com/PharmaKart/order-svc/internal/models"
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
	stmt := r.db.Session(&gorm.Session{DryRun: true}).Create(order).Statement
	sql := stmt.SQL.String()
	fmt.Printf("Generated SQL: %s\n", sql)
	fmt.Printf("Variables: %v\n", stmt.Vars)
	// return r.db.Create(order).Error
	if err := r.db.Create(order).Error; err != nil {
		return "", err
	}

	return order.ID.String(), nil
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

func (r *orderRepository) ListCustomersOrders(customerID string, page int32, limit int32, sortBy string, sortOrder string, filter string, filterValue string) ([]models.Order, int32, error) {
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
		return nil, 0, err
	}

	err = query.Model(&models.Order{}).Where("customer_id = ?", customerID).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}
	return orders, int32(total), err
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
		return nil, 0, err
	}

	err = query.Model(&models.Order{}).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	return orders, int32(total), err
}

func (r *orderRepository) UpdateOrderStatus(orderID string, status string) error {
	return r.db.Model(&models.Order{}).Where("id = ?", orderID).Update("status", status).Error
}
