package main

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
)

func Test_handler_challenge(t *testing.T) {
	body := bytes.NewBufferString(`{"challenge":"challengetoken"}`)
	req := httptest.NewRequest("POST", "/", body)
	req.Header.Add("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	handler(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status code should be 200, but %d", rec.Code)
	}

	respBody := rec.Body.String()
	if respBody != "challengetoken" {
		t.Errorf("response should be 'challengetoken', but %s", respBody)
	}
}

func Test_handler_else(t *testing.T) {
	cfg = &config{}
	body := bytes.NewBufferString(`{"event":{"text":"not number","ts":"1.23"}}`)
	req := httptest.NewRequest("POST", "/", body)
	req.Header.Add("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	handler(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Errorf("status code should be 204, but %d", rec.Code)
	}
}
