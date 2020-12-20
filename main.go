package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

type SlackEvent struct {
	Challenge string `json:"challenge"`
}

func main() {
	http.HandleFunc("/", handler)
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("failed to listen and serve: %v", err)
	}
}

func handler(w http.ResponseWriter, r *http.Request) {
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

	var event SlackEvent
	if err := json.Unmarshal(body, &event); err != nil {
		log.Printf("failed to unmarshal request body: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if event.Challenge != "" {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"challenge":"%s"}`, event.Challenge)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
