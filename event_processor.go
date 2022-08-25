package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/nownabe/moneysaver/slack"
)

type slackEvent struct {
	Channel string `json:"channel"`
	Text    string `json:"text"`
	TS      string `json:"ts"`
}

func (e *slackEvent) expenditure() (*expenditure, bool) {
	a, err := strconv.ParseInt(e.Text, 10, 64)
	if err != nil {
		return nil, false
	}

	return &expenditure{e.TS, a}, true
}

func (e *slackEvent) timestamp() (time.Time, error) {
	ts, err := strconv.ParseFloat(e.TS, 64)
	if err != nil {
		return time.Time{}, fmt.Errorf("strconv.ParseFloat: %w", err)
	}

	return time.Unix(int64(ts), 0), nil
}

// https://api.slack.com/events-api#the-events-api__receiving-events__callback-field-overview
type slackMessage struct {
	Token     string      `json:"token"`
	Challenge string      `json:"challenge"`
	TeamID    string      `json:"team_id"`
	Event     *slackEvent `json:"event"`
}

func (msg *slackMessage) isChallenge() bool {
	return msg.Challenge != ""
}

type eventProcessor struct {
	cfg         *config
	store       *storeClient
	slack       slack.Client
	channelRepo *channelRepo
}

// process returns response body and error.
func (p *eventProcessor) process(ctx context.Context, msg *slackMessage) (string, error) {
	// Challenge request
	// https://api.slack.com/apis/connections/events-api
	if msg.isChallenge() {
		return fmt.Sprintf(`{"challenge":"%s"}`, msg.Challenge), nil
	}

	if err := p.processExpenditure(ctx, msg.Event); err != nil {
		return "", wrap(http.StatusInternalServerError, "p.processExpenditure: %w", err)
	}

	return "", nil
}

func (p *eventProcessor) processExpenditure(ctx context.Context, ev *slackEvent) error {
	ch, err := p.channelRepo.findByID(ctx, ev.Channel)
	if err != nil {
		if errors.Is(err, errNotFound) {
			return nil
		}

		return fmt.Errorf("p.channelRepo.findByID: %w", err)
	}

	ex, ok := ev.expenditure()
	if !ok {
		// Messages not to be processed
		return nil
	}

	ts, err := ev.timestamp()
	if err != nil {
		return fmt.Errorf("ev.timestamp: %w", err)
	}

	if err := p.store.add(ctx, ev.Channel, ts, ex); err != nil {
		err := fmt.Errorf("p.store.add: %w", err)

		if err := p.replyError(ctx, ev.Channel, err); err != nil {
			logger.Printf("failed to reply error: %v", err)
		}

		return err
	}

	total, err := p.store.total(ctx, ev.Channel, ts)
	if err != nil {
		err := fmt.Errorf("p.store.total: %w", err)

		if err := p.replyError(ctx, ev.Channel, err); err != nil {
			logger.Printf("failed to reply error: %v", err)
		}

		return err
	}

	if err := p.replySuccess(ctx, ev.Channel, ch.Budget, total, ex); err != nil {
		return fmt.Errorf("p.replySuccess: %w", err)
	}

	return nil
}

func (p *eventProcessor) replyError(ctx context.Context, channel string, err error) error {
	r := &slack.ChatPostMessageReq{
		Channel:   channel,
		Text:      "```\n" + err.Error() + "\n```",
		Username:  "MoneySaver",
		IconEmoji: ":money_with_wings:",
	}

	if err := p.slack.ChatPostMessage(ctx, r); err != nil {
		return fmt.Errorf("p.slack.ChatPostMessage: %w", err)
	}

	return nil
}

func (p *eventProcessor) replySuccess(ctx context.Context, channel string, limit, total int64, ex *expenditure) error {
	r := &slack.ChatPostMessageReq{
		Channel:   channel,
		Text:      "カード利用を登録しました",
		Username:  "MoneySaver",
		IconEmoji: ":money_with_wings:",
		Attachments: []*slack.Attachment{
			{
				Fields: []*slack.AttachmentField{
					{
						Title: "利用額",
						Value: humanize(ex.amount),
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
