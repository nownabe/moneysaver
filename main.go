package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/nownabe/moneysaver/slack"
)

const (
	timeoutSec = 60
)

// Align chi middleware
// https://github.com/go-chi/chi/blob/b6a2c5a909f66db8b2166b69628fff095ed51adc/middleware/logger.go#L170
var logger = log.New(os.Stdout, "", log.LstdFlags)

func main() {
	c, err := newConfig()
	if err != nil {
		panic(err)
	}

	ctx := context.Background()

	fs, err := firestore.NewClient(ctx, c.ProjectID)
	if err != nil {
		panic(err)
	}

	ep := &eventProcessor{
		cfg:         c,
		store:       &storeClient{fs},
		slack:       slack.New(c.SlackBotToken),
		channelRepo: &channelRepo{fs},
	}

	cp := &commandProcessor{
		channelRepo: &channelRepo{fs},
	}

	h := &handler{
		eventProcessor:   ep,
		commandProcessor: cp,
	}

	r := newRouter(h, c.SlackSigningSecret)

	// Start server
	if err := http.ListenAndServe(":8080", r); err != nil {
		logger.Fatalf("failed to listen and serve: %v", err)
	}
}

func newRouter(h *handler, signingSecret string) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.RequestLogger(&middleware.DefaultLogFormatter{Logger: logger, NoColor: false}))
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(timeoutSec * time.Second))
	r.Use(slackVerifier(signingSecret))

	r.Post("/", h.handleEvents)
	r.Post("/commands", h.handleCommands)

	return r
}
