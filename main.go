package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
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

func handleChirpValidation(c Chirp) (Chirp, error) {
	if len(c.Body) > 140 {
		return Chirp{}, errors.New("Chirp too long")
	}

	return cleanChirp(c)
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

func crudChirps(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		db, err := database.NewDB("/Users/walidoutaleb/workspace/github.com/eikonoklastes/chirpy/database.json")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		chirps, err := db.GetChirps()
		if err != nil {
			http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
		}
		respondWithJson(w, http.StatusOK, &chirps)
	} else if r.Method == http.MethodPost {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		defer r.Body.Close()
		chirpBody := Chirp{}
		erro := json.Unmarshal(body, &chirpBody)
		if erro != nil {
			http.Error(w, erro.Error(), http.StatusInternalServerError)
		}
		validChirp, err := handleChirpValidation(chirpBody)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		chirp, err := database.CreateChirp(validChirp)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		respondWithJson(w, http.StatusCreated, &chirp)
	}
}
