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

	"golang.org/x/crypto/bcrypt"

	_ "github.com/go-sql-driver/mysql"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/trace"

	"github.com/golang-jwt/jwt/v5"

	"github.com/gorilla/mux"
	"github.com/moodykhalif23/scalebit/internal/pkg/security"
	"github.com/moodykhalif23/scalebit/internal/pkg/telemetry"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// User model
type User struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

// Add Register and Login request structs

type RegisterRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
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

	// Add /register and /login handlers
	r.HandleFunc("/register", registerHandler(db)).Methods("POST")
	r.HandleFunc("/login", loginHandler(db)).Methods("POST")

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
		dsn = "root:password@tcp(localhost:3306)/scalebit_platform?parseTime=true"
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

// Update createUser to require password and hash it
func createUser(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req RegisterRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if req.Name == "" || req.Email == "" || req.Password == "" {
			http.Error(w, "All fields required", http.StatusBadRequest)
			return
		}
		hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			http.Error(w, "Failed to hash password", http.StatusInternalServerError)
			return
		}
		result, err := db.Exec("INSERT INTO users (name, email, password_hash) VALUES (?, ?, ?)", req.Name, req.Email, string(hash))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		id, _ := result.LastInsertId()
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]interface{}{"id": id, "name": req.Name, "email": req.Email})
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

// Add /register and /login handlers
func registerHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req RegisterRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if req.Name == "" || req.Email == "" || req.Password == "" {
			http.Error(w, "All fields required", http.StatusBadRequest)
			return
		}
		// Hash password
		hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			http.Error(w, "Failed to hash password", http.StatusInternalServerError)
			return
		}
		result, err := db.Exec("INSERT INTO users (name, email, password_hash) VALUES (?, ?, ?)", req.Name, req.Email, string(hash))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		id, _ := result.LastInsertId()
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]interface{}{"id": id, "name": req.Name, "email": req.Email})
	}
}

func loginHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req LoginRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if req.Email == "" || req.Password == "" {
			http.Error(w, "Email and password required", http.StatusBadRequest)
			return
		}
		var id int
		var name, email, passwordHash string
		err := db.QueryRow("SELECT id, name, email, password_hash FROM users WHERE email = ?", req.Email).Scan(&id, &name, &email, &passwordHash)
		if err == sql.ErrNoRows {
			http.Error(w, "Invalid credentials", http.StatusUnauthorized)
			return
		} else if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(req.Password)); err != nil {
			http.Error(w, "Invalid credentials", http.StatusUnauthorized)
			return
		}
		// Create JWT
		jwtSecret := os.Getenv("JWT_SECRET")
		if jwtSecret == "" {
			http.Error(w, "JWT secret not set", http.StatusInternalServerError)
			return
		}
		claims := map[string]interface{}{
			"id":    id,
			"name":  name,
			"email": email,
			"exp":   time.Now().Add(24 * time.Hour).Unix(),
		}
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims(claims))
		tokenString, err := token.SignedString([]byte(jwtSecret))
		if err != nil {
			http.Error(w, "Failed to sign token", http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(map[string]string{"token": tokenString})
	}
}
