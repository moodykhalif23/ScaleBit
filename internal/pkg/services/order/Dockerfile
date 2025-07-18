# Build stage
FROM golang:1.24.5 AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY internal/pkg/services/order/ ./order/
WORKDIR /app/order
RUN go build -o order-service main.go

# Run stage
FROM gcr.io/distroless/base-debian11
WORKDIR /app
COPY --from=builder /app/order-service .
EXPOSE 8082
ENTRYPOINT ["/app/order-service"] 