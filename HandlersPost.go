package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/OminousOmelet/chirpy/internal/auth"
	"github.com/OminousOmelet/chirpy/internal/database"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerPostChirp(w http.ResponseWriter, r *http.Request) {
	type chirpParams struct {
		Body string `json:"body"`
	}

	decoder := json.NewDecoder(r.Body)
	params := chirpParams{}
	err := decoder.Decode(&params)
	if err != nil {
		w.WriteHeader(500)
		log.Fatalf("Error decoding parameters: %s", err)
	}
	if len(params.Body) > 140 {
		respondWithError(w, 400, "Chirp is too long")
		return
	}

	tokenStr, err := auth.GetBearerToken(r.Header)
	if err != nil {
		log.Fatal(err)
	}
	userID, err := auth.ValidateJWT(tokenStr, cfg.secret)
	if err != nil {
		respondWithError(w, 401, "JWT validation failed")
		log.Printf("Chirp-post validation failed: %s", err)
		return
	}

	chirp, err := cfg.dbQueries.PostChirp(context.Background(), database.PostChirpParams{Body: params.Body, UserID: userID})
	if err != nil {
		log.Printf("Error posting chirp: %s", err)
		return
	}

	//Write JSON response after posting chirp to database
	cleanChirp := prepChirp(chirp)
	respondWithJSON(w, 201, cleanChirp)
}

func (cfg *apiConfig) handlerDeleteChirp(w http.ResponseWriter, r *http.Request) {
	headerToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, 401, "DELETE FAILED")
		log.Printf("ERROR DELETING: failed to get token from header:, %s", err)
		return
	}

	userID, err := auth.ValidateJWT(headerToken, cfg.secret)
	if err != nil {
		respondWithError(w, 401, "DELETE FAILED")
		log.Printf("ERROR DELETING CHIRP: JWT validation failed: %s", err)
		return
	}

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
			log.Print("ERROR DELETING CHIRP: chirp not found")
		} else {
			w.WriteHeader(500)
			log.Printf("ERROR DELETING CHIRP: Error getting chirp by ID: %s", err)
		}
		return
	}

	if userID != chirp.UserID {
		respondWithError(w, 403, "UNAUTHORIZED")
		log.Print("Chirp deletion denied, user ID mismatch")
		return
	}

	err = cfg.dbQueries.DeleteUsers(context.Background())
	if err != nil {
		w.WriteHeader(500)
		log.Printf("Failed to delete chirp: %s", err)
		return
	}

	w.WriteHeader(204)
}
