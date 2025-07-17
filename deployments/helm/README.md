# sme-platform-istio Helm Chart

This Helm chart deploys Istio service mesh configuration for the SME Platform microservices, including:

- Namespace with Istio sidecar injection
- mTLS (mutual TLS) enforcement
- Canary traffic splitting for User, Product, Order, and Payment services

## Usage

1. **Install Istio** in your Kubernetes cluster (see [Istio docs](https://istio.io/latest/docs/setup/)).
2. **Deploy the chart:**

```sh
helm install sme-platform-istio ./ --namespace sme-platform --create-namespace
```

3. **Customize canary weights:**

Edit `values.yaml` to set the percentage of traffic to route to the canary (v2) version of each service:

```yaml
services:
  - name: user-service
    canaryWeight: 10 # 10% to v2, 90% to v1
```

4. **Upgrade canary rollout:**

Change the `canaryWeight` and run:

```sh
helm upgrade sme-platform-istio ./ --namespace sme-platform
```

## Files
- `templates/namespace.yaml`: Namespace with Istio injection
- `templates/mtls.yaml`: PeerAuthentication for mTLS
- `templates/service-destinationrule.yaml`: DestinationRules for all services
- `templates/service-virtualservice.yaml`: VirtualServices for all services

## Requirements
- Istio installed and running
- Kubernetes 1.20+ 