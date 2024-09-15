package utils

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/joho/godotenv"
)

const S3_BUCKET = "filesharingsys"


func ExpiredFileWorker(interval time.Duration) {
    ticker := time.NewTicker(interval) 
    defer ticker.Stop()

    for range ticker.C {
        log.Println("Running expired file cleanup...")
        err := cleanupExpiredFiles()
        if err != nil {
            log.Printf("Error during file cleanup: %v", err)
        }
    }
}


func cleanupExpiredFiles() error {
    
    rows, err := Db.Query(context.Background(), "SELECT id, file_name, s3_url FROM files WHERE expiration_date < NOW()")
    if err != nil {
        return err
    }
    defer rows.Close()

    var fileIDs []int
    var objectsToDelete []*s3.ObjectIdentifier

    
    for rows.Next() {
        var fileID int
        var fileName, s3URL string
        err := rows.Scan(&fileID, &fileName, &s3URL)
        if err != nil {
            log.Printf("Error scanning file: %v", err)
            continue
        }

        fileIDs = append(fileIDs, fileID)
        objectsToDelete = append(objectsToDelete, &s3.ObjectIdentifier{Key: aws.String(fileName)})
    }

    if len(objectsToDelete) == 0 {
        log.Println("No expired files found for deletion.")
        return nil
    }

  
    err = deleteFilesFromS3(objectsToDelete)
    if err != nil {
        log.Printf("Error deleting files from S3: %v", err)
        return err
    }

    
    err = deleteFilesFromDB(fileIDs)
    if err != nil {
        log.Printf("Error deleting file metadata from PostgreSQL: %v", err)
    }

    return nil
}


func deleteFilesFromS3(objectsToDelete []*s3.ObjectIdentifier) error {
    err2 := godotenv.Load(".env")
	if err2 != nil {
		 log.Printf("error loading .env file: %w", err2)
	}
    
    sess, err := session.NewSession(&aws.Config{Region: aws.String("eu-north-1")})
    if err != nil {
        return err
    }

    svc := s3.New(sess)

 
    _, err = svc.DeleteObjects(&s3.DeleteObjectsInput{
        Bucket: aws.String(S3_BUCKET),
        Delete: &s3.Delete{
            Objects: objectsToDelete,
        },
    })

    if err != nil {
        return err
    }

    log.Printf("Successfully deleted %d files from S3", len(objectsToDelete))

    return nil
}


func deleteFilesFromDB(fileIDs []int) error {
    if len(fileIDs) == 0 {
        return nil
    }

    
    tx, err := Db.Begin(context.Background())
    if err != nil {
        return err
    }

    defer func() {
        if err != nil {
            tx.Rollback(context.Background())
        } else {
            tx.Commit(context.Background())
        }
    }()

    for _, fileID := range fileIDs {
        _, err := tx.Exec(context.Background(), "DELETE FROM files WHERE id = $1", fileID)
        if err != nil {
            log.Printf("Error deleting metadata for file ID %d: %v", fileID, err)
            return err
        }
    }

    log.Printf("Successfully deleted metadata for %d files from PostgreSQL", len(fileIDs))

    return nil
}
