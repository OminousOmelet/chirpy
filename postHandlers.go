package main

import (
	"encoding/json"
	"log"
	"net/http"
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
		respondWithJSON(w, 200, params)
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

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	_, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write([]byte(`{"valid":true}`))
}
