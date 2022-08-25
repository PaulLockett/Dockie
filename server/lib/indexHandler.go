package lib

import (
	"app/models"
	"encoding/json"
	"log"
	"net/http"
)

// indexHandler responds to requests with the internal
func (env *Env) IndexGetHandler(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("X-API-KEY") != env.ApiKey {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	checkpointJson, err := json.Marshal(env.Checkpoint)
	if err != nil {
		log.Fatal(err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(checkpointJson)
}

func (env *Env) IndexPutHandler(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("X-API-KEY") != env.ApiKey {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	userIDs := []string{}
	err := json.NewDecoder(r.Body).Decode(&userIDs)
	if err != nil {
		log.Fatal(err)
	}
	for _, userID := range userIDs {
		// check if userID is already in the checkpoint
		if userStub, ok := env.Checkpoint.UserMap[userID]; !ok {
			env.Checkpoint.UserMap[userID] = models.LocalUserStub{
				InNextEpoc:      true,
				TimesUsed:       0,
				InServedStorage: false,
				UserAuthKey:     "",
			}
		} else {
			userStub.InNextEpoc = true
			env.Checkpoint.UserMap[userID] = userStub
		}
	}
	err = env.Storage.PutCheckpoint(env.Checkpoint)
	if err != nil {
		log.Fatal(err)
	}
}

// Log when an appengine warmup request is used to create the new instance.
// Warmup steps are taken in main for consistency with "cold start" instances.
func (env *Env) WarmUpHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("warmup done")
}
