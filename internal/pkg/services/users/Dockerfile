# Build stage
FROM golang:1.22 AS builder
WORKDIR /app
COPY . .
COPY go.mod go.sum ./
RUN go mod download
RUN go build -o user-service main.go

# Run stage
FROM gcr.io/distroless/base-debian11
WORKDIR /app
COPY --from=builder /app/user-service .
EXPOSE 8080
ENTRYPOINT ["/app/user-service"]
