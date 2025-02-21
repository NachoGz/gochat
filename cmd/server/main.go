package main

import (
	"log"
	"net/http"
	"os"

	"github.com/NachoGz/gochat/internal/server"
	"github.com/joho/godotenv"
)

func main() {
	server := server.NewServer()
	go server.Start()

	router := http.NewServeMux()

	router.HandleFunc("/ws", server.HandleWS)
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "web/templates/index.html")
	})

	godotenv.Load()
	port := os.Getenv("PORT")
	if port == "" {
		log.Fatal("PORT must be set")
	}

	log.Printf("Server listening on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, router))
}
