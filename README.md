 # SME Platform: Scalable Microservices Platform for SMEs

## Overview

The SME Platform is a turnkey, production-ready microservices platform designed for small and medium enterprises. Built with Go, Kubernetes, and modern DevOps tooling, it enables rapid development, deployment, and scaling of business-critical services with minimal operational overhead.

## Architecture

- **Microservices**: User, Product, Order, and Payment services (Go, REST, OpenTelemetry, Prometheus)
- **API Gateway**: KrakenD for routing, authentication, rate limiting, and CORS
- **Service Discovery & Load Balancing**: Kubernetes DNS and a Go-based reverse proxy
- **Monitoring**: Prometheus for metrics, ELK Stack (Elasticsearch, Logstash, Kibana) for logging
- **Database**: MySQL (schema in `deployments/aws/main.sql`)
- **Deployment**: Docker, Kubernetes manifests, and operator-based CRDs

## Services

- **User Service**: CRUD for users, JWT auth, metrics
- **Product Service**: CRUD for products, metrics
- **Order Service**: CRUD for orders, metrics
- **Payment Service**: CRUD for payments, metrics

Each service exposes RESTful endpoints and a `/metrics` endpoint for Prometheus.

## Getting Started

### Prerequisites
- Go 1.22+
- Docker
- Kubernetes (minikube, k3s, or cloud provider)
- MySQL database

### Local Development
1. **Clone the repository**
2. **Build and run a service**:
   ```sh
   cd internal/pkg/services/users
   go run main.go
   ```
3. **Run all services with Docker Compose** (if provided) or build images manually:
   ```sh
   docker build -t user-service:latest internal/pkg/services/users
   docker build -t product-service:latest internal/pkg/services/product
   docker build -t order-service:latest internal/pkg/services/order
   docker build -t payment-service:latest internal/pkg/services/payment
   ```
4. **Apply the database schema**:
   ```sh
   mysql < deployments/aws/main.sql
   ```

### Kubernetes Deployment
1. **Build and push Docker images to your registry**
2. **Apply CRDs and manifests**:
   ```sh
   kubectl apply -f internal/pkg/services/users/microservice.yaml
   kubectl apply -f internal/pkg/services/product/microservice.yaml
   kubectl apply -f internal/pkg/services/order/microservice.yaml
   kubectl apply -f internal/pkg/services/payment/microservice.yaml
   kubectl apply -f deployments/aws/elasticsearch.yaml
   kubectl apply -f deployments/aws/kibana.yaml
   # Add other manifests as needed
   ```
3. **Deploy API Gateway**:
   - Use the KrakenD config in `deployments/aws/krakend.json`
   - Example Docker run:
     ```sh
     docker run -p 80:80 -v $(pwd)/deployments/aws/krakend.json:/etc/krakend/krakend.json devopsfaith/krakend
     ```

### Monitoring & Logging
- **Prometheus**: Use `deployments/aws/prometheus.yaml` to scrape metrics from all services
- **ELK Stack**:
  - Filebeat: `deployments/aws/filebeat.yaml`
  - Logstash: `deployments/aws/logstash.conf`
  - Elasticsearch: `deployments/aws/elasticsearch.yaml`
  - Kibana: `deployments/aws/kibana.yaml`

## API Gateway
- **KrakenD** routes all API traffic and enforces JWT authentication, CORS, and rate limiting.
- Update `krakend.json` as needed for your environment and secrets.

## Database
- MySQL schema is provided in `deployments/aws/main.sql`.
- Update connection strings in each service as needed.

## Contribution Guidelines
- Fork the repository and create feature branches for changes
- Write clear, professional commit messages
- Ensure all code passes linting and tests
- Submit pull requests for review
- Follow the architecture and security best practices outlined in the documentation

## License
This project is licensed under the MIT License.

### What’s been generated:
- **Istio Manifests** (`deployments/istio/`): mTLS, traffic splitting, and canary support for each service.
- **Helm Chart** (`deployments/helm/`):
  - Templated namespace, mTLS, DestinationRule, and VirtualService for all services.
  - Configurable canary weights in `values.yaml`.
  - Usage documentation in `README.md`.

### How to use:
1. **Install Istio** in your cluster.
2. **Deploy the Helm chart**:
   ```sh
   cd deployments/helm
   helm install sme-platform-istio ./ --namespace sme-platform --create-namespace
   ```
3. **Adjust canary rollout** by editing `values.yaml` and running:
   ```sh
   helm upgrade sme-platform-istio ./ --namespace sme-platform
   ```

