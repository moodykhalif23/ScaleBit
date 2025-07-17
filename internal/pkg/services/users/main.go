package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"encoding/json"
	"strconv"

	_ "github.com/go-sql-driver/mysql"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/trace"

	"github.com/gorilla/mux"
	"github.com/moodykhalif23/sme-platform/internal/pkg/security"
	"github.com/moodykhalif23/sme-platform/internal/pkg/telemetry"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// User model
type User struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

func main() {
	// Initialize tracing
	tp := initTracer()
	defer tp.Shutdown(context.Background())

	// Database setup
	db := setupDB()
	defer db.Close()

	// Prometheus metrics
	meter := otel.GetMeterProvider().Meter("user-service")
	telemetry.InitMetrics(meter)

	r := mux.NewRouter()

	// Health endpoint
	r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	}).Methods("GET")

	// Prometheus metrics endpoint
	r.Handle("/metrics", promhttp.Handler())

	// User CRUD endpoints
	userRouter := r.PathPrefix("/users").Subrouter()
	userRouter.HandleFunc("", getUsers(db)).Methods("GET")
	userRouter.HandleFunc("", createUser(db)).Methods("POST")
	userRouter.HandleFunc("/{id:[0-9]+}", getUser(db)).Methods("GET")
	userRouter.HandleFunc("/{id:[0-9]+}", updateUser(db)).Methods("PUT")
	userRouter.HandleFunc("/{id:[0-9]+}", deleteUser(db)).Methods("DELETE")

	// Secure endpoints with JWT and metrics middleware
	handler := telemetry.Middleware(security.JWTValidationMiddleware(r))

	srv := &http.Server{
		Addr:    ":8080",
		Handler: handler,
	}

	go func() {
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	// Graceful shutdown
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
	// Replace these with your actual database credentials
	dsn := os.Getenv("DATABASE_DSN")
	if dsn == "" {
		dsn = "root:password@tcp(localhost:3306)/sme_platform?parseTime=true"
	}

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Test the connection
	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	return db
}

func authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get the Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Authorization header required", http.StatusUnauthorized)
			return
		}

		// Check if it's a Bearer token
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			http.Error(w, "Invalid authorization format", http.StatusUnauthorized)
			return
		}

		token := parts[1]
		// TODO: Implement proper token validation
		if token == "" {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		// Call the next handler
		next.ServeHTTP(w, r)
	})
}

// CRUD Handlers
func getUsers(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rows, err := db.Query("SELECT id, name, email FROM users")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()
		users := []User{}
		for rows.Next() {
			var u User
			if err := rows.Scan(&u.ID, &u.Name, &u.Email); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			users = append(users, u)
		}
		json.NewEncoder(w).Encode(users)
	}
}

func createUser(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var u User
		if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		result, err := db.Exec("INSERT INTO users (name, email) VALUES (?, ?)", u.Name, u.Email)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		id, _ := result.LastInsertId()
		u.ID = int(id)
		json.NewEncoder(w).Encode(u)
	}
}

func getUser(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id := vars["id"]
		var u User
		err := db.QueryRow("SELECT id, name, email FROM users WHERE id = ?", id).Scan(&u.ID, &u.Name, &u.Email)
		if err == sql.ErrNoRows {
			http.Error(w, "User not found", http.StatusNotFound)
			return
		} else if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(u)
	}
}

func updateUser(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id := vars["id"]
		var u User
		if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		_, err := db.Exec("UPDATE users SET name = ?, email = ? WHERE id = ?", u.Name, u.Email, id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		u.ID, _ = strconv.Atoi(id)
		json.NewEncoder(w).Encode(u)
	}
}

func deleteUser(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id := vars["id"]
		_, err := db.Exec("DELETE FROM users WHERE id = ?", id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}
