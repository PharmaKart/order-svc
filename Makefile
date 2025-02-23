# Variables
PROJECT_NAME = order-svc
GO = go
PROTO_DIR = internal/proto
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
	@echo "Generating Go code from Proto files..."
	protoc --go_out=$(PROTO_OUT) --go-grpc_out=$(PROTO_OUT) $(PROTO_DIR)/*.proto

# Clean up build artifacts
clean:
	@echo "Cleaning up..."
	rm -rf bin/$(PROJECT_NAME)