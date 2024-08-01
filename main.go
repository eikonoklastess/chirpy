package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"eikono.chirpy/internal/database"
	"github.com/eikonoklastess/chirpy/internal/database"
)

const (
	port         = "8080"
	filePathRoot = "."
)

func main() {
	cfg := apiConfig{
		fileserverHits: 0,
	}
	mux := http.NewServeMux()
	mux.Handle("/app/*", cfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(filePathRoot)))))
	mux.HandleFunc("GET /api/healthz", handleReadiness)
	mux.HandleFunc("GET /admin/metrics", cfg.handleAppHits)
	mux.HandleFunc("/api/reset", cfg.resetAppHits)
	mux.HandleFunc("/api/chirps", crudChirps)

	newServer := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	log.Printf("Serving file from %s on port %s", filePathRoot, port)
	log.Fatal(newServer.ListenAndServe())
}

func handleReadiness(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(http.StatusText(http.StatusOK)))
}

type apiConfig struct {
	fileserverHits int
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits++
		next.ServeHTTP(w, r)
	})
}

func (cfg *apiConfig) handleAppHits(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf("<html><body><h1>Welcome, Chirpy Admin</h1><p>Chirpy has been visited %d times!</p></body></html>", cfg.fileserverHits)))
}

func (cfg *apiConfig) resetAppHits(w http.ResponseWriter, r *http.Request) {
	cfg.fileserverHits = 0
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("reset Hits"))
}

type Chirp struct {
	id   int
	Body string `json:"body"`
}

func handleChirpValidation(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST methode is allowed", http.StatusMethodNotAllowed)
		return
	}

	type validResp struct {
		Valid bool `json:"valid"`
	}

	type errorResp struct {
		Error string `json:"error"`
	}

	decoder := json.NewDecoder(r.Body)
	chirpReq := chirp{}
	err := decoder.Decode(&chirpReq)
	if err != nil {
		http.Error(w, fmt.Sprintf("Decode() error: %v", err), http.StatusInternalServerError)
		return
	}

	if len(chirpReq.Body) > 140 {
		resp := errorResp{Error: "Chrip is too long"}
		respondWithJson(w, http.StatusBadRequest, &resp)
		return
	}

	cleanedBody := cleanChirp(chirpReq)
	respondWithJson(w, http.StatusOK, &cleanedBody)
	return
}

func respondWithJson(w http.ResponseWriter, code int, payload interface{}) {
	dat, err := json.Marshal(payload)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(code)
	w.Write(dat)
}

func cleanChirp(chirpReq chirp) struct {
	Cleaned_body string `json:"cleaned_body"`
} {
	words := strings.Split(chirpReq.Body, " ")
	newWords := []string{}
	for _, word := range words {
		w := strings.ToLower(word)
		if w == "kerfuffle" || w == "sharbert" || w == "fornax" {
			newWords = append(newWords, "****")
		} else {
			newWords = append(newWords, word)
		}
	}

	clean := struct {
		Cleaned_body string `json:"cleaned_body"`
	}{
		Cleaned_body: strings.Join(newWords, " "),
	}
	return clean
}

func crudChirps(w http.http.ResponseWriter, r *http.Request) {
	r.Method == http.MethodGet {
		var db = database.DB{}
		db := database.NewDB("/Users/walidoutaleb/workspace/github.com/eikonoklastes/chirpy/database.json")
		r
	}
	
}









