package main

import (
	"app/lib"
	"app/models"
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"cloud.google.com/go/storage"
	"github.com/gorilla/mux"
)

func main() {
	startupTime := time.Now()
	client, err := storage.NewClient(context.Background())
	if err != nil {
		log.Fatalf("setup: %v", err)
	}

	env := &lib.Env{
		StartTime: startupTime,
		Storage: &models.StorageModel{
			Client: client,
		},
	}

	r := mux.NewRouter()

	r.HandleFunc("/_ah/warmup", env.WarmUpHandler)

	r.HandleFunc("/", env.IndexHandler).Methods("GET")

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
		log.Printf("Defaulting to port %s", port)
	}
	log.Printf("Listening on port %s", port)

	log.Fatal(http.ListenAndServe(":"+port, r))
}
