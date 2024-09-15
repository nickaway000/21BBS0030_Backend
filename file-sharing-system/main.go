package main

import (
    "file-sharing-system/handlers"
    "file-sharing-system/utils"
    "fmt"
    "log"
    "net/http"
    "time"
)

func main() {
    
    utils.ConnectDB()
    utils.ConnectRedis()

   
    go utils.ExpiredFileWorker(1 * time.Hour)

   
    mux := http.NewServeMux()

    
    mux.HandleFunc("/register", handlers.Register)
    mux.HandleFunc("/login", handlers.Login)
    mux.HandleFunc("/upload", handlers.UploadFile)
    mux.HandleFunc("/retrieve", handlers.RetrieveFile)
    mux.HandleFunc("/share", handlers.ShareFile)
    mux.HandleFunc("/serve", handlers.ServeFile)
    mux.HandleFunc("/search", handlers.SearchFiles)

  
    mux.HandleFunc("/ws", utils.WebSocketHandler)


    mux.Handle("/", http.FileServer(http.Dir("./static")))

    
    rateLimitedMux := utils.RateLimitMiddleware(mux)

 
    srv := &http.Server{
        Addr:         ":9080",
        Handler:      rateLimitedMux,  
        ReadTimeout:  10 * time.Second,
        WriteTimeout: 10 * time.Second,
        IdleTimeout:  120 * time.Second,
    }

    fmt.Printf("Starting server at port 9080\n")
    log.Fatal(srv.ListenAndServe())
}
