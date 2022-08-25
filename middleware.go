package main

import (
	"bytes"
	"io/ioutil"
	"net/http"

	"github.com/slack-go/slack"
)

func slackVerifier(signingSecret string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			body, err := ioutil.ReadAll(r.Body)
			if err != nil {
				logger.Printf("ioutil.ReadAll: %v", err)
				w.WriteHeader(http.StatusBadRequest)

				return
			}

			if err := r.Body.Close(); err != nil {
				logger.Printf("r.Body.Close: %v", err)
				w.WriteHeader(http.StatusInternalServerError)

				return
			}

			r.Body = ioutil.NopCloser(bytes.NewBuffer(body))

			verifier, err := slack.NewSecretsVerifier(r.Header, signingSecret)
			if err != nil {
				logger.Printf("slack.NewSecretsVerifier: %v", err)
				w.WriteHeader(http.StatusBadRequest)

				return
			}

			if _, err := verifier.Write(body); err != nil {
				logger.Printf("verifier.Write: %v", err)
				w.WriteHeader(http.StatusInternalServerError)

				return
			}

			if err := verifier.Ensure(); err != nil {
				w.WriteHeader(http.StatusBadRequest)

				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
