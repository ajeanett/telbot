package services

import (
    "net/http"
	"log"
	"os"
)

func StartHealthServer() {
    go func() {
        http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
            w.Write([]byte("🤖 Bot is alive!"))
        })
        
        // Порт берётся из env-переменной (для хостинг-платформ)
        port := os.Getenv("PORT")
        if port == "" {
            port = "8080"
        }
        
        log.Printf("Health server started on :%s", port)
        http.ListenAndServe(":"+port, nil)
    }()
}