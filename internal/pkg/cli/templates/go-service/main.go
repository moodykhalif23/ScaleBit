package main

import (
	"fmt"
	"net/http"
	"os"
)

func main() {
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "{{PORT}}"
	}
	fmt.Printf("Starting {{SERVICE_NAME}} on :%s\n", port)
	http.ListenAndServe(":"+port, nil)
}
