package main

import (
	"context"
	"log"
	"net/http"

	"github.com/nownabe/moneysaver/slack"
)

func main() {
	c, err := newConfig()
	if err != nil {
		panic(err)
	}

	s, err := newStoreClient(context.Background(), c.ProjectID)
	if err != nil {
		panic(err)
	}

	h := &handler{
		cfg:   c,
		store: s,
		slack: slack.New(c.SlackBotToken),
	}

	// Start server
	if err := http.ListenAndServe(":8080", h); err != nil {
		log.Fatalf("failed to listen and serve: %v", err)
	}
}
