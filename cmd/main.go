package main

import (
	"net"

	"github.com/PharmaKart/order-svc/internal/handlers"
	"github.com/PharmaKart/order-svc/internal/proto"
	"github.com/PharmaKart/order-svc/internal/repositories"
	"github.com/PharmaKart/order-svc/pkg/config"
	"github.com/PharmaKart/order-svc/pkg/utils"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"gorm.io/gorm"
)

func main() {
	// Initialize logger
	utils.InitLogger()

	// Load configuration
	cfg := config.LoadConfig()

	// Initialize repositories
	orderRepo := repositories.NewOrderRepository(&gorm.DB{})
	orderItemRepo := repositories.NewOrderItemRepository(&gorm.DB{})

	// Initialize product client
	conn, err := grpc.NewClient(cfg.ProductServiceURL, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		utils.Logger.Fatal("Failed to connect to product service", map[string]interface{}{
			"error": err,
		})
	}

	productClient := proto.NewProductServiceClient(conn)
	defer conn.Close()

	// Initialize handlers
	orderHandler := handlers.NewOrderHandler(orderRepo, orderItemRepo, &productClient)

	// Initialize gRPC server
	lis, err := net.Listen("tcp", ":"+cfg.Port)

	if err != nil {
		utils.Logger.Fatal("Failed to listen", map[string]interface{}{
			"error": err,
		})
	}

	grpcServer := grpc.NewServer()
	proto.RegisterOrderServiceServer(grpcServer, orderHandler)

	utils.Info("Starting order service", map[string]interface{}{
		"port": cfg.Port,
	})

	if err := grpcServer.Serve(lis); err != nil {
		utils.Logger.Fatal("Failed to serve", map[string]interface{}{
			"error": err,
		})
	}
}
