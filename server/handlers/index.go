package handlers

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

// indexHandler responds to requests with our greeting.
func (env *Env) IndexHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	uptime := time.Since(env.StartTime).Seconds()
	fmt.Fprintf(w, "Hello, World! Uptime: %.2fs\n", uptime)
}

// Log when an appengine warmup request is used to create the new instance.
// Warmup steps are taken in setup for consistency with "cold start" instances.
func (env *Env) WarmUpHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("warmup done")
}
