package services

import (
	"context"
	"fmt"
	"time"

	"github.com/PharmaKart/order-svc/internal/models"
	"github.com/PharmaKart/order-svc/internal/proto"
	"github.com/PharmaKart/order-svc/internal/repositories"
	"github.com/PharmaKart/order-svc/pkg/errors"
	"github.com/google/uuid"
)

type OrderService interface {
	CreateOrder(order models.Order, orderItems []models.OrderItem) (string, string, error)
	GetOrderByID(orderID string) (*models.Order, *[]models.OrderItem, error)
	ListCustomersOrders(customerID string, filter models.Filter, sortBy string, sortOrder string, page, limit int32) (*[]OrderResponse, int32, error)
	ListAllOrders(filter models.Filter, sortBy string, sortOrder string, page, limit int32) (*[]OrderResponse, int32, error)
	UpdateOrderStatus(orderID, customerID, status string) error
	GenerateNewPaymentUrl(orderID, customerID string) (string, error)
}

type orderService struct {
	orderRepo     repositories.OrderRepository
	orderItemRepo repositories.OrderItemRepository
	productClient proto.ProductServiceClient
	paymentClient proto.PaymentServiceClient
}

type OrderResponse struct {
	OrderID         string
	CustomerID      string
	Status          string
	PrescriptionURL *string
	ShippingCost    float64
	Subtotal        float64
	CreatedAt       time.Time
	UpdatedAt       time.Time
	Items           []models.OrderItem
}

func NewOrderService(orderRepo repositories.OrderRepository, orderItemRepo repositories.OrderItemRepository, productClient *proto.ProductServiceClient, paymentClient *proto.PaymentServiceClient) OrderService {
	return &orderService{
		orderRepo:     orderRepo,
		orderItemRepo: orderItemRepo,
		productClient: *productClient,
		paymentClient: *paymentClient,
	}
}

func (s *orderService) CreateOrder(order models.Order, orderItems []models.OrderItem) (string, string, error) {
	// Check Product Service for product stock
	ctx := context.Background()
	orderItemsList := []models.OrderItem{}
	subtotal := 0.0
	for _, item := range orderItems {
		if err := uuid.Validate("uuid..."); err == nil {
			return "", "", errors.NewValidationError("product_id", "Invalid product ID")
		}

		if item.Quantity <= 0 {
			return "", "", errors.NewValidationError("quantity", "Quantity must be greater than 0")
		}

		if item.ProductName == "" {
			return "", "", errors.NewValidationError("product_name", "Product name is required")
		}

		product, err := s.productClient.GetProduct(ctx, &proto.GetProductRequest{ProductId: item.ProductID.String()})
		if err != nil {
			return "", "", err
		}
		if int(product.Product.Stock) < item.Quantity {
			return "", "", errors.NewValidationError("stock", fmt.Sprintf("Not enough stock for product %s", item.ProductName))
		}

		// Check if product is prescription based
		if product.Product.RequiresPrescription && order.PrescriptionURL == nil {
			return "", "", errors.NewValidationError("prescription", fmt.Sprintf("Prescription required for product %s", item.ProductName))
		}

		// Deduct stock from product
		_, err = s.productClient.UpdateStock(ctx, &proto.UpdateStockRequest{
			ProductId:      item.ProductID.String(),
			QuantityChange: int32(item.Quantity) * -1,
			Reason:         "order_placed",
		})
		if err != nil {
			return "", "", err
		}
		item.Price = product.Product.Price
		orderItemsList = append(orderItemsList, item)
		subtotal += item.Price * float64(item.Quantity)
	}

	order.Subtotal = subtotal

	if subtotal > 40.00 {
		order.ShippingCost = 0.00
	} else {
		order.ShippingCost = 10.00
	}

	order_id, err := s.orderRepo.CreateOrder(&order)
	if err != nil {
		return "", "", err
	}

	for _, item := range orderItemsList {

		item.OrderID = order.ID

		s.orderItemRepo.AddOrderItem(&item)
	}

	paymentURL, err := s.paymentClient.GeneratePaymentURL(ctx, &proto.GeneratePaymentURLRequest{
		OrderId:    order_id,
		CustomerId: order.CustomerID.String(),
	})
	if err != nil {
		return "", "", err
	}

	return order_id, paymentURL.Url, nil
}

func (s *orderService) GenerateNewPaymentUrl(orderID, customerID string) (string, error) {
	ctx := context.Background()

	// First, get the order to check its status
	order, _, err := s.orderRepo.GetOrderByID(orderID)
	if err != nil {
		return "", err
	}

	// Check if order status is payment_pending
	if order.Status != "payment_pending" {
		return "", errors.NewConflictError("Order already paid for")
	}

	// Proceed with generating payment URL
	paymentURL, err := s.paymentClient.GeneratePaymentURL(ctx, &proto.GeneratePaymentURLRequest{
		OrderId:    orderID,
		CustomerId: customerID,
	})
	if err != nil {
		return "", err
	}

	if !paymentURL.Success {
		return "", &errors.AppError{
			Type:    errors.InternalError,
			Message: paymentURL.Error.Message,
		}
	}

	return paymentURL.Url, nil
}

func (s *orderService) GetOrderByID(orderID string) (*models.Order, *[]models.OrderItem, error) {
	order, items, err := s.orderRepo.GetOrderByID(orderID)
	if err != nil {
		return nil, nil, err
	}

	return order, items, nil
}

func (s *orderService) ListCustomersOrders(customerID string, filter models.Filter, sortBy string, sortOrder string, page, limit int32) (*[]OrderResponse, int32, error) {
	ordersResponse := []OrderResponse{}

	orders, total, err := s.orderRepo.ListCustomersOrders(customerID, filter, sortBy, sortOrder, page, limit)
	if err != nil {
		return nil, 0, err
	}

	for _, order := range orders {
		items, err := s.orderItemRepo.GetItemsByOrderID(order.ID.String())
		if err != nil {
			return nil, 0, err
		}
		ordersResponse = append(ordersResponse, OrderResponse{
			OrderID:         order.ID.String(),
			CustomerID:      order.CustomerID.String(),
			Status:          order.Status,
			PrescriptionURL: order.PrescriptionURL,
			ShippingCost:    order.ShippingCost,
			Subtotal:        order.Subtotal,
			CreatedAt:       order.CreatedAt,
			UpdatedAt:       order.UpdatedAt,
			Items:           items,
		})
	}

	return &ordersResponse, total, nil
}

func (s *orderService) ListAllOrders(filter models.Filter, sortBy string, sortOrder string, page, limit int32) (*[]OrderResponse, int32, error) {
	ordersResponse := []OrderResponse{}

	orders, total, err := s.orderRepo.ListAllOrders(filter, sortBy, sortOrder, page, limit)
	if err != nil {
		return nil, 0, err
	}

	for _, order := range orders {
		items, err := s.orderItemRepo.GetItemsByOrderID(order.ID.String())
		if err != nil {
			return nil, 0, err
		}
		ordersResponse = append(ordersResponse, OrderResponse{
			OrderID:         order.ID.String(),
			CustomerID:      order.CustomerID.String(),
			Status:          order.Status,
			PrescriptionURL: order.PrescriptionURL,
			ShippingCost:    order.ShippingCost,
			Subtotal:        order.Subtotal,
			CreatedAt:       order.CreatedAt,
			UpdatedAt:       order.UpdatedAt,
			Items:           items,
		})
	}

	return &ordersResponse, total, nil
}

func (s *orderService) UpdateOrderStatus(orderID, customerID, status string) error {
	order, _, err := s.orderRepo.GetOrderByID(orderID)
	if err != nil {
		return err
	}

	if order.Status == "cancelled" {
		return errors.NewConflictError("Order already cancelled")
	}
	if order.Status == "completed" {
		return errors.NewConflictError("Order already completed")
	}

	switch {
	case customerID == "admin":
		return s.orderRepo.UpdateOrderStatus(orderID, status)

	case customerID == "payment_service" && status == "paid":
		return s.orderRepo.UpdateOrderStatus(orderID, status)

	case customerID == order.CustomerID.String() && status == "cancelled" && order.Status != "shipping":
		return s.orderRepo.UpdateOrderStatus(orderID, status)

	default:
		return errors.NewAuthError("Access denied")
	}
}
