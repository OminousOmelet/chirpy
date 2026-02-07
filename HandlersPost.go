package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/OminousOmelet/chirpy/internal/database"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerPostChirp(w http.ResponseWriter, r *http.Request) {
	type chirpParams struct {
		Body   string    `json:"body"`
		UserID uuid.UUID `json:"user_id"`
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
	}

	chirp, err := cfg.dbQueries.PostChirp(context.Background(), database.PostChirpParams{Body: params.Body, UserID: params.UserID})
	if err != nil {
		log.Fatalf("Error posting chirp: %s", err)
	}

	//Write JSON response after posting chirp to database
	cleanChirp := prepChirp(chirp)
	respondWithJSON(w, 201, cleanChirp)
}

// Create user
func (cfg *apiConfig) handlerUser(w http.ResponseWriter, r *http.Request) {
	type email struct{ Email string }
	var userEmail email
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&userEmail)
	if err != nil {
		log.Printf("Error decoding parameters: %s", err)
		w.WriteHeader(500)
	}

	user, err := cfg.dbQueries.CreateUser(context.Background(), userEmail.Email)
	if err != nil {
		log.Fatalf("Error creating user: %s", err)
	}

	newUser := User{
		ID: user.ID, CreatedAt: user.CreatedAt, UpdatedAt: user.UpdatedAt, Email: user.Email,
	}

	dat, err := json.Marshal(newUser)
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(201)
	w.Write([]byte(dat))
}
