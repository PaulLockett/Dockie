package main

import (
	"app/lib"
	"app/models"
	"context"
	"log"
	"net/http"
	"os"
	"time"

	twitter "github.com/g8rswimmer/go-twitter/v2"
	"github.com/go-co-op/gocron"
	"github.com/gorilla/mux"
)

func main() {
	// setup data structures and sources
	startupTime := time.Now()

	// setup google cloud storage
	storageModel := new(models.StorageModel)
	storageModel.Setup(context.Background())
	checkpoint, err := storageModel.GetCheckpoint()
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
		StartTime:     startupTime,
		Checkpoint:    checkpoint,
		Storage:       storageModel,
		TwitterClient: client,
	}

	// setup cron job
	refreshScheduler := gocron.NewScheduler(time.UTC)
	refreshScheduler.SingletonMode().Every(1).Day().At("00:00").Do(env.Refresh)
	refreshScheduler.StartAsync()

	router := mux.NewRouter()

	router.HandleFunc("/_ah/warmup", env.WarmUpHandler)

	router.HandleFunc("/", env.IndexGetHandler).Methods("GET")
	router.HandleFunc("/", env.IndexPutHandler).Methods("PUT")

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
		log.Printf("Defaulting to port %s", port)
	}
	log.Printf("Listening on port %s", port)

	log.Fatal(http.ListenAndServe(":"+port, router))
}
