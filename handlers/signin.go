package handlers

import (
	// Стандартные библиотеки
	"encoding/json"
	"net/http"
	"os"
	"time"

	// Внешние библиотеки
	"github.com/golang-jwt/jwt/v5"

	// Внутренние библиотеки
	"go_final_project/model"
)

var jwtKey = []byte("my_secret_key")

func GenerateJWT(passwordHash string) (string, error) {
	expirationTime := time.Now().Add(8 * time.Hour)
	claims := &model.Claims{
		PasswordHash: passwordHash,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

func SigninHandler(w http.ResponseWriter, r *http.Request) {
	var creds struct {
		Password string `json:"password"`
	}

	err := json.NewDecoder(r.Body).Decode(&creds)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	pass := os.Getenv("TODO_PASSWORD")
	if creds.Password != pass {
		respondWithError(w, http.StatusUnauthorized, "Неверный пароль")
		return
	}

	token, err := GenerateJWT(creds.Password)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to generate token")
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"token": token})

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]string{"token": token}); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error encoding response")
	}
}
