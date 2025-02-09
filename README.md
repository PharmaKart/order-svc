# Order Service

The **Order Service** is a critical component of the Pharmakart platform, responsible for managing customer orders, order statuses, and prescription uploads. It provides secure endpoints for creating, retrieving, updating, and listing orders, as well as handling prescription-related operations.

---

## Table of Contents
1. [Overview](#overview)
2. [Features](#features)
3. [Prerequisites](#prerequisites)
4. [Setup and Installation](#setup-and-installation)
5. [Running the Service](#running-the-service)
6. [Environment Variables](#environment-variables)
7. [Contributing](#contributing)
8. [License](#license)

---

## Overview

The Order Service handles:
- Order management (create, retrieve, update, list).
- Prescription management (upload, approve/reject).
- Integration with the **Product Service** for inventory updates.

It is built using **gRPC** for communication and **PostgreSQL** for data storage.

---

## Features

- **Order Management**:
  - Create, retrieve, update, and list orders.
  - Update order status (e.g., pending, shipped, delivered, canceled).
- **Prescription Management**:
  - Upload prescriptions for orders.
  - Approve or reject prescriptions.
- **Inventory Integration**:
  - Automatically update product inventory when an order is created.

---

## Prerequisites

Before setting up the service, ensure you have the following installed:
- **Docker**
- **Go** (for building and running the service)
- **Protobuf Compiler** (`protoc`) for generating gRPC/protobuf files

---

## Setup and Installation

### 1. Clone the Repository
Clone the repository and navigate to the order service directory:
```bash
git clone https://github.com/PharmaKart/order-svc.git
cd order-svc
```

### 2. Generate Protobuf Files
Generate the protobuf files using the provided `Makefile`:
```bash
make proto
```

### 3. Install Dependencies
Run the following command to ensure all dependencies are installed:
```bash
go mod tidy
```

### 4. Build the Service
To build the service, run:
```bash
make build
```

---

## Running the Service

### Option 1: Run Using Docker
To run the service using Docker, execute:
```bash
docker run -p 50053:50053 pharmakart/order-svc
```

### Option 2: Run Using Makefile
To run the service directly using Go, execute:
```bash
make run
```

The service will be available at:
- **gRPC**: `localhost:50053`

---

## Environment Variables

The service requires the following environment variables. Create a `.env` file in the `order-svc` directory with the following:

```env
ORDER_DB_HOST=postgres
ORDER_DB_PORT=5432
ORDER_DB_USER=postgres
ORDER_DB_PASSWORD=yourpassword
ORDER_DB_NAME=pharmakartdb
PRODUCT_SERVICE_URL=localhost:50052
REMINDER_SERVICE_URL=localhost:50055
```

---

## Contributing

Contributions are welcome! Please follow these steps:
1. Fork the repository.
2. Create a new branch for your feature or bugfix.
3. Submit a pull request with a detailed description of your changes.

---

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.

---

## Support

For any questions or issues, please open an issue in the repository or contact the maintainers.