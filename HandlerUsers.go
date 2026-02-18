package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/OminousOmelet/chirpy/internal/auth"
	"github.com/OminousOmelet/chirpy/internal/database"
)

type UserCreds struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Create user
func (cfg *apiConfig) handlerUser(w http.ResponseWriter, r *http.Request) {
	var u UserCreds
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&u)
	if err != nil {
		w.WriteHeader(500)
		log.Printf("Error decoding parameters: %s", err)
		return
	}

	hashPass, err := auth.HashPassword(u.Password)
	if err != nil {
		log.Printf("Failed to hash password: %s", err)
		return
	}
	user, err := cfg.dbQueries.CreateUser(context.Background(), database.CreateUserParams{Email: u.Email, HashedPassword: hashPass})
	if err != nil {
		log.Printf("Error creating user: %s", err)
		return
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
	var creds UserCreds
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&creds)
	if err != nil {
		w.WriteHeader(500)
		log.Printf("Error decoding parameters: %s", err)
	}

	user, err := cfg.dbQueries.GetUserByEmail(context.Background(), creds.Email)
	if err != nil {
		if strings.HasPrefix(err.Error(), "sql: no rows") {
			respondWithError(w, 401, "Unauthorized (user doesn't exist)")
		} else {
			log.Printf("Failed to get user: %s", err)
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
		return
	} else {
		secureUser = User{
			ID: user.ID, CreatedAt: user.CreatedAt, UpdatedAt: user.UpdatedAt, Email: user.Email,
		}
	}

	tokenStr, err := auth.MakeJWT(secureUser.ID, cfg.secret)
	if err != nil {
		log.Printf("Error making JWT: %s", err)
		return
	}
	secureUser.Token = tokenStr

	refreshStr, err := auth.MakeRefreshToken()
	if err != nil {
		respondWithError(w, 401, "")
		log.Printf("Failed to make refresh token: %s", err)
		return
	}
	secureUser.RefreshToken = refreshStr

	refreshParams := database.StoreRefreshTokenParams{
		Token: refreshStr, UserID: secureUser.ID,
	}
	_, err = cfg.dbQueries.StoreRefreshToken(context.Background(), refreshParams)
	if err != nil {
		respondWithError(w, 401, "")
		log.Printf("Failed to store refresh token %s", err)
		return
	}

	respondWithJSON(w, 200, secureUser)
}

func (cfg *apiConfig) handlerUpdateUser(w http.ResponseWriter, r *http.Request) {
	headerToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, 401, "UPDATE FAILED")
		log.Printf("ERROR UPDATING USER: failed to get token from header:, %s", err)
		return
	}

	userID, err := auth.ValidateJWT(headerToken, cfg.secret)
	if err != nil {
		respondWithError(w, 401, "UPDATE FAILED")
		log.Printf("ERROR UPDATING USER: JWT validation failed: %s", err)
		return
	}

	var creds UserCreds
	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(&creds)
	if err != nil {
		respondWithError(w, 401, "UPDATE FAILED")
		log.Printf("ERROR UPDATING USER: json decoding error: %s", err)
		return
	}

	hashPass, err := auth.HashPassword(creds.Password)
	if err != nil {
		log.Printf("ERROR UPDATING USER: Failed to hash password: %s", err)
		return
	}
	params := database.UpateUserCredsParams{
		Email: creds.Email, HashedPassword: hashPass, ID: userID,
	}
	user, err := cfg.dbQueries.UpateUserCreds(context.Background(), params)
	if err != nil {
		respondWithError(w, 401, "UPDATED FAILED")
		log.Printf("ERROR UPDATING USER: Failed to store credentials: %s", err)
		return
	}
	secureUser := User{
		ID: user.ID, CreatedAt: user.CreatedAt, UpdatedAt: user.UpdatedAt, Email: user.Email,
	}
	respondWithJSON(w, 200, secureUser)
}
