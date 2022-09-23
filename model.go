package main

import (
	"fmt"
	"strconv"
	"time"

	"github.com/slack-go/slack/slackevents"
)

var errNotExpenditureMessage = fmt.Errorf("not expenditure message")

// https://firebase.google.com/docs/firestore/manage-data/data-types?hl=ja#data_types
type channel struct {
	ID     string `firestore:"-"`
	Budget int64  `firestore:"budget"`
}

type expenditure struct {
	Channel string `firestore:"-"`
	// Slack timestamp which is used to identify the message
	TS        string    `firestore:"-"`
	Amount    int64     `firestore:"amount"`
	Timestamp time.Time `firestore:"timestamp"`
}

func newExpenditure(ev *slackevents.MessageEvent) (*expenditure, error) {
	a, err := strconv.ParseInt(ev.Text, 10, 64)
	if err != nil {
		return nil, errNotExpenditureMessage
	}

	ut, err := strconv.ParseFloat(ev.TimeStamp, 64)
	if err != nil {
		return nil, fmt.Errorf("strconv.ParseFloat: %w", err)
	}

	return &expenditure{
		Channel:   ev.Channel,
		TS:        ev.TimeStamp,
		Amount:    a,
		Timestamp: time.Unix(int64(ut), 0),
	}, nil
}

func newExpenditureFromPreviousMessage(ev *slackevents.MessageEvent) (*expenditure, error) {
	ex, err := newExpenditure(ev.PreviousMessage)
	if err != nil {
		return nil, fmt.Errorf("newExpenditure: %w", err)
	}

	ex.Channel = ev.Channel

	return ex, nil
}
