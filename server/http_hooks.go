package main

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/mattermost/mattermost/server/public/model"
	"github.com/mattermost/mattermost/server/public/plugin"
	"github.com/mattermost/mattermost/server/public/pluginapi"
)

func (p *Plugin) ServeHTTP(c *plugin.Context, w http.ResponseWriter, r *http.Request) {
	p.router.ServeHTTP(w, r)
}

func (p *Plugin) initializeAPI() {
	router := mux.NewRouter()

	router.HandleFunc("/alert", p.handleAlert)

	p.router = router
}

func (p *Plugin) ensureBot() {
	// if p.botID != "" {
	// 	p.client.Bot.DeletePermanently(p.botID)
	// }
	botID, err := p.client.Bot.EnsureBot(&model.Bot{
		Username:    "sentrybot",
		DisplayName: "Sentry Bot",
		Description: "A bot account created by sentry-mattermost-plugin.",
	}, pluginapi.ProfileImagePath("/assets/sentry-icon.png"))

	if err != nil {
		return
	}

	p.botID = botID
}

func (p *Plugin) handleAlert(w http.ResponseWriter, r *http.Request) {
	// configuration := p.getConfiguration()
	defer r.Body.Close()

	// Read the request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}

	// Parse the request body into a map
	var sentryPayload map[string]interface{}
	if err := json.Unmarshal(body, &sentryPayload); err != nil {
		http.Error(w, "Failed to parse request body", http.StatusBadRequest)
		return
	}

	// Extract information from the Sentry payload
	event := sentryPayload["data"].(map[string]interface{})["event"].(map[string]interface{})
	datetime := event["datetime"].(string)
	exception := event["exception"].(map[string]interface{})["values"].([]interface{})[0].(map[string]interface{})
	exceptionType := exception["type"].(string)
	exceptionMessage := exception["value"].(string)

	tags := event["tags"].([]interface{})
	exceptionURL := ""
	for _, tag := range tags {
		tagMap := tag.([]interface{})
		if tagMap[0].(string) == "url" {
			exceptionURL = tagMap[1].(string)
			break
		}
	}

	alertRule := sentryPayload["data"].(map[string]interface{})["triggered_rule"].(string)
	issueURL := event["web_url"].(string)

	type DebugInfo struct {
		Datetime         string `json:"datetime"`
		ExceptionType    string `json:"exceptionType"`
		ExceptionMessage string `json:"exceptionMessage"`
		ExceptionURL     string `json:"exceptionURL"`
		AlertRule        string `json:"alertRule"`
		IssueURL         string `json:"issueURL"`
	}

	debugInfo := &DebugInfo{
		Datetime:         datetime,
		ExceptionType:    exceptionType,
		ExceptionMessage: exceptionMessage,
		ExceptionURL:     exceptionURL,
		AlertRule:        alertRule,
		IssueURL:         issueURL,
	}

	// Create a post to cyberbiz-checkout-alert
	if _, err := p.API.CreatePost(&model.Post{
		UserId:    p.botID,
		ChannelId: "1bws98wmotdttprzgsqa1bqt5o", // cyberbiz-checkout-alert
		Message:   "test error sent from sentry",
	}); err != nil {
		http.Error(w, "Failed to create the post", http.StatusBadRequest)
		return
	}

	responseJSON, _ := json.Marshal(debugInfo)

	w.Header().Set("Content-Type", "application/json")
	if _, err := w.Write(responseJSON); err != nil {
		p.API.LogError("Failed to write status", "err", err.Error())
	}
}

// func (p *Plugin) ServeHTTP2(c *plugin.Context, w http.ResponseWriter, r *http.Request) {
// 	fmt.Fprint(w, "Hello, world!")
// }
