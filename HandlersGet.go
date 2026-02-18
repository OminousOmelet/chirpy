package main

import (
	"context"
	"log"
	"net/http"
	"strings"

	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerGetChirps(w http.ResponseWriter, r *http.Request) {
	chirps, err := cfg.dbQueries.GetChirps(context.Background())
	if err != nil {
		log.Printf("Error getting chirps: %s", err)
		return
	}

	var chirpList []Chirp
	for _, chirp := range chirps {
		chirpList = append(chirpList, prepChirp(chirp))
	}

	respondWithJSON(w, 200, chirpList)
}

func (cfg *apiConfig) handlerChirpByID(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("chirpID")
	chirpID, err := uuid.Parse(id)
	if err != nil {
		log.Printf("Failed to parse ID %s", err)
		return
	}

	chirp, err := cfg.dbQueries.GetChirpByID(context.Background(), chirpID)
	if err != nil {
		if strings.HasPrefix(err.Error(), "sql: no rows") {
			respondWithError(w, 404, "Chirp not found")
		} else {
			w.WriteHeader(500)
			log.Printf("Error getting chirp by ID: %s", err)
		}
		return
	}

	// chirp must be cleaned and prepped for JSON
	cleanChirp := prepChirp(chirp)
	respondWithJSON(w, 200, cleanChirp)
}
