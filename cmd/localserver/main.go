package main

import (
	"log"
	"net/http"
	"os"

	ptsync "github.com/moviegeek/pt-rss-sync"
)

//main function is for local test
func main() {
	log.Printf("local testing mode...")

	port := os.Getenv("PORT")
	if port == "" {
		port = ":8080"
	}

	http.HandleFunc("/", ptsync.Handler)

	log.Printf("listening on %s", port)
	http.ListenAndServe(port, nil)
}
