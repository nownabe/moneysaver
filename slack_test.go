package main

import (
	"context"

	"github.com/nownabe/moneysaver/slack"
)

type slackMock struct {
	recorder []*slack.ChatPostMessageReq
}

func newSlackMock() slack.Client {
	return &slackMock{
		recorder: []*slack.ChatPostMessageReq{},
	}
}

func (c *slackMock) ChatPostMessage(ctx context.Context, r *slack.ChatPostMessageReq) error {
	c.recorder = append(c.recorder, r)
	return nil
}

func (c *slackMock) requests() []*slack.ChatPostMessageReq {
	return c.recorder
}
