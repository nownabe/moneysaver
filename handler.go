package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type handler struct {
	eventProcessor *eventProcessor
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
