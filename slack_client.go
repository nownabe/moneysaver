package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"

	"golang.org/x/xerrors"
)

type slackClient struct {
	token  string
	client *http.Client
}

type slackAttachment struct {
	Fields []slackAttachmentField `json:"fields"`
}

type slackAttachmentField struct {
	Title string `json:"title"`
	Value string `json:"value"`
	Short bool   `json:"short"`
}

type slackAPIResponse struct {
	OK    bool   `json:"ok"`
	Error string `json:"error"`
}

func newSlackClient(token string) *slackClient {
	return &slackClient{
		token:  token,
		client: &http.Client{},
	}
}

type slackChatPostMessageReq struct {
	Channel     string            `json:"channel"`
	IconEmoji   string            `json:"icon_emoji,omitempty"`
	Text        string            `json:"text"`
	Username    string            `json:"username,omitempty"`
	Attachments []slackAttachment `json:"attachments"`
}

func (c *slackClient) chatPostMessage(ctx context.Context, r *slackChatPostMessageReq) error {
	reqBody, err := json.Marshal(r)
	if err != nil {
		return xerrors.Errorf("failed to marshal slack chat.postMessage request: %w", err)
	}

	req, err := http.NewRequest("POST", "https://slack.com/api/chat.postMessage", bytes.NewReader(reqBody))
	if err != nil {
		return xerrors.Errorf("failed to build http request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+cfg.SlackBotToken)

	resp, err := c.client.Do(req.WithContext(ctx))
	if err != nil {
		return xerrors.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return xerrors.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode >= http.StatusBadRequest {
		return xerrors.Errorf(
			"slack chat.postMessage failed with status code %d (%s)", resp.StatusCode, body)
	}

	var sres slackAPIResponse
	if err := json.Unmarshal(body, &sres); err != nil {
		return xerrors.Errorf("failed to unmarshal response body: %w", err)
	}

	if !sres.OK {
		return xerrors.Errorf("slack chat.postMessageMessage returned an error: %s", sres.Error)
	}

	return nil
}
