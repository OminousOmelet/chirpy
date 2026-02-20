package main

import (
	"context"
	"log"
	"net/http"
	"sort"
	"strings"

	"github.com/OminousOmelet/chirpy/internal/database"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerGetChirps(w http.ResponseWriter, r *http.Request) {
	authorID := r.URL.Query().Get("author_id")
	var chirps []database.Chirp

	if authorID != "" {
		parsedID, err := uuid.Parse(authorID)
		if err != nil {
			log.Printf("ERROR GETTING CHIRPS: failed to parse author id: %s", err)
			return
		}
		chirps, err = cfg.dbQueries.GetChirpsByUserID(context.Background(), parsedID)
		if err != nil {
			log.Printf("ERROR GETTING CHIRPS: failed id lookup: %s", err)
			return
		}
	} else {
		// "err" needed declaration so out-of-scope "chirps" would be used
		var err error
		chirps, err = cfg.dbQueries.GetChirps(context.Background())
		if err != nil {
			log.Printf("Error getting chirps: %s", err)
			return
		}
	}

	var chirpList []Chirp
	for _, chirp := range chirps {
		chirpList = append(chirpList, prepChirp(chirp))
	}

	// Ugly-ass sorting code (with ascending or descending decision)
	sortDesc := r.URL.Query().Get("sort") == "desc"
	sort.Slice(chirpList, func(i, j int) bool {
		sortDir := chirpList[i].CreatedAt.Before(chirpList[j].CreatedAt)
		if sortDesc {
			sortDir = !sortDir
		}
		return sortDir
	})
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
