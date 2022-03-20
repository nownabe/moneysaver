package main

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
)

func Test_handler_challenge(t *testing.T) {
	h := &handler{
		cfg:   &config{},
		store: nil,
		slack: newSlackMock(),
	}

	body := bytes.NewBufferString(`{"challenge":"challengetoken"}`)
	req := httptest.NewRequest("POST", "/", body)
	req.Header.Add("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status code should be 200, but %d", rec.Code)
	}

	contentType := rec.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Content-Type header should be 'application/json', but %s", contentType)
	}

	respBody := rec.Body.String()
	if respBody != `{"challenge":"challengetoken"}` {
		t.Errorf(`response should be '{"challenge":"challengetoken"}', but %s`, respBody)
	}
}

func Test_handler_else(t *testing.T) {
	h := &handler{
		cfg:   &config{},
		store: nil,
		slack: newSlackMock(),
	}

	body := bytes.NewBufferString(`{"event":{"text":"not number","ts":"1.23"}}`)
	req := httptest.NewRequest("POST", "/", body)
	req.Header.Add("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Errorf("status code should be 204, but %d", rec.Code)
	}
}
