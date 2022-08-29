package main

import (
	"strconv"

	"github.com/slack-go/slack/slackevents"
)

// https://firebase.google.com/docs/firestore/manage-data/data-types?hl=ja#data_types
type channel struct {
	ID     string `firestore:"-"`
	Budget int64  `firestore:"budget"`
}

type expenditure struct {
	ts     string // Slack timestamp which is used to identify the message
	amount int64
}

func newExpenditure(ev *slackevents.MessageEvent) (*expenditure, bool) {
	a, err := strconv.ParseInt(ev.Text, 10, 64)
	if err != nil {
		return nil, false
	}

	return &expenditure{ev.TimeStamp, a}, true
}
