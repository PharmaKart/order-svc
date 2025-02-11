package handlers

import (
	"context"
	"errors"

	"github.com/PharmaKart/order-svc/internal/models"
	"github.com/PharmaKart/order-svc/internal/proto"
	"github.com/PharmaKart/order-svc/internal/repositories"
	"github.com/PharmaKart/order-svc/internal/services"
	"github.com/google/uuid"
)

type OrderHandler interface {
	PlaceOrder(ctx context.Context, req *proto.PlaceOrderRequest) (*proto.PlaceOrderResponse, error)
	GetOrder(ctx context.Context, req *proto.GetOrderRequest) (*proto.GetOrderResponse, error)
	ListCustomersOrders(ctx context.Context, req *proto.ListCustomersOrdersRequest) (*proto.ListCustomersOrdersResponse, error)
	ListAllOrders(ctx context.Context, req *proto.ListAllOrdersRequest) (*proto.ListAllOrdersResponse, error)
	UpdateOrderStatus(ctx context.Context, req *proto.UpdateOrderStatusRequest) (*proto.UpdateOrderStatusResponse, error)
}

type orderHandler struct {
	proto.UnimplementedOrderServiceServer
	orderService services.OrderService
}

func NewOrderHandler(orderRepo repositories.OrderRepository, orderItemRepo repositories.OrderItemRepository, productClient *proto.ProductServiceClient, paymentClient *proto.PaymentServiceClient) *orderHandler {
	return &orderHandler{
		orderService: services.NewOrderService(orderRepo, orderItemRepo, productClient, paymentClient),
	}
}

func (h *orderHandler) PlaceOrder(ctx context.Context, req *proto.PlaceOrderRequest) (*proto.PlaceOrderResponse, error) {
	customerId, err := uuid.Parse(req.CustomerId)
	if err != nil {
		return nil, err
	}
	order := &models.Order{
		CustomerID:      customerId,
		Status:          "payment_pending",
		PrescriptionURL: req.PrescriptionUrl,
	}
	orderItems := make([]models.OrderItem, len(req.Items))

	for i, item := range req.Items {
		productId, err := uuid.Parse(item.ProductId)
		if err != nil {
			return nil, err
		}
		orderItems[i] = models.OrderItem{
			ProductID:   productId,
			ProductName: item.ProductName,
			Quantity:    int(item.Quantity),
		}
	}

	orderId, paymentUrl, err := h.orderService.CreateOrder(*order, orderItems)
	if err != nil {
		return nil, err
	}

	return &proto.PlaceOrderResponse{
		OrderId:    orderId,
		PaymentUrl: paymentUrl,
	}, nil
}

func (h *orderHandler) GetOrder(ctx context.Context, req *proto.GetOrderRequest) (*proto.GetOrderResponse, error) {
	order, orderItems, err := h.orderService.GetOrderByID(req.OrderId)
	if err != nil {
		return nil, err
	}

	customerId := req.CustomerId

	if customerId != "admin" && order.CustomerID.String() != customerId {
		return nil, errors.New("Access denied")
	}

	protoOrderItems := make([]*proto.OrderItem, len(*orderItems))
	for i, item := range *orderItems {
		protoOrderItems[i] = &proto.OrderItem{
			ProductId:   item.ProductID.String(),
			ProductName: item.ProductName,
			Quantity:    int32(item.Quantity),
			Price:       item.Price,
		}
	}

	return &proto.GetOrderResponse{
		OrderId:         order.ID.String(),
		CustomerId:      order.CustomerID.String(),
		Status:          order.Status,
		PrescriptionUrl: order.PrescriptionURL,
		Items:           protoOrderItems,
	}, nil
}

func (h *orderHandler) ListCustomersOrders(ctx context.Context, req *proto.ListCustomersOrdersRequest) (*proto.ListCustomersOrdersResponse, error) {
	orders, total, err := h.orderService.ListCustomersOrders(req.CustomerId, req.Page, req.Limit, req.SortBy, req.SortOrder, req.Filter, req.FilterValue)
	if err != nil {
		return nil, err
	}

	protoOrders := make([]*proto.GetOrderResponse, len(*orders))
	for i, order := range *orders {
		protoOrders[i] = &proto.GetOrderResponse{
			OrderId:         order.OrderID,
			CustomerId:      order.CustomerID,
			Status:          order.Status,
			PrescriptionUrl: &order.PrescriptionURL,
		}
		protoOrderItems := make([]*proto.OrderItem, len(order.Items))
		for j, item := range order.Items {
			protoOrderItems[j] = &proto.OrderItem{
				ProductId:   item.ProductID.String(),
				ProductName: item.ProductName,
				Quantity:    int32(item.Quantity),
				Price:       float64(item.Price),
			}
		}
		protoOrders[i].Items = protoOrderItems
	}

	return &proto.ListCustomersOrdersResponse{
		Orders: protoOrders,
		Total:  total,
		Page:   req.Page,
		Limit:  req.Limit,
	}, nil
}

func (h *orderHandler) ListAllOrders(ctx context.Context, req *proto.ListAllOrdersRequest) (*proto.ListAllOrdersResponse, error) {
	orders, total, err := h.orderService.ListAllOrders(req.Page, req.Limit, req.SortBy, req.SortOrder, req.Filter, req.FilterValue)
	if err != nil {
		return nil, err
	}

	protoOrders := make([]*proto.GetOrderResponse, len(*orders))
	for i, order := range *orders {
		protoOrders[i] = &proto.GetOrderResponse{
			OrderId:         order.OrderID,
			CustomerId:      order.CustomerID,
			Status:          order.Status,
			PrescriptionUrl: &order.PrescriptionURL,
		}
		protoOrderItems := make([]*proto.OrderItem, len(order.Items))
		for j, item := range order.Items {
			protoOrderItems[j] = &proto.OrderItem{
				ProductId:   item.ProductID.String(),
				ProductName: item.ProductName,
				Quantity:    int32(item.Quantity),
				Price:       float64(item.Price),
			}
		}
		protoOrders[i].Items = protoOrderItems
	}

	return &proto.ListAllOrdersResponse{
		Orders: protoOrders,
		Total:  total,
		Page:   req.Page,
		Limit:  req.Limit,
	}, nil
}

func (h *orderHandler) UpdateOrderStatus(ctx context.Context, req *proto.UpdateOrderStatusRequest) (*proto.UpdateOrderStatusResponse, error) {
	err := h.orderService.UpdateOrderStatus(req.OrderId, req.Status)
	if err != nil {
		return nil, err
	}

	return &proto.UpdateOrderStatusResponse{}, nil
}
