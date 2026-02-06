package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/OminousOmelet/chirpy/internal/database"
	"github.com/google/uuid"
)

type User struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
}

type Chirp struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string    `json:"body"`
	UserID    uuid.UUID `json:"user_id"`
}

func (cfg *apiConfig) handlerPostChirp(w http.ResponseWriter, r *http.Request) {
	type chirpParams struct {
		Body   string    `json:"body"`
		UserID uuid.UUID `json:"user_id"`
	}

	decoder := json.NewDecoder(r.Body)
	params := chirpParams{}
	err := decoder.Decode(&params)
	if err != nil {
		log.Fatalf("Error decoding parameters: %s", err)
		w.WriteHeader(500)
	}
	if len(params.Body) > 140 {
		respondWithError(w, 400, "Chirp is too long")
		return
	}

	chirp, err := cfg.dbQueries.PostChirp(context.Background(), database.PostChirpParams{Body: params.Body, UserID: params.UserID})
	if err != nil {
		log.Fatalf("Error posting chirp: %s", err)
	}

	//Write JSON response after posting chirp to database
	newChirp := Chirp{
		ID: chirp.ID, CreatedAt: chirp.CreatedAt, UpdatedAt: chirp.CreatedAt, Body: chirp.Body, UserID: chirp.UserID,
	}
	respondWithJSON(w, 201, newChirp)

}

func respondWithError(w http.ResponseWriter, code int, msg string) {
	type jsonError struct {
		Error string `json:"error"`
	}
	errBody := jsonError{Error: msg}
	dat, err := json.Marshal(errBody)
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(dat)
}

func respondWithJSON(w http.ResponseWriter, code int, chirp Chirp) {
	// "clean" the response body (filtered words)
	bodyStr := chirp.Body
	words := strings.Split(bodyStr, " ")
	for i := range words {
		switch strings.ToLower(words[i]) {
		case "kerfuffle":
			fallthrough
		case "sharbert":
			fallthrough
		case "fornax":
			words[i] = "****"
		}
	}
	chirp.Body = strings.Join(words, " ")

	dat, err := json.Marshal(chirp)
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write([]byte(dat))
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
