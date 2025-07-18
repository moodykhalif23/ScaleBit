# Build stage
FROM golang:1.24.5 AS builder
WORKDIR /app
COPY . .
COPY go.mod go.sum ./
RUN go mod download
RUN go build -o product-service main.go

# Run stage
FROM gcr.io/distroless/base-debian11
WORKDIR /app
COPY --from=builder /app/product-service .
EXPOSE 8081
ENTRYPOINT ["/app/product-service"] 