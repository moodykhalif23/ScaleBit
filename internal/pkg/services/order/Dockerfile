# Build stage
FROM golang:1.22 AS builder
WORKDIR /app
COPY . .
RUN go build -o order-service main.go

# Run stage
FROM gcr.io/distroless/base-debian11
WORKDIR /app
COPY --from=builder /app/order-service .
EXPOSE 8082
ENTRYPOINT ["/app/order-service"] 