# Build stage
FROM golang:1.24.5 AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY internal/ ./internal/
COPY internal/pkg/services/product/ ./product/
WORKDIR /app/product
RUN go build -o product-service main.go

# Run stage
FROM debian:bullseye-slim
WORKDIR /app
COPY --from=builder /app/product/product-service .
EXPOSE 8081
ENTRYPOINT ["/app/product-service"] 