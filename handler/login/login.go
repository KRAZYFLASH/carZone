package login

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/KRAZYFLASH/carZone/models"
	"github.com/golang-jwt/jwt/v4"
)



func LoginHandler(w http.ResponseWriter, r *http.Request) {
	var credentials models.Credential
	if err := json.NewDecoder(r.Body).Decode(&credentials); err != nil {
		http.Error(w, "Invalid Request Body", http.StatusBadRequest)
		return
	}

	valid := (credentials.Username == "admin" && credentials.Password == "admin123")

	if !valid {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	tokenString, err := GenerateToken(credentials.Username)
	if err != nil {
		http.Error(w, "Error generating token", http.StatusInternalServerError)
		log.Println("Error generating token:", err)
		return
	}

	response := map[string]string{"token": tokenString}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}


func GenerateToken(username string) (string, error) {
	expiration := time.Now().Add(24 * time.Hour)

	claims := &jwt.RegisteredClaims{
		Subject:   username,
		ExpiresAt: jwt.NewNumericDate(expiration),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte("some_value")) // TODO: ganti ke ENV:
	if err != nil {
		return "", err
	}

	return signedToken, nil
}