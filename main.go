package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"golang.org/x/xerrors"
)

var (
	cfg *config
	sc  *slackClient
)

func init() {
	c, err := newConfig()
	if err != nil {
		panic(err)
	}
	cfg = c

	sc = newSlackClient(cfg.SlackBotToken)
}

func main() {
	http.HandleFunc("/", handler)
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("failed to listen and serve: %v", err)
	}
}

func handler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

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

	log.Printf("%s", body)

	msg, err := newSlackMessage(body)
	if err != nil {
		log.Printf("unexpexted request body: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Challenge request
	// https://api.slack.com/events-api#the-events-api__subscribing-to-event-types__events-api-request-urls__request-url-configuration--verification
	if msg.isChallenge() {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, msg.Challenge)
		return
	}

	// Messages not to be processed
	if !msg.ok {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	if err := addRecord(ctx, msg); err != nil {
		e := xerrors.Errorf("failed to add record: %w", err)
		log.Printf("%v", e)
		if err := replyError(ctx, msg, e); err != nil {
			log.Printf("failed to reply error: %v", err)
		}
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	total, err := aggregate(ctx, msg)
	if err != nil {
		e := xerrors.Errorf("failed to aggregate: %w", err)
		log.Printf("%v", e)
		if err := replyError(ctx, msg, e); err != nil {
			log.Printf("failed to reply error: %v", err)
		}
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := replySuccess(ctx, msg, total); err != nil {
		log.Printf("failed to reply: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func replySuccess(ctx context.Context, msg *slackMessage, total int64) error {
	limit, ok := cfg.getLimit(msg.Event.Channel)
	if !ok {
		return xerrors.Errorf("limit is not configured: %s", msg.Event.Channel)
	}

	r := &slackChatPostMessageReq{
		Channel:  msg.Event.Channel,
		Text:     "カード利用を登録しました",
		Username: "MoneySaver",
		Attachments: []slackAttachment{
			slackAttachment{
				Fields: []slackAttachmentField{
					slackAttachmentField{
						Title: "利用額",
						Value: humanize(msg.amount),
						Short: true,
					},
					slackAttachmentField{
						Title: "今月の利用可能残額",
						Value: humanize(limit - total),
						Short: true,
					},
					slackAttachmentField{
						Title: "今月の合計利用額",
						Value: humanize(total),
						Short: true,
					},
					slackAttachmentField{
						Title: "今月の設定上限額",
						Value: humanize(limit),
						Short: true,
					},
				},
			},
		},
	}

	return sc.chatPostMessage(ctx, r)
}

func replyError(ctx context.Context, msg *slackMessage, err error) error {
	r := &slackChatPostMessageReq{
		Channel:  msg.Event.Channel,
		Text:     err.Error(),
		Username: "MoneySaver",
	}

	return sc.chatPostMessage(ctx, r)
}

func humanize(n int64) string {
	s := fmt.Sprint(n)
	l := (len(s) + 3 - 1) / 3
	parts := make([]string, l)

	for i := 0; i < l; i++ {
		start := len(s) - (l-i)*3
		end := start + 3
		if start < 0 {
			start = 0
		}
		parts[i] = s[start:end]
	}
	return "¥" + strings.Join(parts, ",")
}
