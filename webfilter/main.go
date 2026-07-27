package main

import (
	"log"
	"net/http"
)

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /filters/{name}", processHandler)
	mux.HandleFunc("GET /health", healthHandler)
	mux.Handle("/", http.FileServer(http.Dir("static")))

	log.Println("listening on :8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatal(err)
	}
}
