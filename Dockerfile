FROM golang:1.22.4-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN apk add --no-cache git && go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /usr/bin/service ./cmd/service

FROM alpine:3.19
RUN apk add --no-cache ca-certificates
COPY --from=builder /usr/bin/service /app/service
ENTRYPOINT ["/app/service"]