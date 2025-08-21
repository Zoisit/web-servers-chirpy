package main

import (
	"fmt"
	"net/http"
	"sync/atomic"
)

type apiConfig struct {
	fileserverHits atomic.Int32
}

func main() {
	apiCfg := apiConfig{}
	serveMux := http.NewServeMux()
	appHandler := http.StripPrefix("/app", http.FileServer(http.Dir(".")))
	serveMux.Handle("/app/", apiCfg.middlewareMetricsInc(appHandler))
	serveMux.Handle("/", apiCfg.middlewareMetricsInc(http.HandlerFunc(handlerReadyCheck)))
	serveMux.Handle("/metrics", http.HandlerFunc(apiCfg.handlerGetApiHits))
	serveMux.Handle("/reset", http.HandlerFunc(apiCfg.handlerResetApiHits))

	server := &http.Server{
		Addr:    ":8080",
		Handler: serveMux,
	}

	server.ListenAndServe()
}

func handlerReadyCheck(rw http.ResponseWriter, req *http.Request) {
	rw.Header().Add("Content-Type", "text/plain; charset=utf-8")
	rw.WriteHeader(200)
	rw.Write([]byte("OK"))
}

func (cfg *apiConfig) handlerGetApiHits(rw http.ResponseWriter, req *http.Request) {
	rw.Header().Add("Content-Type", "text/plain; charset=utf-8")
	rw.WriteHeader(200)
	rw.Write([]byte(fmt.Sprintf("Hits: %v", cfg.fileserverHits.Load())))
}

func (cfg *apiConfig) handlerResetApiHits(rw http.ResponseWriter, req *http.Request) {
	rw.Header().Add("Content-Type", "text/plain; charset=utf-8")
	rw.WriteHeader(200)
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
