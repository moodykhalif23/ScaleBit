package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/moodykhalif23/scalebit/internal/pkg/security"
	"github.com/moodykhalif23/scalebit/internal/pkg/telemetry"

	_ "github.com/lib/pq"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/trace"
)

type Product struct {
	ID    int     `json:"id"`
	Name  string  `json:"name"`
	Price float64 `json:"price"`
	Stock int     `json:"stock"`
}

func main() {
	tp := initTracer()
	defer tp.Shutdown(context.Background())

	db := setupDB()
	defer db.Close()

	meter := otel.GetMeterProvider().Meter("product-service")
	telemetry.InitMetrics(meter)

	// Create public router for health and metrics endpoints
	publicRouter := mux.NewRouter()
	publicRouter.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	}).Methods("GET")
	publicRouter.Handle("/metrics", promhttp.Handler())

	productRouter := publicRouter.PathPrefix("/products").Subrouter()
	productRouter.HandleFunc("", getProducts(db)).Methods("GET")
	productRouter.HandleFunc("", createProduct(db)).Methods("POST")
	productRouter.HandleFunc("/{id:[0-9]+}", getProduct(db)).Methods("GET")
	productRouter.HandleFunc("/{id:[0-9]+}", updateProduct(db)).Methods("PUT")
	productRouter.HandleFunc("/{id:[0-9]+}", deleteProduct(db)).Methods("DELETE")

	handler := telemetry.Middleware(publicRouter)

	srv := &http.Server{
		Addr:    ":8081",
		Handler: handler,
	}

	go func() {
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	srv.Shutdown(ctx)
}

func initTracer() *trace.TracerProvider {
	exporter, _ := otlptracegrpc.New(context.Background())
	tp := trace.NewTracerProvider(
		trace.WithBatcher(exporter),
	)
	otel.SetTracerProvider(tp)
	return tp
}

func setupDB() *sql.DB {
	dsn := os.Getenv("DATABASE_DSN")
	if dsn == "" {
		dsn = "postgres://postgres:password@localhost:5432/scalebit_platform?sslmode=disable"
	}

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Verify the connection
	err = db.Ping()
	if err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	return db
}

// CRUD Handlers
func getProducts(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rows, err := db.Query("SELECT id, name, price, stock FROM products")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()
		products := []Product{}
		for rows.Next() {
			var p Product
			if err := rows.Scan(&p.ID, &p.Name, &p.Price, &p.Stock); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			products = append(products, p)
		}
		json.NewEncoder(w).Encode(products)
	}
}

func createProduct(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var p Product
		if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		result, err := db.Exec("INSERT INTO products (name, price, stock) VALUES ($1, $2, $3)", p.Name, p.Price, p.Stock)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		id, _ := result.LastInsertId()
		p.ID = int(id)
		json.NewEncoder(w).Encode(p)
	}
}

func getProduct(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id := vars["id"]
		var p Product
		err := db.QueryRow("SELECT id, name, price, stock FROM products WHERE id = ?", id).Scan(&p.ID, &p.Name, &p.Price, &p.Stock)
		if err == sql.ErrNoRows {
			http.Error(w, "Product not found", http.StatusNotFound)
			return
		} else if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(p)
	}
}

func updateProduct(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id := vars["id"]
		var p Product
		if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		_, err := db.Exec("UPDATE products SET name = ?, price = ?, stock = ? WHERE id = ?", p.Name, p.Price, p.Stock, id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		p.ID, _ = strconv.Atoi(id)
		json.NewEncoder(w).Encode(p)
	}
}

func deleteProduct(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id := vars["id"]
		_, err := db.Exec("DELETE FROM products WHERE id = ?", id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}
