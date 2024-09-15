package handlers

import (
	"context"
	"encoding/json"
	"file-sharing-system/models"
	"file-sharing-system/utils"
	"log"
	"net/http"
)

func Register(w http.ResponseWriter, r *http.Request) {
    var user models.User
    err := json.NewDecoder(r.Body).Decode(&user)
    if err != nil {
        http.Error(w, "Invalid input", http.StatusBadRequest)
        return
    }


    if user.Email == "" || user.Password == "" {
        http.Error(w, "Email and password are required", http.StatusBadRequest)
        return
    }

    _, err = utils.Db.Exec(context.Background(),
        "INSERT INTO users (email, password) VALUES ($1, $2)", user.Email, user.Password)
    if err != nil {
        log.Printf("Error registering user: %v", err)  
        http.Error(w, "Error registering user", http.StatusInternalServerError)
        return
    }

    w.Write([]byte("User registered successfully"))
}


func Login(w http.ResponseWriter, r *http.Request) {
    var user models.User
    json.NewDecoder(r.Body).Decode(&user)

    var dbPassword string
    var userID int
    err := utils.Db.QueryRow(context.Background(),
        "SELECT id, password FROM users WHERE email = $1", user.Email).Scan(&userID, &dbPassword)

    if err != nil || dbPassword != user.Password {
        http.Error(w, "Invalid credentials", http.StatusUnauthorized)
        return
    }

    token, _ := utils.GenerateJWT(user.Email, userID)
    w.Write([]byte(token))
}
