# NodePay Facilitator (Go Backend)

This directory contains the reference implementation of the NodePay Server in Go. It provides the API endpoints for processing charges and the middleware for enforcing the x402 protocol.

## Features

-   **x402 Middleware**: Intercepts requests, checks for `Sign-In-With-X` or `Payment-Signature` headers, and returns `402 Payment Required` with the correct wire protocol (`PaymentRequired` JSON) if authentication is missing.
-   **API Handlers**:
    -   `POST /v1/charges`: Create new charges.
    -   `GET /v1/charges/{id}`: Retrieve charge status.
    -   `GET /v1/discovery`: Expose service metadata (x402 Discovery).
-   **In-Memory Store**: A simple concurrent map storage for demonstration purposes.

## Project Structure

-   `cmd/api/main.go`: Application entrypoint.
-   `internal/models`: Go structs matching `charges.yaml` and `x402_protocol.yaml` schemas.
-   `internal/handlers`: HTTP handlers for the API.
-   `internal/middleware`: Paywall logic.
-   `internal/store`: Data persistence layer.

## Getting Started

### Prerequisites

-   Go 1.21+

### Running the Server

```bash
go run cmd/api/main.go
```

The server will start on `http://localhost:8080`.

-   Discovery endpoint: `http://localhost:8080/v1/discovery`
-   Create Charge: `POST http://localhost:8080/v1/charges`

### Running Tests

```bash
go test ./...
```
