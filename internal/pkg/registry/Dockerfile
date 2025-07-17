# Build stage
FROM golang:1.22 AS builder
WORKDIR /app
COPY . .
RUN go build -o registry-server server.go

# Run stage
FROM gcr.io/distroless/base-debian11
WORKDIR /app
COPY --from=builder /app/registry-server .
EXPOSE 8000
ENTRYPOINT ["/app/registry-server"] 