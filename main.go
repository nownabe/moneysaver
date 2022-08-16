package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

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

	s, err := newStoreClient(context.Background(), c.ProjectID)
	if err != nil {
		panic(err)
	}

	h := &handler{
		cfg:   c,
		store: s,
		slack: slack.New(c.SlackBotToken),
	}

	r := newRouter(h)

	// Start server
	if err := http.ListenAndServe(":8080", r); err != nil {
		logger.Fatalf("failed to listen and serve: %v", err)
	}
}

func newRouter(h *handler) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.RequestLogger(&middleware.DefaultLogFormatter{logger, false}))
	r.Use(middleware.Recoverer)

	r.Use(middleware.Timeout(timeoutSec * time.Second))

	r.Post("/", h.handleEvents)

	return r
}
