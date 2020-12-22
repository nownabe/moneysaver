package main

import (
	"encoding/json"
	"strconv"
	"time"

	"golang.org/x/xerrors"
)

type slackEvent struct {
	Channel string `json:"channel"`
	Text    string `json:"text"`
	TS      string `json:"ts"`
}

// https://api.slack.com/events-api#the-events-api__receiving-events__callback-field-overview
type slackMessage struct {
	Token     string     `json:"token"`
	Challenge string     `json:"challenge"`
	TeamID    string     `json:"team_id"`
	Event     slackEvent `json:"event"`

	ok        bool
	amount    int64
	timestamp time.Time
}

func newSlackMessage(b []byte) (*slackMessage, error) {
	var msg slackMessage

	if err := json.Unmarshal(b, &msg); err != nil {
		return nil, xerrors.Errorf("slack message parse error: %w", err)
	}

	ts, err := strconv.ParseFloat(msg.Event.TS, 64)
	if err != nil {
		return nil, xerrors.Errorf("parse error event.ts: %w", err)
	}
	msg.timestamp = time.Unix(int64(ts), 0)

	a, err := strconv.ParseInt(msg.Event.Text, 10, 64)
	if err != nil {
		return &msg, nil
	}
	msg.amount = a
	msg.ok = true

	return &msg, nil
}

func (msg *slackMessage) isChallenge() bool {
	return msg.Challenge != ""
}

func (msg *slackMessage) month() string {
	return msg.timestamp.Format("2006-01")
}
