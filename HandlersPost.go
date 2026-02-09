package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/OminousOmelet/chirpy/internal/auth"
	"github.com/OminousOmelet/chirpy/internal/database"
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
	}

	tokenStr, err := auth.GetBearerToken(r.Header)
	if err != nil {
		log.Fatal(err)
	}
	userID, err := auth.ValidateJWT(tokenStr, cfg.secret)
	if err != nil {
		log.Fatal(err)
	}

	chirp, err := cfg.dbQueries.PostChirp(context.Background(), database.PostChirpParams{Body: params.Body, UserID: userID})
	if err != nil {
		log.Fatalf("Error posting chirp: %s", err)
	}

	//Write JSON response after posting chirp to database
	cleanChirp := prepChirp(chirp)
	respondWithJSON(w, 201, cleanChirp)
}

// Create user
func (cfg *apiConfig) handlerUser(w http.ResponseWriter, r *http.Request) {
	type UserParams struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	var u UserParams
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&u)
	if err != nil {
		w.WriteHeader(500)
		log.Printf("Error decoding parameters: %s", err)
	}

	hashPass, err := auth.HashPassword(u.Password)
	if err != nil {
		log.Fatalf("Failed to hash password: %s", err)
	}
	user, err := cfg.dbQueries.CreateUser(context.Background(), database.CreateUserParams{Email: u.Email, HashedPassword: hashPass})
	if err != nil {
		log.Fatalf("Error creating user: %s", err)
	}

	// just for display
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

// login user with password
func (cfg *apiConfig) handlerLogin(w http.ResponseWriter, r *http.Request) {
	type UserCreds struct {
		Email         string        `json:"email"`
		Password      string        `json:"password"`
		ExpiresInSecs time.Duration `json:"expires_in_seconds"`
	}

	var creds UserCreds
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&creds)
	if err != nil {
		w.WriteHeader(500)
		log.Printf("Error decoding parameters: %s", err)
	}

	// Convert duration value to seconds (or default to 1 hour)
	if creds.ExpiresInSecs == 0 {
		creds.ExpiresInSecs = time.Hour
	} else {
		creds.ExpiresInSecs *= time.Second
	}

	user, err := cfg.dbQueries.GetUserByEmail(context.Background(), creds.Email)
	if err != nil {
		if strings.HasPrefix(err.Error(), "sql: no rows") {
			respondWithError(w, 401, "Unauthorized (user doesn't exist)")
		} else {
			log.Fatalf("Failed to get user: %s", err)
		}
		return
	}

	authorized, err := auth.CheckPasswordHash(creds.Password, user.HashedPassword)
	if err != nil {
		log.Fatalf("Error checking password: %s", err)
	}

	var secureUser User
	if !authorized {
		respondWithError(w, 401, "Unauthorized")
	} else {
		secureUser = User{
			ID: user.ID, CreatedAt: user.CreatedAt, UpdatedAt: user.UpdatedAt, Email: user.Email,
		}
	}

	tokenStr, err := auth.MakeJWT(secureUser.ID, cfg.secret, creds.ExpiresInSecs)
	if err != nil {
		log.Fatalf("Failed to get token: %s", err)
	}
	secureUser.Token = tokenStr
	respondWithJSON(w, 200, secureUser)
}
