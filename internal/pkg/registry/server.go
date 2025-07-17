package registry

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

type Service struct {
	Name    string
	Address string
	Status  string // "UP", "DOWN"
}

var (
	services = make(map[string][]Service)
	mu       sync.RWMutex
)

func RegisterService(serviceName, address string) {
	mu.Lock()
	defer mu.Unlock()

	services[serviceName] = append(services[serviceName], Service{
		Name:    serviceName,
		Address: address,
		Status:  "UP",
	})
}

func HealthCheck() {
	ticker := time.NewTicker(30 * time.Second)
	for {
		<-ticker.C
		mu.Lock()
		for svcName, instances := range services {
			for i, inst := range instances {
				resp, err := http.Get(inst.Address + "/health")
				if err != nil || resp.StatusCode != 200 {
					services[svcName][i].Status = "DOWN"
				} else {
					services[svcName][i].Status = "UP"
				}
			}
		}
		mu.Unlock()
	}
}

func main() {
	// Start health checks in background
	go HealthCheck()

	mux := http.NewServeMux()
	mux.HandleFunc("/register", registerServiceHandler)
	mux.HandleFunc("/services", listServicesHandler)
	mux.HandleFunc("/services/", listServiceInstancesHandler)

	srv := &http.Server{
		Addr:    ":8000",
		Handler: mux,
	}

	go func() {
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatalf("Registry server failed: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down registry server...")
	srv.Close()
}

func registerServiceHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		Service string `json:"service"`
		Address string `json:"address"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	RegisterService(req.Service, req.Address)
	w.WriteHeader(http.StatusOK)
}

func listServicesHandler(w http.ResponseWriter, r *http.Request) {
	mu.RLock()
	defer mu.RUnlock()
	json.NewEncoder(w).Encode(services)
}

func listServiceInstancesHandler(w http.ResponseWriter, r *http.Request) {
	// URL: /services/{name}
	name := r.URL.Path[len("/services/"):]
	mu.RLock()
	defer mu.RUnlock()
	instances, ok := services[name]
	if !ok {
		http.Error(w, "Service not found", http.StatusNotFound)
		return
	}
	json.NewEncoder(w).Encode(instances)
}
