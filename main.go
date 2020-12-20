package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
)

var numRegexp *regexp.Regexp = regexp.MustCompile(`^\d+$`)

type slackEvent struct {
	Channel string `json:"channel"`
	Text    string `json:"text"`
}

type slackMessage struct {
	Challenge string     `json:"challenge"`
	TeamID    string     `json:"team_id"`
	Event     slackEvent `json:"event"`
}

func main() {
	http.HandleFunc("/", handler)
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("failed to listen and serve: %v", err)
	}
}

func handler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	if r.Header.Get("Content-Type") != "application/json" {
		w.WriteHeader(http.StatusUnsupportedMediaType)
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("failed to read request body: %v", err)
		w.WriteHeader(http.StatusConflict)
		return
	}
	defer r.Body.Close()

	var msg slackMessage
	if err := json.Unmarshal(body, &msg); err != nil {
		log.Printf("failed to unmarshal request body: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if msg.Challenge != "" {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"challenge":"%s"}`, msg.Challenge)
		return
	}

	if !numRegexp.Match([]byte(msg.Event.Text)) {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	if err := reply(ctx, msg.Event.Channel, 12345); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
