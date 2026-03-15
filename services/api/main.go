package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	gen "clan-budget/services/api/gen"
	"clan-budget/services/api/internal/handlers"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func main() {
	// Load .env from possible locations
	for _, p := range []string{".env", "../../.env"} {
		_ = godotenv.Load(p)
	}

	// Connect to Postgres
	db, err := sql.Open("postgres", dbConnStr())
	if err != nil {
		log.Fatalf("Failed to open DB: %v", err)
	}
	defer db.Close()
	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping DB: %v", err)
	}
	log.Println("Connected to Postgres")

	// Build router
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// CORS for local frontend (Next.js on :3000)
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			next.ServeHTTP(w, r)
		})
	})

	server := handlers.NewServer(db)
	gen.HandlerWithOptions(server, gen.ChiServerOptions{
		BaseURL:    "/api/v1",
		BaseRouter: r,
	})

	addr := ":8080"
	log.Printf("API server listening on %s", addr)
	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

func dbConnStr() string {
	must := func(key string) string {
		v := os.Getenv(key)
		if v == "" {
			log.Fatalf("Missing required env var: %s", key)
		}
		return v
	}
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		must("POSTGRES_HOST"),
		must("POSTGRES_PORT"),
		must("POSTGRES_USER"),
		must("POSTGRES_PASSWORD"),
		must("POSTGRES_DB"),
	)
}
