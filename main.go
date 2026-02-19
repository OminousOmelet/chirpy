package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync/atomic"
	"time"

	"github.com/OminousOmelet/chirpy/internal/database"
	"github.com/joho/godotenv"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	dbQueries      *database.Queries
	platform       string
	secret         string
	polkaKey       string
}

type User struct {
	ID           uuid.UUID `json:"id"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	Email        string    `json:"email"`
	Token        string    `json:"token"`
	RefreshToken string    `json:"refresh_token"`
	IsChirpyRed  bool      `json:"is_chirpy_red"`
}

type Chirp struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string    `json:"body"`
	UserID    uuid.UUID `json:"user_id"`
}

func main() {
	// removes timestamps from log messages
	log.SetFlags(0)

	godotenv.Load()
	dbURL := os.Getenv("DB_URL")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal(err)
	}

	const filepathRoot = "."
	const port = "8080"
	apiCfg := apiConfig{}
	apiCfg.dbQueries = database.New(db)
	apiCfg.platform = os.Getenv("PLATFORM")
	apiCfg.secret = os.Getenv("SECRET")
	apiCfg.polkaKey = os.Getenv("POLKA_KEY")

	mux := http.NewServeMux()
	mux.Handle("/app/", apiCfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(filepathRoot)))))
	mux.HandleFunc("GET /api/healthz", handlerReadiness)
	mux.HandleFunc("GET /admin/metrics", apiCfg.handlerMetrics)
	mux.HandleFunc("GET /api/chirps", apiCfg.handlerGetChirps)
	mux.HandleFunc("GET /api/chirps/{chirpID}", apiCfg.handlerChirpByID)
	mux.HandleFunc("POST /admin/reset", apiCfg.handlerReset)
	mux.HandleFunc("POST /api/users", apiCfg.handlerUser)
	mux.HandleFunc("POST /api/chirps", apiCfg.handlerPostChirp)
	mux.HandleFunc("POST /api/login", apiCfg.handlerLogin)
	mux.HandleFunc("POST /api/refresh", apiCfg.handlerRefreshToken)
	mux.HandleFunc("POST /api/revoke", apiCfg.handlerRevokeToken)
	mux.HandleFunc("PUT /api/users", apiCfg.handlerUpdateUser)
	mux.HandleFunc("DELETE /api/chirps/{chirpID}", apiCfg.handlerDeleteChirp)
	mux.HandleFunc("POST /api/polka/webhooks", apiCfg.HandlerUpgradeUser)

	svr := &http.Server{Handler: mux, Addr: ":" + port}
	log.Printf("Serving files from %s on port: %s\n", filepathRoot, port)
	log.Fatal(svr.ListenAndServe())
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func handlerReadiness(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(http.StatusText(http.StatusOK)))
}

func (cfg *apiConfig) handlerMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(fmt.Appendf(nil, `
		<html>
			<body>
				<h1>Welcome, Chirpy Admin</h1>
				<p>Chirpy has been visited %d times!</p>
			</body>
		</html>
			`, cfg.fileserverHits.Load()))
}

func (cfg *apiConfig) handlerReset(w http.ResponseWriter, r *http.Request) {
	cfg.fileserverHits.Store(0)
	w.Header().Set("Content-Type", "application/json")

	if cfg.platform != "dev" {
		w.WriteHeader(403)
		w.Write([]byte("HTTP 403 FORBIDDEN"))
		return
	}

	cfg.dbQueries.DeleteUsers(context.Background())
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("All users deleted"))
}
