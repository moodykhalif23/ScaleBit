package registry

import (
	"net/http"
	"sync"
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
