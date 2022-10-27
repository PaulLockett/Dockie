package main

import (
	"app/lib"
	"app/models"
	"log"
	"net/http"
	"os"
	"time"

	twitter "github.com/g8rswimmer/go-twitter/v2"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

func main() {
	// setup data structures and sources
	startupTime := time.Now()

	err := godotenv.Load(".env.local")
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// setup storage
	storage := models.StorageModel{}
	if err := storage.Setup(os.Getenv("SURREALDB_URL"), os.Getenv("SURREALDB_USER"), os.Getenv("SURREALDB_PASSWORD"), os.Getenv("SURREALDB_NAMESPACE"), os.Getenv("SURREALDB_DATABASE")); err != nil {
		log.Fatal(err)
	}

	log.Printf("storage setup complete")
	// setup checkpoint
	checkpoint, err := storage.GetCheckpoint()
	if err != nil {
		checkpoint = models.LocalCheckpoint{
			CurrentEpoc: 0,
			UserMap:     make(map[string]models.LocalUserStub),
		}
	}

	// setup twitter client
	client := &twitter.Client{
		Authorizer: models.Authorize{
			Token: os.Getenv("BEARER_TOKEN"),
		},
		Client: http.DefaultClient,
		Host:   "https://api.twitter.com",
	}

	env := &lib.Env{
		APIKey:        os.Getenv("API_KEY"),
		StartTime:     startupTime,
		Checkpoint:    checkpoint,
		Storage:       &storage,
		TwitterClient: client,
	}

	// go env.Refresh()

	router := mux.NewRouter()

	router.HandleFunc("/", env.IndexGetHandler).Methods("GET")
	router.HandleFunc("/", env.IndexPutHandler).Methods("PUT")
	router.HandleFunc("/Refresh", env.RefreshHandler).Methods("Get")

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
		log.Printf("Defaulting to port %s", port)
	}
	log.Printf("Listening on port %s", port)

	log.Fatal(http.ListenAndServe(":"+port, router))
}
