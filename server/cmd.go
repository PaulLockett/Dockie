package main

import (
	"app/lib"
	"app/models"
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
)

func main() {
	startupTime := time.Now()

	storageModel := new(models.StorageModel)
	storageModel.Setup(context.Background())
	checkpoint, err := storageModel.GetCheckpoint()
	if err != nil {
		checkpoint = models.LocalCheckpoint{
			CurrentEpoc: 0,
			UserMap:     make(map[string]models.LocalUserStub),
		}
	}

	env := &lib.Env{
		StartTime:  startupTime,
		Checkpoint: checkpoint,
		Storage:    storageModel,
	}

	r := mux.NewRouter()

	r.HandleFunc("/_ah/warmup", env.WarmUpHandler)

	r.HandleFunc("/", env.IndexGetHandler).Methods("GET")

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
		log.Printf("Defaulting to port %s", port)
	}
	log.Printf("Listening on port %s", port)

	log.Fatal(http.ListenAndServe(":"+port, r))
}
