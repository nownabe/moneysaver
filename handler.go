package main

import (
	"fmt"
	"net/http"
)

type handler struct {
	eventProcessor *eventProcessor
}

func (h *handler) handleEvents(w http.ResponseWriter, r *http.Request) {
	if contentType := r.Header.Get("Content-Type"); contentType != "application/json" {
		logger.Printf("Unsupported content type: %s", contentType)
		w.WriteHeader(http.StatusUnsupportedMediaType)

		return
	}

	resBody, err := h.eventProcessor.process(r)
	if err != nil {
		logger.Printf("h.eventProcessor.process: %v", err)
		writeErrorHeader(w, err)

		return
	}

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, resBody)
}
