package auth

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func MakeJWT(userID uuid.UUID, tokenSecret string) (string, error) {
	claims := jwt.RegisteredClaims{
		Issuer:    "chirpy",
		IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
		ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(time.Hour)),
		Subject:   uuid.UUID.String(userID),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString([]byte(tokenSecret))
	if err != nil {
		return "", err
	}
	return tokenStr, nil
}

func ValidateJWT(tokenString, tokenSecret string) (uuid.UUID, error) {
	token, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(*jwt.Token) (any, error) {
		return []byte(tokenSecret), nil
	})
	if err != nil {
		return uuid.Nil, fmt.Errorf("ERROR #1: %s", err)
	}
	idString, err := token.Claims.GetSubject()
	if err != nil {
		return uuid.Nil, fmt.Errorf("ERROR #2: %s", err)
	}
	id, err := uuid.Parse(idString)
	if err != nil {
		return uuid.Nil, fmt.Errorf("ERROR #3: %s", err)
	}

	return id, nil
}

func GetBearerToken(headers http.Header) (string, error) {
	tokenStr := headers.Get("Authorization")
	if tokenStr == "" {
		return "", errors.New("header not found")
	}
	wordlist := strings.Split(tokenStr, " ")

	return wordlist[1], nil

}

func MakeRefreshToken() (string, error) {
	key := make([]byte, 32)
	_, err := rand.Read(key)
	if err != nil {
		return "", err
	}
	tokenStr := hex.EncodeToString(key)
	return tokenStr, nil
}
