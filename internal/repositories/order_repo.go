package repositories

import (
	"fmt"
	"strings"

	"github.com/PharmaKart/order-svc/internal/models"
	"github.com/PharmaKart/order-svc/pkg/errors"
	"github.com/PharmaKart/order-svc/pkg/utils"
	"gorm.io/gorm"
)

type OrderRepository interface {
	CreateOrder(order *models.Order) (string, error)
	GetOrderByID(orderID string) (*models.Order, *[]models.OrderItem, error)
	ListCustomersOrders(customerID string, filter models.Filter, sortBy string, sortOrder string, page, limit int32) ([]models.Order, int32, error)
	ListAllOrders(filter models.Filter, sortBy string, sortOrder string, page, limit int32) ([]models.Order, int32, error)
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

func (r *orderRepository) ListCustomersOrders(customerID string, filter models.Filter, sortBy string, sortOrder string, page, limit int32) ([]models.Order, int32, error) {
	var orders []models.Order
	var total int64

	allowedColumns := utils.GetModelColumns(&models.Order{})

	allowedOperators := map[string]string{
		"eq":      "=",           // Equal to
		"neq":     "!=",          // Not equal to
		"gt":      ">",           // Greater than
		"gte":     ">=",          // Greater than or equal to
		"lt":      "<",           // Less than
		"lte":     "<=",          // Less than or equal to
		"like":    "LIKE",        // LIKE for pattern matching
		"ilike":   "ILIKE",       // Case insensitive LIKE (for PostgreSQL)
		"in":      "IN",          // IN for multiple values
		"null":    "IS NULL",     // IS NULL check
		"notnull": "IS NOT NULL", // IS NOT NULL check
	}

	query := r.db.Model(&models.Order{}).Where("customer_id = ?", customerID)

	if filter != (models.Filter{}) {
		if _, allowed := allowedColumns[filter.Column]; !allowed {
			return nil, 0, errors.NewBadRequestError("invalid filter column: " + filter.Column)
		}

		op, allowed := allowedOperators[filter.Operator]
		if !allowed {
			return nil, 0, errors.NewBadRequestError("invalid filter operator: " + filter.Operator)
		}

		switch filter.Operator {
		case "like", "ilike":
			query = query.Where(filter.Column+" "+op+" ?", "%"+filter.Value+"%")
		case "in":
			values := strings.Split(filter.Value, ",")
			query = query.Where(filter.Column+" "+op+" (?)", values)
		case "null", "notnull":
			query = query.Where(filter.Column + " " + op)
		default:
			query = query.Where(filter.Column+" "+op+" ?", filter.Value)
		}
	}

	if sortBy != "" {
		if _, allowed := allowedColumns[sortBy]; !allowed {
			return nil, 0, errors.NewBadRequestError("invalid sort column: " + sortBy)
		}

		sortOrder = strings.ToLower(sortOrder)
		if sortOrder != "asc" && sortOrder != "desc" {
			sortOrder = "asc"
		}

		query = query.Order(sortBy + " " + sortOrder)
	}

	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, errors.NewInternalError(err)
	}

	if limit > 0 {
		offset := max(int((page-1)*limit), 0)
		query = query.Offset(offset).Limit(int(limit))
	}

	err = query.Find(&orders).Error
	if err != nil {
		return nil, 0, errors.NewInternalError(err)
	}

	return orders, int32(total), nil
}

func (r *orderRepository) ListAllOrders(filter models.Filter, sortBy string, sortOrder string, page, limit int32) ([]models.Order, int32, error) {
	var orders []models.Order
	var total int64

	allowedColumns := utils.GetModelColumns(&models.Order{})

	allowedOperators := map[string]string{
		"eq":      "=",           // Equal to
		"neq":     "!=",          // Not equal to
		"gt":      ">",           // Greater than
		"gte":     ">=",          // Greater than or equal to
		"lt":      "<",           // Less than
		"lte":     "<=",          // Less than or equal to
		"like":    "LIKE",        // LIKE for pattern matching
		"ilike":   "ILIKE",       // Case insensitive LIKE (for PostgreSQL)
		"in":      "IN",          // IN for multiple values
		"null":    "IS NULL",     // IS NULL check
		"notnull": "IS NOT NULL", // IS NOT NULL check
	}

	query := r.db.Model(&models.Order{})

	if filter != (models.Filter{}) {
		if _, allowed := allowedColumns[filter.Column]; !allowed {
			return nil, 0, errors.NewBadRequestError("invalid filter column: " + filter.Column)
		}

		op, allowed := allowedOperators[filter.Operator]
		if !allowed {
			return nil, 0, errors.NewBadRequestError("invalid filter operator: " + filter.Operator)
		}

		switch filter.Operator {
		case "like", "ilike":
			query = query.Where(filter.Column+" "+op+" ?", "%"+filter.Value+"%")
		case "in":
			values := strings.Split(filter.Value, ",")
			query = query.Where(filter.Column+" "+op+" (?)", values)
		case "null", "notnull":
			query = query.Where(filter.Column + " " + op)
		default:
			query = query.Where(filter.Column+" "+op+" ?", filter.Value)
		}
	}

	if sortBy != "" {
		if _, allowed := allowedColumns[sortBy]; !allowed {
			return nil, 0, errors.NewBadRequestError("invalid sort column: " + sortBy)
		}

		sortOrder = strings.ToLower(sortOrder)
		if sortOrder != "asc" && sortOrder != "desc" {
			sortOrder = "asc"
		}

		query = query.Order(sortBy + " " + sortOrder)
	}

	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, errors.NewInternalError(err)
	}

	if limit > 0 {
		offset := max(int((page-1)*limit), 0)
		query = query.Offset(offset).Limit(int(limit))
	}

	err = query.Find(&orders).Error
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
