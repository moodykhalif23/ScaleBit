# Build stage
FROM golang:1.24.5 AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY internal/ ./internal/
COPY internal/pkg/services/payment/ ./payment/
WORKDIR /app/payment
RUN CGO_ENABLED=0 GOOS=linux go build -o payment-service main.go

# Run stage
FROM debian:bullseye-slim
WORKDIR /app
COPY --from=builder /app/payment/payment-service .
EXPOSE 8080
ENTRYPOINT ["/app/payment-service"] 