package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v4"
)

var jwtKey = []byte("some_value") // TODO: ganti ke ENV: os.Getenv("JWT_SECRET")

// Gunakan RegisteredClaims (v4), bukan StandardClaims
type Claims struct {
	Username string `json:"username"`
	jwt.RegisteredClaims
}

// Hindari string sebagai context key
type ctxKey string

const userCtxKey ctxKey = "username"

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Missing Authorization header", http.StatusUnauthorized)
			return
		}
		if !strings.HasPrefix(authHeader, "Bearer") {
			http.Error(w, "Invalid Authorization scheme", http.StatusUnauthorized)
			return
		}

		tokenString := strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer "))

		claims := &Claims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			// Pastikan algoritma yang dipakai sesuai (HMAC/HS256)
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return jwtKey, nil
		})

		if err != nil || token == nil || !token.Valid {
			http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
			return
		}

		// (Optional) Validasi tambahan iss/aud kalau kalian set saat membuat token
		// if claims.Issuer != "your-issuer" { ... }

		ctx := context.WithValue(r.Context(), userCtxKey, claims.Username)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
