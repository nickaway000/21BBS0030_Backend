package utils

import (
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
)

const RequestLimitPerMinute = 100

func RateLimitMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	
		if r.URL.Path == "/register" || r.URL.Path == "/login" {
			next.ServeHTTP(w, r)
			return
		}

	
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			log.Println("Authorization header is missing")
			http.Error(w, "Missing Authorization header", http.StatusUnauthorized)
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		token, err := ValidateJWT(tokenString)
		if err != nil || !token.Valid {
			log.Printf("Invalid token: %v", err)
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}


		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			log.Println("Failed to parse JWT claims")
			http.Error(w, "Failed to parse claims", http.StatusBadRequest)
			return
		}

		email, ok := claims["email"].(string)
		if !ok {
			log.Println("email is missing or not a string in token")
			http.Error(w, "email is missing or not a string in token", http.StatusBadRequest)
			return
		}

		log.Printf("Email extracted from token: %s", email)

	
		isAllowed, err := RateLimiter(email, RequestLimitPerMinute, time.Minute) 
		if err != nil {
			log.Printf("Rate limiter error: %v", err)
			http.Error(w, "Error applying rate limit", http.StatusInternalServerError)
			return
		}
		if !isAllowed {
			log.Println("Rate limit exceeded for email:", email)
			http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
			return
		}


		log.Println("Rate limit passed, proceeding to next handler")
		next.ServeHTTP(w, r)
	})
}