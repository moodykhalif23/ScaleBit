# Build stage
FROM golang:1.22 AS builder
WORKDIR /app
COPY . .
RUN go build -o {{SERVICE_NAME}} main.go

# Run stage
FROM gcr.io/distroless/base-debian11
WORKDIR /app
COPY --from=builder /app/{{SERVICE_NAME}} .
EXPOSE 8080
ENTRYPOINT ["/app/{{SERVICE_NAME}}"] 