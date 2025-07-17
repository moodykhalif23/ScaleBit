# Scalable Microservices Platform for SMEs Using Golang

## Introduction

Small and medium enterprises (SMEs) often struggle with scaling their applications as their businesses grow. Monolithic architectures can become inefficient, and adopting microservices can seem daunting due to complexity and resource demands. This document outlines a scalable microservices platform built with Golang, designed specifically for SMEs. It offers high performance, reliability, and simplicity, empowering SMEs to compete with larger organizations.

### Benefits for SMEs

- **Cost-Effective**: Minimizes infrastructure and staffing needs.
- **Scalable**: Adapts to growing loads with additional service instances.
- **Reliable**: Incorporates fault tolerance and monitoring.
- **Simple to Manage**: Uses intuitive tools suited for smaller teams.

### Why Golang?

Golang is ideal due to its simplicity, high performance, and efficient concurrency model. Its compiled nature ensures fast execution, and its standard library supports robust networking and concurrency, making it perfect for resource-limited environments.

---

## Architecture Overview

The platform includes the following components:

- **Microservices**: Independent services for specific business functions (e.g., user management, product catalog).
- **API Gateway**: Central entry point for routing, authentication, and rate limiting.
- **Service Discovery**: A registry for services to register and be located by the API gateway.
- **Load Balancing**: Distributes requests across service instances for scalability.
- **Monitoring**: Tools for logging, metrics, and alerting to ensure system health.

*Diagram Placeholder: Imagine a flowchart with Microservices connected to an API Gateway, linked to Service Discovery, Load Balancing, and Monitoring.*

---

## Services

SMEs typically require services like:

- **User Management**: User registration, authentication, profiles.
- **Product Catalog**: Product listings, categories, inventory.
- **Order Processing**: Order handling, payment, fulfillment.
- **Payment Handling**: Secure payment gateway integration.

### Example: User Management Service

Below is a Golang microservice implementing CRUD operations for user management:

```go
package main

import (
    "database/sql"
    "encoding/json"
    "log"
    "net/http"

    _ "github.com/go-sql-driver/mysql"
)

type User struct {
    ID    int    `json:"id"`
    Name  string `json:"name"`
    Email string `json:"email"`
}

var db *sql.DB

func main() {
    var err error
    db, err = sql.Open("mysql", "user:password@tcp(127.0.0.1:3306)/dbname")
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()

    http.HandleFunc("/users", getUsers)
    http.HandleFunc("/users/create", createUser)

    log.Println("Starting User Service on :8080")
    log.Fatal(http.ListenAndServe(":8080", nil))
}

func getUsers(w http.ResponseWriter, r *http.Request) {
    rows, err := db.Query("SELECT id, name, email FROM users")
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    defer rows.Close()

    users := []User{}
    for rows.Next() {
        var u User
        if err := rows.Scan(&u.ID, &u.Name, &u.Email); err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }
        users = append(users, u)
    }
    json.NewEncoder(w).Encode(users)
}

func createUser(w http.ResponseWriter, r *http.Request) {
    var u User
    if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }
    result, err := db.Exec("INSERT INTO users (name, email) VALUES (?, ?)", u.Name, u.Email)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    id, _ := result.LastInsertId()
    u.ID = int(id)
    json.NewEncoder(w).Encode(u)
}
```

---

## API Gateway

The API gateway routes client requests to microservices and handles concerns like authentication. We recommend **KrakenD**, a lightweight, Golang-based gateway.

### Configuration Example

A KrakenD configuration with routing and JWT authentication:

```json
{
  "version": 2,
  "port": 80,
  "endpoints": [
    {
      "endpoint": "/users",
      "method": "GET",
      "backend": [
        {
          "url_pattern": "/users",
          "host": ["http://user-service:8080"]
        }
      ]
    },
    {
      "endpoint": "/products",
      "method": "GET",
      "backend": [
        {
          "url_pattern": "/products",
          "host": ["http://product-service:8081"]
        }
      ]
    }
  ],
  "extra_config": {
    "github.com/devopsfaith/krakend-jwt": {
      "keys": ["your_jwt_secret"]
    }
  }
}
```

---

## Service Discovery

A simple registry allows services to register themselves and be discovered by the API gateway.

### Registry Server Example

```go
package main

import (
    "encoding/json"
    "log"
    "net/http"
    "sync"
)

var (
    services = make(map[string][]string)
    mu       sync.Mutex
)

func registerService(w http.ResponseWriter, r *http.Request) {
    var data struct {
        Service string `json:"service"`
        Address string `json:"address"`
    }
    if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }
    mu.Lock()
    services[data.Service] = append(services[data.Service], data.Address)
    mu.Unlock()
    w.WriteHeader(http.StatusOK)
}

func main() {
    http.HandleFunc("/register", registerService)
    log.Println("Starting Service Registry on :8000")
    log.Fatal(http.ListenAndServe(":8000", nil))
}
```

---

## Load Balancing

A Golang reverse proxy can distribute requests across service instances.

### Load Balancer Example

```go
package main

import (
    "log"
    "net/http"
    "net/http/httputil"
    "net/url"
)

func main() {
    targets := []string{"http://service1:8080", "http://service2:8080"}
    var proxies []*httputil.ReverseProxy
    for _, target := range targets {
        url, _ := url.Parse(target)
        proxies = append(proxies, httputil.NewSingleHostReverseProxy(url))
    }

    var index int
    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        proxy := proxies[index%len(proxies)]
        index++
        proxy.ServeHTTP(w, r)
    })

    log.Println("Starting Load Balancer on :9000")
    log.Fatal(http.ListenAndServe(":9000", nil))
}
```

---

## Monitoring

Use **Prometheus** for metrics and the **ELK Stack** (Elasticsearch, Logstash, Kibana) for logging.

### Prometheus Instrumentation

Add to your service:

```go
import (
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
    httpRequestsTotal = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "http_requests_total",
            Help: "Total number of HTTP requests.",
        },
        []string{"method", "endpoint"},
    )
)

func init() {
    prometheus.MustRegister(httpRequestsTotal)
}

func main() {
    http.Handle("/metrics", promhttp.Handler())
    // ... rest of your service code
}
```

For logging, configure services to send logs to Logstash.

---

## Deployment

Use **Docker** for containerization and **Kubernetes** for orchestration, simplified for SMEs.

### Dockerfile Example

```dockerfile
FROM golang:1.18
WORKDIR /app
COPY . .
RUN go build -o service
CMD ["./service"]
```

### Docker Compose Example

```yaml
version: '3'
services:
  user-service:
    build: ./user-service
    ports:
      - "8080:8080"
  api-gateway:
    build: ./api-gateway
    ports:
      - "80:80"
```

### Kubernetes Deployment Example

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: user-service
spec:
  replicas: 3
  selector:
    matchLabels:
      app: user-service
  template:
    metadata:
      labels:
        app: user-service
    spec:
      containers:
      - name: user-service
        image: user-service:latest
        ports:
        - containerPort: 8080
---
apiVersion: v1
kind: Service
metadata:
  name: user-service
spec:
  selector:
    app: user-service
  ports:
    - protocol: TCP
      port: 80
      targetPort: 8080
```

---

## Security

### Authentication Middleware (JWT)

```go
func authMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        tokenStr := r.Header.Get("Authorization")
        // Add JWT validation logic here
        // If valid, next.ServeHTTP(w, r)
        // Else, http.Error(w, "Unauthorized", http.StatusUnauthorized)
    })
}
```

- Use HTTPS for all communications.
- Encrypt sensitive data at rest and in transit.

---

## Data Management

- **Databases**: Use `github.com/go-sql-driver/mysql` for MySQL or `github.com/lib/pq` for PostgreSQL.
- **Caching**: Use Redis with `github.com/go-redis/redis`.
- **Message Queues**: Use RabbitMQ with `github.com/streadway/amqp`.

---

## Step-by-Step Setup Guide

1. **Build and Deploy Services**:

   - Clone the repo, run `go build` in each service directory.
   - Containerize with Docker.

2. **Configure API Gateway**:

   - Set up routing and integrate with the registry.

3. **Set Up Monitoring**:

   - Install Prometheus and ELK Stack.

4. **Deploy**:

   - Use Docker Compose locally or Kubernetes for production.

---

## Conclusion

This platform equips SMEs with a scalable, reliable microservices solution using Golang. Future enhancements could include automated testing, CI/CD pipelines, and advanced monitoring like distributed tracing.