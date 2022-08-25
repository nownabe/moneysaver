package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/slack-go/slack"
)

type handler struct {
	eventProcessor   *eventProcessor
	commandProcessor *commandProcessor
}

func (h *handler) handleEvents(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if contentType := r.Header.Get("Content-Type"); contentType != "application/json" {
		logger.Printf("Unsupported content type: %s", contentType)
		w.WriteHeader(http.StatusUnsupportedMediaType)

		return
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		logger.Printf("ioutil.ReadAll: %v", err)
		w.WriteHeader(http.StatusConflict)

		return
	}
	defer r.Body.Close()

	logger.Printf("%s", body)

	var msg slackMessage

	if err := json.Unmarshal(body, &msg); err != nil {
		logger.Printf("json.Unmarshal: %v", err)
		w.WriteHeader(http.StatusBadRequest)

		return
	}

	resBody, err := h.eventProcessor.process(ctx, &msg)
	if err != nil {
		logger.Printf("h.eventProcessor.process: %v", err)
		writeErrorHeader(w, err)

		return
	}

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, resBody)
}

func (h *handler) handleCommands(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	s, err := slack.SlashCommandParse(r)
	if err != nil {
		logger.Printf("slack.SlashCommandParse: %v", err)
		w.WriteHeader(http.StatusInternalServerError)

		return
	}

	resp, err := h.commandProcessor.process(ctx, s)
	if err != nil {
		logger.Printf("h.commandProcessor.process: %v", err)
		writeErrorHeader(w, err)

		return
	}

	b, err := json.Marshal(resp)
	if err != nil {
		logger.Printf("json.Marshal: %v", err)
		w.WriteHeader(http.StatusInternalServerError)

		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if _, err := w.Write(b); err != nil {
		logger.Printf("w.Write: %v", err)
	}
}
