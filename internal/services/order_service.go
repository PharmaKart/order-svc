package services

import (
	"context"
	"errors"

	"github.com/PharmaKart/order-svc/internal/models"
	"github.com/PharmaKart/order-svc/internal/proto"
	"github.com/PharmaKart/order-svc/internal/repositories"
)

type OrderService interface {
	CreateOrder(order models.Order, orderItems []models.OrderItem) (string, string, error)
	GetOrderByID(orderID string) (*models.Order, *[]models.OrderItem, error)
	ListCustomersOrders(customerID string, page int32, limit int32, sortBy string, sortOrder string, filter string, filterValue string) (*[]OrderResponse, int32, error)
	ListAllOrders(page int32, limit int32, sortBy string, sortOrder string, filter string, filterValue string) (*[]OrderResponse, int32, error)
	UpdateOrderStatus(orderID string, status string) error
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
	PrescriptionURL string
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
	for _, item := range orderItems {
		product, err := s.productClient.GetProduct(ctx, &proto.GetProductRequest{ProductId: item.ProductID.String()})
		if err != nil {
			return "", "", err
		}
		if int(product.Product.Stock) < item.Quantity {
			return "", "", errors.New("Product out of stock")
		}
		// Deduct stock from product
		_, err = s.productClient.UpdateStock(ctx, &proto.UpdateStockRequest{
			ProductId: item.ProductID.String(),
			Quantity:  int32(item.Quantity) * -1,
			Reason:    "order_placed",
		})
		if err != nil {
			return "", "", err
		}
		item.Price = product.Product.Price
		item.OrderID = order.ID

		s.orderItemRepo.AddOrderItem(&item)
	}

	order_id, err := s.orderRepo.CreateOrder(&order)
	if err != nil {
		return "", "", err
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

func (s *orderService) GetOrderByID(orderID string) (*models.Order, *[]models.OrderItem, error) {
	order, items, err := s.orderRepo.GetOrderByID(orderID)
	if err != nil {
		return nil, nil, err
	}

	return order, items, nil
}

func (s *orderService) ListCustomersOrders(customerID string, page int32, limit int32, sortBy string, sortOrder string, filter string, filterValue string) (*[]OrderResponse, int32, error) {
	ordersResponse := []OrderResponse{}

	orders, total, err := s.orderRepo.ListCustomersOrders(customerID, page, limit, sortBy, sortOrder, filter, filterValue)
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
			PrescriptionURL: *order.PrescriptionURL,
			Items:           items,
		})
	}

	return &ordersResponse, total, nil
}

func (s *orderService) ListAllOrders(page int32, limit int32, sortBy string, sortOrder string, filter string, filterValue string) (*[]OrderResponse, int32, error) {
	ordersResponse := []OrderResponse{}

	orders, total, err := s.orderRepo.ListAllOrders(page, limit, sortBy, sortOrder, filter, filterValue)
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
			PrescriptionURL: *order.PrescriptionURL,
			Items:           items,
		})
	}

	return &ordersResponse, total, nil
}

func (s *orderService) UpdateOrderStatus(orderID string, status string) error {
	return s.orderRepo.UpdateOrderStatus(orderID, status)
}
