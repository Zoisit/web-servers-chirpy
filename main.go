package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"sync/atomic"

	"github.com/joho/godotenv"

	"github.com/Zoisit/web-servers-chirpy/internal/database"
	_ "github.com/lib/pq"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	db             *database.Queries
	platform       string
}

func main() {
	godotenv.Load()
	dbURL := os.Getenv("DB_URL")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		println(err)
		return
	}
	dbQueries := database.New(db)

	apiCfg := apiConfig{
		fileserverHits: atomic.Int32{},
		db:             dbQueries,
		platform:       os.Getenv("PLATFORM"),
	}

	serveMux := http.NewServeMux()

	appHandler := http.StripPrefix("/app", http.FileServer(http.Dir(".")))
	serveMux.Handle("/app/", apiCfg.middlewareMetricsInc(appHandler))

	serveMux.HandleFunc("POST /api/chirps", apiCfg.handlerCreateChirp)
	serveMux.HandleFunc("GET /api/chirps", apiCfg.handlerGetAllChirps)
	serveMux.HandleFunc("GET /api/chirps/{chirpID}", apiCfg.handlerGetSingleChirp)

	serveMux.HandleFunc("POST /api/users", apiCfg.handlerCreateUser)

	serveMux.HandleFunc("GET /api/healthz", handlerReadyCheck)
	serveMux.HandleFunc("GET /admin/metrics", apiCfg.handlerGetApiHits)
	serveMux.HandleFunc("POST /admin/reset", apiCfg.handlerResetApiHits)

	server := &http.Server{
		Addr:    ":8080",
		Handler: serveMux,
	}

	server.ListenAndServe()
}

func handlerReadyCheck(rw http.ResponseWriter, req *http.Request) {
	rw.Header().Add("Content-Type", "text/plain; charset=utf-8")
	rw.WriteHeader(http.StatusOK)
	rw.Write([]byte(http.StatusText(http.StatusOK)))
}

func (cfg *apiConfig) handlerGetApiHits(rw http.ResponseWriter, req *http.Request) {
	rw.Header().Add("Content-Type", "text/html; charset=utf-8")
	rw.WriteHeader(http.StatusOK)
	rw.Write([]byte(fmt.Sprintf(`<html>
  <body>
    <h1>Welcome, Chirpy Admin</h1>
    <p>Chirpy has been visited %d times!</p>
  </body>
</html>`, cfg.fileserverHits.Load())))
}

func (cfg *apiConfig) handlerResetApiHits(rw http.ResponseWriter, req *http.Request) {
	rw.Header().Add("Content-Type", "text/plain; charset=utf-8")
	if cfg.platform != "dev" {
		rw.WriteHeader(http.StatusForbidden)
		rw.Write([]byte("Reset is only allowed in dev environment."))
		return
	}

	err := cfg.db.ResetUsers(req.Context())
	if err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
		rw.Write([]byte("Failed to reset user database: " + err.Error()))
		return
	}

	rw.WriteHeader(http.StatusOK)
	cfg.fileserverHits.Store(0)
	rw.Write([]byte(fmt.Sprintf("Hits: %v", cfg.fileserverHits.Load())))
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	f := func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
		cfg.fileserverHits.Add(1)
	}

	return http.HandlerFunc(f)
}
