package main

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
)

// Service backends (Kubernetes service DNS names)
var (
	userBackends    = []string{"http://user-service:8080"}
	productBackends = []string{"http://product-service:8081"}
	orderBackends   = []string{"http://order-service:8082"}
	paymentBackends = []string{"http://payment-service:8083"}
)

func main() {
	http.HandleFunc("/users", makeProxyHandler(userBackends))
	http.HandleFunc("/products", makeProxyHandler(productBackends))
	http.HandleFunc("/orders", makeProxyHandler(orderBackends))
	http.HandleFunc("/payments", makeProxyHandler(paymentBackends))

	log.Println("Starting Load Balancer on :9000")
	log.Fatal(http.ListenAndServe(":9000", nil))
}

func makeProxyHandler(backends []string) http.HandlerFunc {
	proxies := make([]*httputil.ReverseProxy, len(backends))
	for i, backend := range backends {
		u, _ := url.Parse(backend)
		proxies[i] = httputil.NewSingleHostReverseProxy(u)
	}
	return func(w http.ResponseWriter, r *http.Request) {
		proxies[0].ServeHTTP(w, r)
	}
}
