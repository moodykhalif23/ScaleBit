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

type Order struct {
	ID        int    `json:"id"`
	UserID    int    `json:"user_id"`
	ProductID int    `json:"product_id"`
	Quantity  int    `json:"quantity"`
	Status    string `json:"status"`
}

func main() {
	tp := initTracer()
	defer tp.Shutdown(context.Background())

	db := setupDB()
	defer db.Close()

	meter := otel.GetMeterProvider().Meter("order-service")
	telemetry.InitMetrics(meter)

	// Create public router for health and metrics endpoints
	publicRouter := mux.NewRouter()
	publicRouter.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	}).Methods("GET")
	publicRouter.Handle("/metrics", promhttp.Handler())

	orderRouter := publicRouter.PathPrefix("/orders").Subrouter()
	orderRouter.HandleFunc("", getOrders(db)).Methods("GET")
	orderRouter.HandleFunc("", createOrder(db)).Methods("POST")
	orderRouter.HandleFunc("/{id:[0-9]+}", getOrder(db)).Methods("GET")
	orderRouter.HandleFunc("/{id:[0-9]+}", updateOrder(db)).Methods("PUT")
	orderRouter.HandleFunc("/{id:[0-9]+}", deleteOrder(db)).Methods("DELETE")

	handler := telemetry.Middleware(publicRouter)

	srv := &http.Server{
		Addr:    ":8082",
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
func getOrders(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rows, err := db.Query("SELECT id, user_id, product_id, quantity, status FROM orders")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()
		orders := []Order{}
		for rows.Next() {
			var o Order
			if err := rows.Scan(&o.ID, &o.UserID, &o.ProductID, &o.Quantity, &o.Status); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			orders = append(orders, o)
		}
		json.NewEncoder(w).Encode(orders)
	}
}

func createOrder(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var o Order
		if err := json.NewDecoder(r.Body).Decode(&o); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		result, err := db.Exec("INSERT INTO orders (user_id, product_id, quantity, status) VALUES ($1, $2, $3, $4)", o.UserID, o.ProductID, o.Quantity, o.Status)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		id, _ := result.LastInsertId()
		o.ID = int(id)
		json.NewEncoder(w).Encode(o)
	}
}

func getOrder(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id := vars["id"]
		var o Order
		err := db.QueryRow("SELECT id, user_id, product_id, quantity, status FROM orders WHERE id = ?", id).Scan(&o.ID, &o.UserID, &o.ProductID, &o.Quantity, &o.Status)
		if err == sql.ErrNoRows {
			http.Error(w, "Order not found", http.StatusNotFound)
			return
		} else if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(o)
	}
}

func updateOrder(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id := vars["id"]
		var o Order
		if err := json.NewDecoder(r.Body).Decode(&o); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		_, err := db.Exec("UPDATE orders SET user_id = ?, product_id = ?, quantity = ?, status = ? WHERE id = ?", o.UserID, o.ProductID, o.Quantity, o.Status, id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		o.ID, _ = strconv.Atoi(id)
		json.NewEncoder(w).Encode(o)
	}
}

func deleteOrder(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id := vars["id"]
		_, err := db.Exec("DELETE FROM orders WHERE id = ?", id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}
