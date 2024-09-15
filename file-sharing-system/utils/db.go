package utils

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"os"

	"github.com/jackc/pgx/v4"
	"github.com/joho/godotenv"
)

var Db *pgx.Conn

func ConnectDB() {
    err2 := godotenv.Load(".env")
	if err2 != nil {
		 log.Printf("error loading .env file: %w", err2)
	}

    var err error
    dbHost := os.Getenv("DB_HOST")
    dbPort := os.Getenv("DB_PORT")
    dbUser := os.Getenv("DB_USER")
    dbPassword := os.Getenv("DB_PASSWORD")
    dbName := os.Getenv("DB_NAME")


    log.Printf("DB_HOST=%s, DB_PORT=%s, DB_USER=%s, DB_NAME=%s", dbHost, dbPort, dbUser, dbName)

  
    dbPassword = url.QueryEscape(dbPassword)

 
    if dbHost == "" || dbPort == "" || dbUser == "" || dbPassword == "" || dbName == "" {
        log.Fatal("Missing required environment variables for database connection")
    }

    connStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", dbUser, dbPassword, dbHost, dbPort, dbName)

  
    Db, err = pgx.Connect(context.Background(), connStr)
    if err != nil {
        log.Fatalf("Unable to connect to the database: %v", err)
    }

    log.Println("Connected to the database successfully!")
}
