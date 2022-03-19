package slack

import (
	"context"
	"net/http"
)

// Client is an interface of Slack Client.
type Client interface {
	ChatPostMessage(context.Context, *ChatPostMessageReq) error
}

type client struct {
	token  string
	client *http.Client
}

// New builds a new slack client.
func New(token string) Client {
	return &client{
		token:  token,
		client: &http.Client{},
	}
}
