package handlers

import (
	"context"

	"github.com/PharmaKart/order-svc/internal/models"
	"github.com/PharmaKart/order-svc/internal/proto"
	"github.com/PharmaKart/order-svc/internal/repositories"
	"github.com/PharmaKart/order-svc/internal/services"
	"github.com/PharmaKart/order-svc/pkg/errors"
	"github.com/PharmaKart/order-svc/pkg/utils"
	"github.com/google/uuid"
)

type OrderHandler interface {
	PlaceOrder(ctx context.Context, req *proto.PlaceOrderRequest) (*proto.PlaceOrderResponse, error)
	GetOrder(ctx context.Context, req *proto.GetOrderRequest) (*proto.GetOrderResponse, error)
	ListCustomersOrders(ctx context.Context, req *proto.ListCustomersOrdersRequest) (*proto.ListCustomersOrdersResponse, error)
	ListAllOrders(ctx context.Context, req *proto.ListAllOrdersRequest) (*proto.ListAllOrdersResponse, error)
	UpdateOrderStatus(ctx context.Context, req *proto.UpdateOrderStatusRequest) (*proto.UpdateOrderStatusResponse, error)
	GenerateNewPaymentUrl(ctx context.Context, req *proto.GenerateNewPaymentUrlRequest) (*proto.GenerateNewPaymentUrlResponse, error)
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
		if appErr, ok := errors.IsAppError(err); ok {
			return &proto.PlaceOrderResponse{
				Success: false,
				Error: &proto.Error{
					Type:    string(appErr.Type),
					Message: appErr.Message,
					Details: utils.ConvertMapToKeyValuePairs(appErr.Details),
				},
			}, nil
		}

		return &proto.PlaceOrderResponse{
			Success: false,
			Error: &proto.Error{
				Type:    string(errors.InternalError),
				Message: "An unexpected error occurred",
			},
		}, nil
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
			if appErr, ok := errors.IsAppError(err); ok {
				return &proto.PlaceOrderResponse{
					Success: false,
					Error: &proto.Error{
						Type:    string(appErr.Type),
						Message: appErr.Message,
						Details: utils.ConvertMapToKeyValuePairs(appErr.Details),
					},
				}, nil
			}

			return &proto.PlaceOrderResponse{
				Success: false,
				Error: &proto.Error{
					Type:    string(errors.InternalError),
					Message: "An unexpected error occurred",
				},
			}, nil
		}
		orderItems[i] = models.OrderItem{
			ProductID:   productId,
			ProductName: item.ProductName,
			Quantity:    int(item.Quantity),
		}
	}

	orderId, paymentUrl, err := h.orderService.CreateOrder(*order, orderItems)
	if err != nil {
		if appErr, ok := errors.IsAppError(err); ok {
			return &proto.PlaceOrderResponse{
				Success: false,
				Error: &proto.Error{
					Type:    string(appErr.Type),
					Message: appErr.Message,
					Details: utils.ConvertMapToKeyValuePairs(appErr.Details),
				},
			}, nil
		}

		return &proto.PlaceOrderResponse{
			Success: false,
			Error: &proto.Error{
				Type:    string(errors.InternalError),
				Message: "An unexpected error occurred",
			},
		}, nil
	}

	return &proto.PlaceOrderResponse{
		Success:    true,
		OrderId:    orderId,
		PaymentUrl: paymentUrl,
	}, nil
}

func (h *orderHandler) GenerateNewPaymentUrl(ctx context.Context, req *proto.GenerateNewPaymentUrlRequest) (*proto.GenerateNewPaymentUrlResponse, error) {
	orderId := req.OrderId
	customerId := req.CustomerId

	paymentUrl, err := h.orderService.GenerateNewPaymentUrl(orderId, customerId)
	if err != nil {
		if appErr, ok := errors.IsAppError(err); ok {
			return &proto.GenerateNewPaymentUrlResponse{
				Success: false,
				Error: &proto.Error{
					Type:    string(appErr.Type),
					Message: appErr.Message,
					Details: utils.ConvertMapToKeyValuePairs(appErr.Details),
				},
			}, nil
		}

		return &proto.GenerateNewPaymentUrlResponse{
			Success: false,
			Error: &proto.Error{
				Type:    string(errors.InternalError),
				Message: "An unexpected error occurred",
			},
		}, nil
	}

	return &proto.GenerateNewPaymentUrlResponse{
		Success:    true,
		PaymentUrl: paymentUrl,
	}, nil
}

func (h *orderHandler) GetOrder(ctx context.Context, req *proto.GetOrderRequest) (*proto.GetOrderResponse, error) {
	order, orderItems, err := h.orderService.GetOrderByID(req.OrderId)
	if err != nil {
		if appErr, ok := errors.IsAppError(err); ok {
			return &proto.GetOrderResponse{
				Success: false,
				Error: &proto.Error{
					Type:    string(appErr.Type),
					Message: appErr.Message,
					Details: utils.ConvertMapToKeyValuePairs(appErr.Details),
				},
			}, nil
		}

		return &proto.GetOrderResponse{
			Success: false,
			Error: &proto.Error{
				Type:    string(errors.InternalError),
				Message: "An unexpected error occurred",
			},
		}, nil
	}

	customerId := req.CustomerId

	if customerId != "admin" && order.CustomerID.String() != customerId {
		return &proto.GetOrderResponse{
			Success: false,
			Error: &proto.Error{
				Type:    string(errors.AuthError),
				Message: "You are not authorized to view this order",
			},
		}, nil
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
		Success:         true,
		OrderId:         order.ID.String(),
		CustomerId:      order.CustomerID.String(),
		Status:          order.Status,
		PrescriptionUrl: order.PrescriptionURL,
		ShippingCost:    order.ShippingCost,
		Subtotal:        order.Subtotal,
		Items:           protoOrderItems,
		CreatedAt:       order.CreatedAt.UnixMilli(),
		UpdatedAt:       order.UpdatedAt.UnixMilli(),
	}, nil
}

func (h *orderHandler) ListCustomersOrders(ctx context.Context, req *proto.ListCustomersOrdersRequest) (*proto.ListCustomersOrdersResponse, error) {
	var filter models.Filter
	if req.Filter != nil {
		filter = models.Filter{
			Column:   req.Filter.Column,
			Operator: req.Filter.Operator,
			Value:    req.Filter.Value,
		}
	}
	orders, total, err := h.orderService.ListCustomersOrders(req.CustomerId, filter, req.SortBy, req.SortOrder, req.Page, req.Limit)
	if err != nil {
		if appErr, ok := errors.IsAppError(err); ok {
			return &proto.ListCustomersOrdersResponse{
				Success: false,
				Error: &proto.Error{
					Type:    string(appErr.Type),
					Message: appErr.Message,
					Details: utils.ConvertMapToKeyValuePairs(appErr.Details),
				},
			}, nil
		}
		return &proto.ListCustomersOrdersResponse{
			Success: false,
			Error: &proto.Error{
				Type:    string(errors.InternalError),
				Message: "An unexpected error occurred",
			},
		}, nil
	}

	protoOrders := make([]*proto.Order, len(*orders))
	for i, order := range *orders {
		protoOrders[i] = &proto.Order{
			OrderId:         order.OrderID,
			CustomerId:      order.CustomerID,
			Status:          order.Status,
			PrescriptionUrl: order.PrescriptionURL,
			ShippingCost:    float64(order.ShippingCost),
			Subtotal:        float64(order.Subtotal),
			CreatedAt:       order.CreatedAt.UnixMilli(),
			UpdatedAt:       order.UpdatedAt.UnixMilli(),
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
		Success: true,
		Orders:  protoOrders,
		Total:   total,
		Page:    req.Page,
		Limit:   req.Limit,
	}, nil
}

func (h *orderHandler) ListAllOrders(ctx context.Context, req *proto.ListAllOrdersRequest) (*proto.ListAllOrdersResponse, error) {
	var filter models.Filter
	if req.Filter != nil {
		filter = models.Filter{
			Column:   req.Filter.Column,
			Operator: req.Filter.Operator,
			Value:    req.Filter.Value,
		}
	}
	orders, total, err := h.orderService.ListAllOrders(filter, req.SortBy, req.SortOrder, req.Page, req.Limit)
	if err != nil {
		if appErr, ok := errors.IsAppError(err); ok {
			return &proto.ListAllOrdersResponse{
				Success: false,
				Error: &proto.Error{
					Type:    string(appErr.Type),
					Message: appErr.Message,
					Details: utils.ConvertMapToKeyValuePairs(appErr.Details),
				},
			}, nil
		}
		return &proto.ListAllOrdersResponse{
			Success: false,
			Error: &proto.Error{
				Type:    string(errors.InternalError),
				Message: "An unexpected error occurred",
			},
		}, nil
	}

	protoOrders := make([]*proto.Order, len(*orders))
	for i, order := range *orders {
		protoOrders[i] = &proto.Order{
			OrderId:         order.OrderID,
			CustomerId:      order.CustomerID,
			Status:          order.Status,
			PrescriptionUrl: order.PrescriptionURL,
			ShippingCost:    float64(order.ShippingCost),
			Subtotal:        float64(order.Subtotal),
			CreatedAt:       order.CreatedAt.UnixMilli(),
			UpdatedAt:       order.UpdatedAt.UnixMilli(),
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
		Success: true,
		Orders:  protoOrders,
		Total:   total,
		Page:    req.Page,
		Limit:   req.Limit,
	}, nil
}

func (h *orderHandler) UpdateOrderStatus(ctx context.Context, req *proto.UpdateOrderStatusRequest) (*proto.UpdateOrderStatusResponse, error) {
	err := h.orderService.UpdateOrderStatus(req.OrderId, req.CustomerId, req.Status)
	if err != nil {
		if appErr, ok := errors.IsAppError(err); ok {
			return &proto.UpdateOrderStatusResponse{
				Success: false,
				Error: &proto.Error{
					Type:    string(appErr.Type),
					Message: appErr.Message,
					Details: utils.ConvertMapToKeyValuePairs(appErr.Details),
				},
			}, nil
		}
		return &proto.UpdateOrderStatusResponse{
			Success: false,
			Error: &proto.Error{
				Type:    string(errors.InternalError),
				Message: "An unexpected error occurred",
			},
		}, nil
	}

	return &proto.UpdateOrderStatusResponse{
		Success: true,
	}, nil
}
