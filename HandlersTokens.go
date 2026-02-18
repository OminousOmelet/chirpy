package main

import (
	"context"
	"log"
	"net/http"

	"github.com/OminousOmelet/chirpy/internal/auth"
)

func (cfg *apiConfig) handlerRefreshToken(w http.ResponseWriter, r *http.Request) {
	headerToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, 401, "")
		log.Print(err)
		return
	}
	token, err := cfg.dbQueries.GetToken(context.Background(), headerToken)
	if err != nil {
		respondWithError(w, 401, err.Error())
		log.Printf("Failed to get token: %s", err)
		return
	}
	if token.RevokedAt.Valid {
		respondWithError(w, 401, "TOKEN EXPIRED")
		log.Print("Failed to refresh token (current token already revoked)")
		return
	}

	userID, err := cfg.dbQueries.GetUserFromRefreshToken(context.Background(), token.Token)
	if err != nil {
		respondWithError(w, 401, "")
		log.Printf("Failed to get user from token: %s", err)
		return
	}

	// Refresh token verified, now make new JWT and pass to JSON response
	newTokenStr, err := auth.MakeJWT(userID, cfg.secret)
	if err != nil {
		respondWithError(w, 401, "")
		log.Printf("Error making new (refreshed) JWT: %s", err)
		return
	}

	type JustTheToken struct {
		Token string `json:"token"`
	}
	respondWithJSON(w, 200, JustTheToken{Token: newTokenStr})
}

func (cfg *apiConfig) handlerRevokeToken(w http.ResponseWriter, r *http.Request) {
	headerToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, 401, "")
		log.Printf("REVOKE TOKEN ERROR: auth failed: %s", err)
		return
	}
	token, err := cfg.dbQueries.GetToken(context.Background(), headerToken)
	if err != nil {
		respondWithError(w, 401, err.Error())
		log.Printf("REVOKE TOKEN ERROR: failed to get token: %s", err)
		return
	}

	err = cfg.dbQueries.RevokeToken(context.Background(), token.Token)
	if err != nil {
		respondWithError(w, 401, err.Error())
		log.Printf("Failed to revoke token: %s", err)
		return
	}
	respondWithJSON(w, 204, nil)
}
