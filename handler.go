package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/nownabe/moneysaver/slack"
	"golang.org/x/xerrors"
)

type handler struct {
	cfg   *config
	store *storeClient
	slack slack.Client
}

func (h *handler) handleEvents(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		logger.Printf("failed to read request body: %v", err)
		w.WriteHeader(http.StatusConflict)

		return
	}
	defer r.Body.Close()

	logger.Printf("%s", body)

	if contentType := r.Header.Get("Content-Type"); contentType != "application/json" {
		logger.Printf("Unsupported content type: %s", contentType)
		w.WriteHeader(http.StatusUnsupportedMediaType)

		return
	}

	msg, err := newSlackMessage(body)
	if err != nil {
		logger.Printf("unexpexted request body: %v", err)
		w.WriteHeader(http.StatusBadRequest)

		return
	}

	// Challenge request
	// https://api.slack.com/apis/connections/events-api#the-events-api__subscribing-to-event-types__events-api-request-urls__request-url-configuration--verification
	if msg.isChallenge() {
		w.WriteHeader(http.StatusOK)
		w.Header().Add("Content-Type", "application/json")
		fmt.Fprintf(w, `{"challenge":"%s"}`, msg.Challenge)

		return
	}

	if msg.Token != h.cfg.SlackVerificationToken {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// Messages not to be processed
	if !msg.ok {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	// Messages in channels not to be processed
	if _, ok := h.cfg.getLimit(msg.Event.Channel); !ok {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	if err := h.store.add(ctx, msg); err != nil {
		e := xerrors.Errorf("failed to add record: %w", err)
		logger.Printf("%v", e)

		if err := h.replyError(ctx, msg, e); err != nil {
			logger.Printf("failed to reply error: %v", err)
		}

		w.WriteHeader(http.StatusInternalServerError)

		return
	}

	total, err := h.store.total(ctx, msg)
	if err != nil {
		e := xerrors.Errorf("failed to aggregate: %w", err)
		logger.Printf("%v", e)

		if err := h.replyError(ctx, msg, e); err != nil {
			logger.Printf("failed to reply error: %v", err)
		}

		w.WriteHeader(http.StatusInternalServerError)

		return
	}

	if err := h.replySuccess(ctx, msg, total); err != nil {
		logger.Printf("failed to reply: %v", err)
		w.WriteHeader(http.StatusInternalServerError)

		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *handler) replySuccess(ctx context.Context, msg *slackMessage, total int64) error {
	limit, ok := h.cfg.getLimit(msg.Event.Channel)
	if !ok {
		return xerrors.Errorf("limit is not configured: %s", msg.Event.Channel)
	}

	r := &slack.ChatPostMessageReq{
		Channel:   msg.Event.Channel,
		Text:      "カード利用を登録しました",
		Username:  "MoneySaver",
		IconEmoji: ":money_with_wings:",
		Attachments: []*slack.Attachment{
			{
				Fields: []*slack.AttachmentField{
					{
						Title: "利用額",
						Value: humanize(msg.amount),
						Short: true,
					},
					{
						Title: "今月の利用可能残額",
						Value: humanize(limit - total),
						Short: true,
					},
					{
						Title: "今月の合計利用額",
						Value: humanize(total),
						Short: true,
					},
					{
						Title: "今月の設定上限額",
						Value: humanize(limit),
						Short: true,
					},
				},
			},
		},
	}

	return h.slack.ChatPostMessage(ctx, r)
}

func (h *handler) replyError(ctx context.Context, msg *slackMessage, err error) error {
	r := &slack.ChatPostMessageReq{
		Channel:   msg.Event.Channel,
		Text:      "```\n" + err.Error() + "\n```",
		Username:  "MoneySaver",
		IconEmoji: ":money_with_wings:",
	}

	return h.slack.ChatPostMessage(ctx, r)
}
