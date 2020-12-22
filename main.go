package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

var (
	limits    = map[string]int64{}
	projectID = os.Getenv("PROJECT_ID")
)

type slackEvent struct {
	Channel string  `json:"channel"`
	Text    string  `json:"text"`
	EventTS float64 `json:"event_ts"`
}

func (e slackEvent) timestamp() time.Time {
	return time.Unix(int64(e.EventTS), 0)
}

func (e slackEvent) amount() (int64, error) {
	return strconv.ParseInt(e.Text, 10, 64)
}

// https://api.slack.com/events-api#the-events-api__receiving-events__callback-field-overview
type slackMessage struct {
	// TODO: Token
	Challenge string     `json:"challenge"`
	TeamID    string     `json:"team_id"`
	Event     slackEvent `json:"event"`
}

func main() {
	parseLimits()
	http.HandleFunc("/", handler)
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("failed to listen and serve: %v", err)
	}
}

func parseLimits() {
	chs := strings.Split(os.Getenv("LIMITS"), ",")
	for _, ch := range chs {
		parts := strings.Split(ch, ":")
		l, err := strconv.ParseInt(parts[1], 10, 64)
		if err != nil {
			panic(fmt.Sprintf("failed to parse LIMITS: %v", err))
		}

		limits[parts[0]] = l
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

	log.Printf("%s", body)

	var msg slackMessage
	if err := json.Unmarshal(body, &msg); err != nil {
		log.Printf("failed to unmarshal request body: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if msg.Challenge != "" {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, msg.Challenge)
		return
	}

	if _, err := msg.Event.amount(); err != nil {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	if err := addRecord(ctx, msg); err != nil {
		log.Printf("failed to add record: %v", err)
		// TODO: reply error
	}

	total, err := aggregate(ctx, msg)
	if err != nil {
		log.Printf("failed to aggregate: %v", err)
		// TODO: reply error
	}

	if err := reply(ctx, msg, total); err != nil {
		log.Printf("failed to reply: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
