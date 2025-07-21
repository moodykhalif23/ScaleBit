# Build stage
FROM golang:1.24.5 AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY internal/ ./internal/
COPY internal/pkg/services/order/ ./order/
WORKDIR /app/order
RUN CGO_ENABLED=0 GOOS=linux go build -o order-service main.go

# Run stage
FROM debian:bullseye-slim
WORKDIR /app
COPY --from=builder /app/order/order-service .
EXPOSE 8080
ENTRYPOINT ["/app/order-service"] 