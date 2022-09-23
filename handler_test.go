package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/nownabe/moneysaver/slack"
)

func Test_handler_challenge(t *testing.T) {
	t.Parallel()

	ep := &eventProcessor{
		slack: newSlackMock(),
	}
	h := &handler{ep, nil}

	body := bytes.NewBufferString(`{"challenge":"challengetoken","type":"url_verification"}`)
	req := httptest.NewRequest(http.MethodPost, "/", body)
	req.Header.Add("Content-Type", "application/json")

	rec := httptest.NewRecorder()

	h.handleEvents(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status code should be 200, but %d", rec.Code)
	}

	contentType := rec.Header().Get("Content-Type")
	if contentType != "text/plain" {
		t.Errorf("Content-Type header should be 'text/plain', but '%s'", contentType)
	}

	if respBody := rec.Body.String(); respBody != `challengetoken` {
		t.Errorf(`response should be 'challengetoken', but '%s'`, respBody)
	}
}

func assertReqs(t *testing.T, expect, actual []*slack.ChatPostMessageReq) {
	t.Helper()

	expectJSON, err := json.Marshal(expect)
	if err != nil {
		panic(err)
	}

	actualJSON, err := json.Marshal(actual)
	if err != nil {
		panic(err)
	}

	if string(expectJSON) != string(actualJSON) {
		t.Errorf("incorrect chat.postMessage requests:\n  expect: %s\n  actual: %s", expectJSON, actualJSON)
	}
}

func Test_event_handler(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		requestBody string
		code        int
		reqs        []*slack.ChatPostMessageReq
	}{
		"not number": {
			`{"token":"valid","event":{"channel":"ch1","text":"not number","ts":"1.23"}}`,
			http.StatusOK,
			nil,
		},
		"no limit": {
			`{"token":"valid","event":{"channel":"unknown-ch","text":"123","ts":"1.23"}}`,
			http.StatusOK,
			[]*slack.ChatPostMessageReq{},
		},
	}

	for name, c := range cases {
		c := c

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			mock := newSlackMock()

			fs := getFirestoreClient(t)
			ep := &eventProcessor{
				slack:           mock,
				channelRepo:     &channelRepo{fs},
				expenditureRepo: &expenditureRepo{fs},
			}

			h := &handler{ep, nil}
			defer flushStore(t)

			body := bytes.NewBufferString(c.requestBody)
			req := httptest.NewRequest(http.MethodPost, "/", body)
			req.Header.Add("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			h.handleEvents(rec, req)

			if rec.Code != c.code {
				t.Errorf("status code should be %d, but %d", c.code, rec.Code)
			}

			if c.reqs != nil {
				m, _ := mock.(*slackMock)
				assertReqs(t, c.reqs, m.requests())
			}
		})
	}
}
