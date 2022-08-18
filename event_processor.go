package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/nownabe/moneysaver/slack"
)

type eventProcessor struct {
	cfg   *config
	store *storeClient
	slack slack.Client
}

// process returns response body and error.
func (p *eventProcessor) process(r *http.Request) (string, error) {
	ctx := r.Context()

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return "", wrap(http.StatusConflict, "ioutil.ReadAll: %w", err)
	}
	defer r.Body.Close()

	logger.Printf("Event JSON: %s", body)

	msg, err := newSlackMessage(body)
	if err != nil {
		return "", wrap(http.StatusBadRequest, "invalid request body: %w", err)
	}

	// Challenge request
	// https://api.slack.com/apis/connections/events-api
	if msg.isChallenge() {
		return fmt.Sprintf(`{"challenge":"%s"}`, msg.Challenge), nil
	}

	if msg.Token != p.cfg.SlackVerificationToken {
		return "", e(http.StatusUnauthorized, "invalid token")
	}

	// Messages not to be processed
	if !msg.ok {
		return "", nil
	}

	// Messages in channels not to be processed
	if _, ok := p.cfg.getLimit(msg.Event.Channel); !ok {
		return "", nil
	}

	if err := p.store.add(ctx, msg); err != nil {
		err := wrap(http.StatusInternalServerError, "p.store.add: %w", err)

		if err := p.replyError(ctx, msg, err); err != nil {
			logger.Printf("failed to reply error: %v", err)
		}

		return "", err
	}

	total, err := p.store.total(ctx, msg)
	if err != nil {
		err := wrap(http.StatusInternalServerError, "h.store.total: %w", err)

		if err := p.replyError(ctx, msg, err); err != nil {
			logger.Printf("failed to reply error: %v", err)
		}

		return "", err
	}

	if err := p.replySuccess(ctx, msg, total); err != nil {
		return "", wrap(http.StatusInternalServerError, "p.replySuccess: %w", err)
	}

	return "", nil
}

func (p *eventProcessor) replyError(ctx context.Context, msg *slackMessage, err error) error {
	r := &slack.ChatPostMessageReq{
		Channel:   msg.Event.Channel,
		Text:      "```\n" + err.Error() + "\n```",
		Username:  "MoneySaver",
		IconEmoji: ":money_with_wings:",
	}

	if err := p.slack.ChatPostMessage(ctx, r); err != nil {
		return fmt.Errorf("p.slack.ChatPostMessage: %w", err)
	}

	return nil
}

func (p *eventProcessor) replySuccess(ctx context.Context, msg *slackMessage, total int64) error {
	limit, ok := p.cfg.getLimit(msg.Event.Channel)
	if !ok {
		return fmt.Errorf("limit is not configured: %s", msg.Event.Channel)
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

	if err := p.slack.ChatPostMessage(ctx, r); err != nil {
		return fmt.Errorf("p.slack.ChatPostMessage: %w", err)
	}

	return nil
}
