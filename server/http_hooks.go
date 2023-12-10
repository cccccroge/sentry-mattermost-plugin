package main

import (
	"encoding/json"
	"net/http"

	sentry "github.com/getsentry/sentry-go"
	"github.com/gorilla/mux"
	"github.com/mattermost/mattermost/server/public/plugin"
)

func (p *Plugin) ServeHTTP(c *plugin.Context, w http.ResponseWriter, r *http.Request) {
	p.router.ServeHTTP(w, r)
}

func (p *Plugin) initializeAPI() {
	router := mux.NewRouter()

	router.HandleFunc("/alert", p.handleAlert)

	p.router = router
}

func (p *Plugin) handleAlert(w http.ResponseWriter, r *http.Request) {
	// configuration := p.getConfiguration()

	var request sentry.Request
	err := json.NewDecoder(r.Body).Decode(&request)

	if err != nil {
		p.API.LogError("Failed to decode sentry request", "err", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	defer r.Body.Close()

	var response = struct {
		Data string `json:"data"`
	}{
		Data: request.Data,
	}

	responseJSON, _ := json.Marshal(response)

	w.Header().Set("Content-Type", "application/json")
	if _, err := w.Write(responseJSON); err != nil {
		p.API.LogError("Failed to write status", "err", err.Error())
	}
}

// func (p *Plugin) ServeHTTP2(c *plugin.Context, w http.ResponseWriter, r *http.Request) {
// 	fmt.Fprint(w, "Hello, world!")
// }
