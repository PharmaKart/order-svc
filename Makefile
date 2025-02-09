# Variables
PROJECT_NAME = order-svc
GATEWAY_NAME = gateway-svc
PAYMENT_SERVICE_NAME = payment-svc
GO = go
PROTO_DIR = internal/proto
PROTO_FILE = $(PROTO_DIR)/order.proto
PROTO_OUT = $(PROTO_DIR)
PORT = 50053

# Targets
.PHONY: build run proto clean

# Build the service
build:
	@echo "Building $(PROJECT_NAME)..."
	$(GO) build -o bin/$(PROJECT_NAME) ./cmd/main.go

# Run the service
run: build
	@echo "Running $(PROJECT_NAME) on port $(PORT)..."
	./bin/$(PROJECT_NAME)

# Generate Go code from .proto file
proto:
	@echo "Generating Go code from $(PROTO_FILE)..."
	protoc --go_out=$(PROTO_OUT) --go-grpc_out=$(PROTO_OUT) $(PROTO_FILE)
	cp $(PROTO_DIR)/order.pb.go ../$(GATEWAY_NAME)/internal/proto/order.pb.go
	cp $(PROTO_DIR)/order_grpc.pb.go ../$(GATEWAY_NAME)/internal/proto/order_grpc.pb.go
	cp $(PROTO_DIR)/order.pb.go ../$(PAYMENT_SERVICE_NAME)/internal/proto/order.pb.go
	cp $(PROTO_DIR)/order_grpc.pb.go ../$(PAYMENT_SERVICE_NAME)/internal/proto/order_grpc.pb.go

# Clean up build artifacts
clean:
	@echo "Cleaning up..."
	rm -rf bin/$(PROJECT_NAME)