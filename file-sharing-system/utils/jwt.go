package utils

import (
	"log"
	"time"

	"github.com/dgrijalva/jwt-go"
)

var jwtSecret = []byte("NzZbFMr2B+3j7BZvin8BCIEr/JcSPTdBvmO0MLjKDDE=")

func GenerateJWT(email string, userID int) (string, error) {
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
        "email":  email,
        "user_id": userID,
        "exp":    time.Now().Add(time.Hour * 24).Unix(),
    })

    return token.SignedString(jwtSecret)
}


func ValidateJWT(tokenString string) (*jwt.Token, error) {
    token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
        return jwtSecret, nil
    })
    if err != nil {
        log.Printf("JWT validation error: %v", err) 
        return nil, err
    }
    return token, nil
}
