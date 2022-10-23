package lib

import (
	"encoding/json"
	"log"
	"net/http"
)

// refreshHandler starts a new refresh of the data set
func (env *Env) RefreshHandler(w http.ResponseWriter, r *http.Request) {
	env.RunLogger.Println("refresh get")
	if r.Header.Get("X-API-KEY") != env.ApiKey {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	go env.Refresh()
	resp, err := json.Marshal("refresh started")
	if err != nil {
		log.Fatal(err)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(resp)
}
