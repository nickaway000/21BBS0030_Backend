package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"file-sharing-system/utils"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/dgrijalva/jwt-go"
	"github.com/joho/godotenv"
)


const ENCRYPTION_PASSPHRASE = "ThisIsA32ByteLongPassphrase2024!" 

const (
    S3_REGION = "eu-north-1" 
    S3_BUCKET = "filesharingsys" 
)

func UploadFile(w http.ResponseWriter, r *http.Request) {
    log.Println("UploadFile request received")

   
    err := godotenv.Load(".env")
    if err != nil {
        log.Printf("Error loading .env file: %v", err)
        http.Error(w, "Internal server error", http.StatusInternalServerError)
        return
    }

   
    authHeader := r.Header.Get("Authorization")
    tokenString := strings.TrimPrefix(authHeader, "Bearer ")
    token, err := utils.ValidateJWT(tokenString)
    if err != nil || !token.Valid {
        log.Printf("Invalid token: %v", err)
        http.Error(w, "Invalid token", http.StatusUnauthorized)
        return
    }

    claims, ok := token.Claims.(jwt.MapClaims)
    if !ok {
        log.Printf("Invalid token claims")
        http.Error(w, "Invalid token claims", http.StatusUnauthorized)
        return
    }

    userID, ok := claims["user_id"].(float64)
    if !ok {
        log.Printf("Invalid user ID in token")
        http.Error(w, "Invalid user ID in token", http.StatusUnauthorized)
        return
    }

    
    file, handler, err := r.FormFile("file")
    if err != nil {
        log.Printf("Error retrieving file: %v", err)
        http.Error(w, "Error retrieving file", http.StatusBadRequest)
        return
    }
    defer file.Close()

    fileType := handler.Header.Get("Content-Type")
    sess, err := session.NewSession(&aws.Config{Region: aws.String(S3_REGION)})
    if err != nil {
        log.Printf("AWS session creation error: %v", err)
        http.Error(w, "Could not create AWS session", http.StatusInternalServerError)
        return
    }

    uploader := s3manager.NewUploader(sess)
    fileKey := fmt.Sprintf("%s_%d_%d", handler.Filename, time.Now().Unix(), int(userID))

    fileBytes, err := io.ReadAll(file)
    if err != nil {
        log.Printf("Error reading file: %v", err)
        http.Error(w, "Could not read file", http.StatusInternalServerError)
        return
    }

   
    encryptedFile, err := utils.Encrypt(fileBytes, ENCRYPTION_PASSPHRASE)
    if err != nil {
        log.Printf("Error encrypting file: %v", err)
        http.Error(w, "Could not encrypt file", http.StatusInternalServerError)
        return
    }

    
    upParams := &s3manager.UploadInput{
        Bucket:      aws.String(S3_BUCKET),
        Key:         aws.String(fileKey),
        Body:        bytes.NewReader(encryptedFile),
        ACL:         aws.String("public-read"),  
        ContentType: aws.String(fileType),      
    }

    result, err := uploader.Upload(upParams)
    if err != nil {
        log.Printf("S3 upload error: %v", err)
        http.Error(w, "Could not upload file to S3", http.StatusInternalServerError)
        return
    }

    publicURL := result.Location
    log.Printf("File uploaded successfully: %s", publicURL)

    
    tx, err := utils.Db.Begin(context.Background())
    if err != nil {
        log.Printf("Error starting transaction: %v", err)
        http.Error(w, "Error starting transaction", http.StatusInternalServerError)
        return
    }
    defer func() {
        if err != nil {
            tx.Rollback(context.Background())
        } else {
            tx.Commit(context.Background())
        }
    }()

    expiryDate := time.Now().Add(30 * 24 * time.Hour)

    _, err = tx.Exec(context.Background(),
        "INSERT INTO files (user_id, file_name, file_type, size, s3_url, upload_date, expiration_date, last_modified) VALUES ($1, $2, $3, $4, $5, NOW(), $6, NOW())",
        int(userID), handler.Filename, fileType, handler.Size, publicURL, expiryDate)
    if err != nil {
        log.Printf("Error saving file metadata: %v", err)
        http.Error(w, "Error saving file metadata", http.StatusInternalServerError)
        return
    }


    response := fmt.Sprintf("File uploaded successfully. Access it at: %s", publicURL)
    log.Printf("Sending response: %s", response)
    w.Header().Set("Content-Type", "text/plain")
    w.WriteHeader(http.StatusOK)
    w.Write([]byte(response))
}











type FileMetadata struct {
    FileName     string    `json:"file_name"`
    Size         int64     `json:"size"`
    PublicURL    string    `json:"public_url"`
    LastModified time.Time `json:"last_modified"`  
}

func RetrieveFile(w http.ResponseWriter, r *http.Request) {
    
    fileID := r.URL.Query().Get("file_id")

  
    authHeader := r.Header.Get("Authorization")
    tokenString := strings.TrimPrefix(authHeader, "Bearer ")

    token, err := utils.ValidateJWT(tokenString)
    if err != nil || !token.Valid {
        http.Error(w, "Invalid token", http.StatusUnauthorized)
        return
    }

   
    cachedMetadata, err := utils.GetCache(fileID)
    if err == nil {
        var cachedFileMetadata FileMetadata
        if err := json.Unmarshal([]byte(cachedMetadata), &cachedFileMetadata); err == nil {
     
            log.Printf("Cache hit for file ID: %s", fileID)
            updateCacheAndServeFile(w, r, fileID, cachedFileMetadata)
            return
        }
    }


    log.Printf("Cache miss for file ID: %s, error: %v", fileID, err)

  
    var dbFileMetadata FileMetadata
    err = utils.Db.QueryRow(context.Background(),
        "SELECT file_name, size, s3_url, last_modified FROM files WHERE id = $1", fileID).
        Scan(&dbFileMetadata.FileName, &dbFileMetadata.Size, &dbFileMetadata.PublicURL, &dbFileMetadata.LastModified)
    if err != nil {
        http.Error(w, "File not found", http.StatusNotFound)
        return
    }


    updateCacheAndServeFile(w, r, fileID, dbFileMetadata)
}


func updateCacheAndServeFile(w http.ResponseWriter, r *http.Request, fileID string, fileMetadata FileMetadata) {

    metadataJSON, _ := json.Marshal(fileMetadata)
    err := utils.SetCache(fileID, string(metadataJSON), 5*time.Minute)
    if err != nil {
        log.Printf("Error caching file metadata for file ID: %s, error: %v", fileID, err)
    }


    serveFileWithDecryption(w, r, fileMetadata.PublicURL)
}


func serveFileWithDecryption(w http.ResponseWriter, r *http.Request, fileURL string) {
    err2 := godotenv.Load(".env")
    if err2 != nil {
        log.Printf("Error loading .env file: %v", err2)
    }

   
    resp, err := http.Get(fileURL)
    if err != nil || resp.StatusCode != http.StatusOK {
        log.Printf("Error fetching file from S3: %v", err)
        http.Error(w, "File not found", http.StatusNotFound)
        return
    }
    defer resp.Body.Close()


    log.Printf("S3 response status: %d", resp.StatusCode)
    log.Printf("S3 response headers: %v", resp.Header)


    encryptedContent, err := io.ReadAll(resp.Body)
    if err != nil {
        log.Printf("Error reading file content: %v", err)
        http.Error(w, "Error reading file", http.StatusInternalServerError)
        return
    }
    log.Printf("Encrypted content length: %d", len(encryptedContent))


    decryptedContent, err := utils.Decrypt(encryptedContent, ENCRYPTION_PASSPHRASE)
    if err != nil {
        log.Printf("Error decrypting file: %v", err)
        http.Error(w, "Error decrypting file", http.StatusInternalServerError)
        return
    }

    log.Printf("Decrypted content length: %d", len(decryptedContent))


    w.Header().Set("Content-Type", resp.Header.Get("Content-Type"))
    w.Write(decryptedContent)
}








func ShareFile(w http.ResponseWriter, r *http.Request) {
    fileID := r.URL.Query().Get("file_id")
    expirationStr := r.URL.Query().Get("expiration") 


    expiration, err := strconv.Atoi(expirationStr)
    if err != nil || expiration <= 0 {
        expiration = 30
    }


    var fileKey string
    err = utils.Db.QueryRow(context.Background(),
        "SELECT s3_url FROM files WHERE id = $1", fileID).Scan(&fileKey)
    if err != nil {
        http.Error(w, "File not found", http.StatusNotFound)
        return
    }


    keyParts := strings.Split(fileKey, "/")
    fileKey = keyParts[len(keyParts)-1] 


    sess, err := session.NewSession(&aws.Config{Region: aws.String(S3_REGION)})
    if err != nil {
        http.Error(w, "Could not create AWS session", http.StatusInternalServerError)
        return
    }

    svc := s3.New(sess)


    req, _ := svc.GetObjectRequest(&s3.GetObjectInput{
        Bucket: aws.String(S3_BUCKET),
        Key:    aws.String(fileKey), 
    })

 
    tempPublicURL, err := req.Presign(time.Duration(expiration) * time.Minute)
    if err != nil {
        http.Error(w, "Failed to sign URL", http.StatusInternalServerError)
        return
    }


    response := struct {
        URL           string `json:"url"`
        DecryptionKey string `json:"decryption_key"`
    }{
        URL:           tempPublicURL,
        DecryptionKey: ENCRYPTION_PASSPHRASE, 
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(response)
}






func ServeFile(w http.ResponseWriter, r *http.Request) {
    fileID := r.URL.Query().Get("file_id")
    expiresStr := r.URL.Query().Get("expires")

    expires, err := strconv.ParseInt(expiresStr, 10, 64)
    if err != nil || time.Now().Unix() > expires {
        http.Error(w, "File link expired", http.StatusGone)
        return
    }

 
    cachedURL, err := utils.GetCache("share_" + fileID)
    if err != nil {
        http.Error(w, "Invalid or expired link", http.StatusGone)
        return
    }


    http.Redirect(w, r, cachedURL, http.StatusSeeOther)
}


type File struct {
    ID         int       `json:"id"`
    FileName   string    `json:"file_name"`
    FileType   string    `json:"file_type"`
    Size       int64     `json:"size"`
    S3URL      string    `json:"s3_url"`
    UploadDate time.Time `json:"upload_date"`
}


func SearchFiles(w http.ResponseWriter, r *http.Request) {
    fileName := r.URL.Query().Get("filename")
    fileType := r.URL.Query().Get("file_type")
    uploadDate := r.URL.Query().Get("upload_date") 


    query := "SELECT id, file_name, file_type, size, s3_url, upload_date FROM files WHERE 1=1"
    args := []interface{}{}
    argIndex := 1

 
    if fileName != "" {
        query += fmt.Sprintf(" AND file_name ILIKE '%%' || $%d || '%%'", argIndex)
        args = append(args, fileName)
        argIndex++
    }
    if fileType != "" {
        query += fmt.Sprintf(" AND file_type = $%d", argIndex)
        args = append(args, fileType)
        argIndex++
    }
    if uploadDate != "" {
        query += fmt.Sprintf(" AND upload_date::DATE = $%d::DATE", argIndex)
        args = append(args, uploadDate)
        argIndex++
    }

   
    rows, err := utils.Db.Query(context.Background(), query, args...)
    if err != nil {
        http.Error(w, "Error searching files", http.StatusInternalServerError)
        return
    }
    defer rows.Close()

   
    var files []File
    for rows.Next() {
        var file File
        if err := rows.Scan(&file.ID, &file.FileName, &file.FileType, &file.Size, &file.S3URL, &file.UploadDate); err != nil {
            http.Error(w, "Error scanning results", http.StatusInternalServerError)
            return
        }
        files = append(files, file)
    }

   
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(files)
}



