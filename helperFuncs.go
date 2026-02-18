package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/OminousOmelet/chirpy/internal/database"
)

func respondWithError(w http.ResponseWriter, code int, msg string) {
	type jsonError struct {
		Error string `json:"error"`
	}
	errBody := jsonError{Error: msg}
	dat, err := json.Marshal(errBody)
	if err != nil {
		log.Printf("ERROR MESSAGE FAILED: %s", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(dat)
}

func respondWithJSON(w http.ResponseWriter, code int, payload any) {
	json, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Failed to marshal payload: %s", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(json)
}

func prepChirp(chirp database.Chirp) Chirp {
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

	// convert database.chirp to chirp (for json parsing)
	newChirp := Chirp{
		ID: chirp.ID, CreatedAt: chirp.CreatedAt, UpdatedAt: chirp.CreatedAt, Body: chirp.Body, UserID: chirp.UserID,
	}
	return newChirp
}
