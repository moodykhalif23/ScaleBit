# ScaleBit: A Scalable Microservices Platform for SMEs

## 1. Project Overview

ScaleBit is a production-ready, open-source microservices platform designed to help small and medium enterprises (SMEs) accelerate the development, deployment, and scaling of their applications. It provides a robust foundation of core services, infrastructure automation, and modern DevOps practices, allowing teams to focus on building business value instead of managing infrastructure complexity.

This platform is built on a foundation of Go, Kubernetes, and other cloud-native technologies, offering a cost-effective and high-performance alternative to traditional monolithic architectures.

## 2. Architecture

The ScaleBit platform follows a distributed, microservices-based architecture. Key components include:

- **Core Services**: A set of essential microservices for common business functions:
  - **User Service**: Manages user authentication, authorization, and profiles.
  - **Product Service**: Handles product catalog and inventory management.
  - **Order Service**: Manages customer orders and orchestrates the fulfillment process.
  - **Payment Service**: Integrates with payment gateways to handle transactions.
- **API Gateway**: A central entry point for all client requests, managed by **KrakenD**. It handles routing, rate limiting, authentication, and CORS.
- **Service Mesh**: **Istio** is used to manage traffic between services, enforce security policies (mTLS), and enable advanced deployment strategies like canary releases.
- **Database**: Each microservice has its own dedicated database or schema, following the database-per-service pattern. This ensures loose coupling and independent scalability. Services should not access each other's databases directly; all inter-service communication is handled via APIs.
- **Observability**:
  - **Monitoring**: **Prometheus** scrapes and stores metrics from all services.
  - **Logging**: The **ELK Stack** (Elasticsearch, Logstash, Kibana) provides a centralized logging solution, with **Filebeat** shipping logs from each service.
- **Deployment**: The platform is designed for deployment on **Kubernetes**. It uses Docker for containerization and a custom Go-based operator for managing `Microservice` Custom Resource Definitions (CRDs).

## 3. Getting Started

Follow these steps to set up the platform for local development.

### Prerequisites

- Go (version 1.22 or later)
- Docker
- A local Kubernetes cluster (e.g., Minikube, k3s, Docker Desktop)
- A running MySQL instance

### Local Development Setup

1. **Clone the Repository**:
   ```sh
   git clone https://github.com/moodykhalif23/scalebit.git
   cd scalebit
   ```

2. **Database Setup**:
   Each service has its own database schema. While a single `main.sql` file is provided for initial setup, in a production environment, you should create separate databases or schemas for each service.

3. **Configure Environment**:
   Each microservice requires environment variables for database connections and other settings. You can create a `.env` file in each service's directory (e.g., `internal/pkg/services/users/.env`) or export them in your shell.

4. **Run a Single Service**:
   To build and run a specific service locally:
   ```sh
   cd internal/pkg/services/users
   go run main.go
   ```

5. **Build and Run with Docker**:
   Build Docker images for each service:
   ```sh
   docker build -t user-service:latest internal/pkg/services/users
   docker build -t product-service:latest internal/pkg/services/product
   docker build -t order-service:latest internal/pkg/services/order
   docker build -t payment-service:latest internal/pkg/services/payment
   ```
   A `docker-compose.yml` file is provided for running all services together.

## 4. Deployment

The platform is designed to be deployed on Kubernetes.

### Kubernetes Deployment with Helm

The recommended way to deploy the platform is by using the provided Helm chart, which leverages Istio for service mesh capabilities.

1. **Install Istio**:
   Ensure Istio is installed in your Kubernetes cluster.

2. **Deploy the Helm Chart**:
   The Helm chart will create the necessary namespace, service meshes, and deployment configurations.
   ```sh
   cd deployments/helm
   helm install scalebit-istio ./ --namespace scalebit --create-namespace
   ```

3. **Canary Deployments**:
   You can perform a canary rollout by adjusting the traffic weights in `deployments/helm/values.yaml` and running a Helm upgrade:
   ```sh
   helm upgrade scalebit-istio ./ --namespace scalebit
   ```

### API Gateway Configuration

The **KrakenD** API Gateway is configured via `deployments/aws/krakend.json`. To run the gateway:
```sh
docker run -p 80:80 -v $(pwd)/deployments/aws/krakend.json:/etc/krakend/krakend.json devopsfaith/krakend
```
Ensure the service endpoints in the configuration file match your deployment environment.

## 5. Contribution Guidelines

We welcome contributions to the ScaleBit platform. To contribute, please follow these guidelines:

- **Fork the repository** and create a new branch for your feature or bug fix.
- **Follow the existing code style** and architectural patterns.
- **Write clear and professional commit messages**.
- **Ensure all tests pass** before submitting a pull request.
- **Submit a pull request** for review.

## 6. License

This project is licensed under the MIT License. See the `LICENSE` file for more details.
