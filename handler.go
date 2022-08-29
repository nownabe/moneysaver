package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
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

	ev, err := slackevents.ParseEvent(json.RawMessage(body), slackevents.OptionNoVerifyToken())
	if err != nil {
		logger.Printf("slackevents.ParseEvent: %v", err)
		w.WriteHeader(http.StatusInternalServerError)

		return
	}

	if ev.Type == slackevents.URLVerification {
		h.handleChallenge(w, body)

		return
	}

	if err := h.eventProcessor.process(ctx, ev); err != nil {
		logger.Printf("h.eventProcessor.process: %v", err)
		writeErrorHeader(w, err)

		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *handler) handleChallenge(w http.ResponseWriter, body []byte) {
	var r *slackevents.ChallengeResponse
	if err := json.Unmarshal(body, &r); err != nil {
		logger.Printf("json.Unmarshal: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte(r.Challenge)); err != nil {
			logger.Printf("w.Write: %v", err)
		}
	}
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
