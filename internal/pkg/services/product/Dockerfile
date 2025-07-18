# Build stage
FROM golang:1.24.5 AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY internal/pkg/services/product/ ./product/
WORKDIR /app/product
RUN go build -o product-service main.go

# Run stage
FROM gcr.io/distroless/base-debian11
WORKDIR /app
COPY --from=builder /app/product-service .
EXPOSE 8081
ENTRYPOINT ["/app/product-service"] 