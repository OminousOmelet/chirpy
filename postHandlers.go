package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strings"
)

func (cfg *apiConfig) handlerPostChirp(w http.ResponseWriter, r *http.Request) {
	type bodyParams struct {
		Body string `json:"body"`
	}

	decoder := json.NewDecoder(r.Body)
	params := bodyParams{}
	err := decoder.Decode(&params)
	if err != nil {
		log.Printf("Error decoding parameters: %s", err)
		w.WriteHeader(500)
	}
	if len(params.Body) > 140 {
		respondWithError(w, 400, "Chirp is too long")
	} else {
		respondWithJSON(w, 200, params.Body)
	}
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

func respondWithJSON(w http.ResponseWriter, code int, payload string) {
	type cleanJSON struct {
		CleanBody string `json:"cleaned_body"`
	}

	dat, err := json.Marshal(cleanJSON{CleanBody: payload})
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		return
	}
	datStr := string(dat)
	datWords := strings.Split(datStr, " ")

	for i := range datWords {
		switch strings.ToLower(datWords[i]) {
		case "kerfuffle":
			fallthrough
		case "sharbert":
			fallthrough
		case "fornax":
			datWords[i] = "****"
		}
	}

	datStr = strings.Join(datWords, " ")

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write([]byte(datStr))
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
		log.Fatal(err)
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
