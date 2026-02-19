package auth

import (
	"errors"
	"net/http"
	"runtime"
	"strings"

	"github.com/alexedwards/argon2id"
)

func HashPassword(password string) (string, error) {
	params := &argon2id.Params{
		Memory:      128 * 1024,
		Iterations:  4,
		Parallelism: uint8(runtime.NumCPU()),
		SaltLength:  0,
		KeyLength:   32,
	}

	hash, err := argon2id.CreateHash(password, params)
	if err != nil {
		return "", err
	}
	return hash, nil
}

func CheckPasswordHash(password, hash string) (bool, error) {
	match, err := argon2id.ComparePasswordAndHash(password, hash)
	if err != nil {
		return false, err
	}
	return match, nil
}

func GetAPIKey(headers http.Header) (string, error) {
	keyStr := headers.Get("Authorization")
	if keyStr == "" {
		return "", errors.New("header not found")
	}
	wordlist := strings.Split(keyStr, " ")

	return wordlist[1], nil
}
