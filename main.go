package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

var cfg *config

func init() {
	c, err := newConfig()
	if err != nil {
		panic(err)
	}
	cfg = c
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

	log.Printf("%s", body)

	msg, err := newSlackMessage(body)
	if err != nil {
		log.Printf("unexpexted request body: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Challenge request
	// https://api.slack.com/events-api#the-events-api__subscribing-to-event-types__events-api-request-urls__request-url-configuration--verification
	if msg.isChallenge() {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, msg.Challenge)
		return
	}

	// Messages not to be processed
	if !msg.ok {
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
