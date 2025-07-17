# Build stage
FROM golang:1.22 AS builder
WORKDIR /app
COPY . .
RUN go build -o product-service main.go

# Run stage
FROM gcr.io/distroless/base-debian11
WORKDIR /app
COPY --from=builder /app/product-service .
EXPOSE 8081
ENTRYPOINT ["/app/product-service"] 