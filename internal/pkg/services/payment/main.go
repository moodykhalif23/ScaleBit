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
	"github.com/moodykhalif23/scalebit/internal/pkg/telemetry"

	_ "github.com/lib/pq"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/trace"
)

type Payment struct {
	ID        int       `json:"id"`
	OrderID   int       `json:"order_id"`
	Amount    float64   `json:"amount"`
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
}

func main() {
	tp := initTracer()
	defer tp.Shutdown(context.Background())

	db := setupDB()
	defer db.Close()

	meter := otel.GetMeterProvider().Meter("payment-service")
	telemetry.InitMetrics(meter)

	// Create public router for health and metrics endpoints
	publicRouter := mux.NewRouter()
	publicRouter.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	}).Methods("GET")
	publicRouter.Handle("/metrics", promhttp.Handler())

	paymentRouter := publicRouter.PathPrefix("/payments").Subrouter()
	paymentRouter.HandleFunc("", getPayments(db)).Methods("GET")
	paymentRouter.HandleFunc("", createPayment(db)).Methods("POST")
	paymentRouter.HandleFunc("/{id:[0-9]+}", getPayment(db)).Methods("GET")
	paymentRouter.HandleFunc("/{id:[0-9]+}", updatePayment(db)).Methods("PUT")
	paymentRouter.HandleFunc("/{id:[0-9]+}", deletePayment(db)).Methods("DELETE")

	handler := telemetry.Middleware(publicRouter)

	srv := &http.Server{
		Addr:    ":8083",
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
func getPayments(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rows, err := db.Query("SELECT id, order_id, amount, status, timestamp FROM payments")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()
		payments := []Payment{}
		for rows.Next() {
			var p Payment
			if err := rows.Scan(&p.ID, &p.OrderID, &p.Amount, &p.Status, &p.Timestamp); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			payments = append(payments, p)
		}
		json.NewEncoder(w).Encode(payments)
	}
}

func createPayment(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var p Payment
		if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		result, err := db.Exec("INSERT INTO payments (order_id, amount, status, timestamp) VALUES ($1, $2, $3, $4)", p.OrderID, p.Amount, p.Status, p.Timestamp)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		id, _ := result.LastInsertId()
		p.ID = int(id)
		json.NewEncoder(w).Encode(p)
	}
}

func getPayment(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id := vars["id"]
		var p Payment
		err := db.QueryRow("SELECT id, order_id, amount, status, timestamp FROM payments WHERE id = $1", id).Scan(&p.ID, &p.OrderID, &p.Amount, &p.Status, &p.Timestamp)
		if err == sql.ErrNoRows {
			http.Error(w, "Payment not found", http.StatusNotFound)
			return
		} else if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(p)
	}
}

func updatePayment(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id := vars["id"]
		var p Payment
		if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		_, err := db.Exec("UPDATE payments SET order_id = $1, amount = $2, status = $3, timestamp = $4 WHERE id = $5", p.OrderID, p.Amount, p.Status, p.Timestamp, id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		p.ID, _ = strconv.Atoi(id)
		json.NewEncoder(w).Encode(p)
	}
}

func deletePayment(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id := vars["id"]
		_, err := db.Exec("DELETE FROM payments WHERE id = $1", id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}
