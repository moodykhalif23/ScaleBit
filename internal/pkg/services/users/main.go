package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"encoding/json"
	"strconv"

	"golang.org/x/crypto/bcrypt"

	_ "github.com/lib/pq"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/trace"

	"github.com/golang-jwt/jwt/v5"

	"github.com/gorilla/mux"
	"github.com/moodykhalif23/scalebit/internal/pkg/security"
	"github.com/moodykhalif23/scalebit/internal/pkg/telemetry"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type User struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Role  string `json:"role"`
}

type RegisterRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Role     string `json:"role,omitempty"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func main() {
	tp := initTracer()
	defer tp.Shutdown(context.Background())

	// Database setup
	db := setupDB()
	defer db.Close()

	// Prometheus metrics
	meter := otel.GetMeterProvider().Meter("user-service")
	telemetry.InitMetrics(meter)

	publicRouter := mux.NewRouter()

	// Health endpoint
	publicRouter.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	}).Methods("GET", "OPTIONS")

	// Prometheus metrics endpoint
	publicRouter.Handle("/metrics", promhttp.Handler())

	// Register public auth routes
	publicRouter.HandleFunc("/register", registerHandler(db)).Methods("POST", "OPTIONS")
	publicRouter.HandleFunc("/login", loginHandler(db)).Methods("POST", "OPTIONS")
	publicRouter.HandleFunc("/debug-jwt", debugJWTHandler()).Methods("GET", "OPTIONS")

	userRouter := publicRouter.PathPrefix("/users").Subrouter()
	userRouter.HandleFunc("", getUsers(db)).Methods("GET")
	userRouter.HandleFunc("", createUser(db)).Methods("POST")
	userRouter.HandleFunc("/{id:[0-9]+}", getUser(db)).Methods("GET")
	userRouter.HandleFunc("/{id:[0-9]+}", updateUser(db)).Methods("PUT")
	userRouter.HandleFunc("/{id:[0-9]+}", deleteUser(db)).Methods("DELETE")

	// CORS middleware to all routes
	handler := telemetry.Middleware(corsMiddleware(publicRouter))

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

	err = db.Ping()
	if err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	return db
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// CRUD Handlers
func getUsers(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rows, err := db.Query("SELECT id, name, email, role FROM users")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()
		users := []User{}
		for rows.Next() {
			var u User
			if err := rows.Scan(&u.ID, &u.Name, &u.Email, &u.Role); err != nil {
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
		role := req.Role
		if role == "" {
			role = "user"
		}
		hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			http.Error(w, "Failed to hash password", http.StatusInternalServerError)
			return
		}
		var id int
		err = db.QueryRow("INSERT INTO users (name, email, password_hash, role) VALUES ($1, $2, $3, $4) RETURNING id", req.Name, req.Email, string(hash), role).Scan(&id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]interface{}{"id": id, "name": req.Name, "email": req.Email, "role": role})
	}
}

func getUser(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id := vars["id"]
		var u User
		err := db.QueryRow("SELECT id, name, email, role FROM users WHERE id = $1", id).Scan(&u.ID, &u.Name, &u.Email, &u.Role)
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
		_, err := db.Exec("UPDATE users SET name = $1, email = $2, role = $3 WHERE id = $4", u.Name, u.Email, u.Role, id)
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
		_, err := db.Exec("DELETE FROM users WHERE id = $1", id)
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
		role := req.Role
		if role == "" {
			role = "user"
		}
		hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			http.Error(w, "Failed to hash password", http.StatusInternalServerError)
			return
		}
		var id int
		err = db.QueryRow("INSERT INTO users (name, email, password_hash, role) VALUES ($1, $2, $3, $4) RETURNING id", req.Name, req.Email, string(hash), role).Scan(&id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]interface{}{"id": id, "name": req.Name, "email": req.Email, "role": role})
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
		log.Printf("Login attempt for email: %s", req.Email)
		var id int
		var name, email, passwordHash, role string
		err := db.QueryRow("SELECT id, name, email, password_hash, role FROM users WHERE email = $1", req.Email).Scan(&id, &name, &email, &passwordHash, &role)
		if err == sql.ErrNoRows {
			log.Printf("No user found with email: %s", req.Email)
			http.Error(w, "Invalid credentials", http.StatusUnauthorized)
			return
		} else if err != nil {
			log.Printf("Database error during login: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		log.Printf("Found user: %s (ID: %d)", email, id)
		log.Printf("Comparing password hash for user: %s", email)
		if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(req.Password)); err != nil {
			log.Printf("Password comparison failed: %v", err)
			http.Error(w, "Invalid credentials", http.StatusUnauthorized)
			return
		}
		// Create JWT
		jwtSecret := os.Getenv("JWT_SECRET")
		if jwtSecret == "" {
			http.Error(w, "JWT secret not set", http.StatusInternalServerError)
			return
		}
		log.Printf("Using JWT secret: %s", jwtSecret)
		now := time.Now()
		claims := map[string]interface{}{
			"id":    id,
			"name":  name,
			"email": email,
			"role":  role,
			"iat":   now.Unix(),
			"exp":   now.Add(24 * time.Hour).Unix(),
			"iss":   "scalebit-platform",
			"aud":   "scalebit-api",
		}
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims(claims))
		// Set the kid in the header for KrakenD validation
		token.Header["kid"] = "scalebit-key-1"
		log.Printf("Generated JWT claims: %+v", claims)
		tokenString, err := token.SignedString([]byte(jwtSecret))
		if err != nil {
			http.Error(w, "Failed to sign token", http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(map[string]string{"token": tokenString, "role": role})
	}
}

// debugJWTHandler returns the JWT claims from the Authorization header for debugging purposes.
func debugJWTHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Authorization header missing", http.StatusBadRequest)
			return
		}
		var tokenString string
		if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
			tokenString = authHeader[7:]
		} else {
			tokenString = authHeader
		}
		jwtSecret := os.Getenv("JWT_SECRET")
		if jwtSecret == "" {
			http.Error(w, "JWT secret not set", http.StatusInternalServerError)
			return
		}
		token, _ := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return []byte(jwtSecret), nil
		})
		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(claims)
			return
		}
		http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
	}
}
