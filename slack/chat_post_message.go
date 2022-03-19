package slack

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"

	"golang.org/x/xerrors"
)

// Attachment is an attachment.
type Attachment struct {
	Fields []*AttachmentField `json:"fields"`
}

// AttachmentField is an attachment field.
type AttachmentField struct {
	Title string `json:"title"`
	Value string `json:"value"`
	Short bool   `json:"short"`
}

// ChatPostMessageReq is a request for chat.postMessage method.
// https://api.slack.com/methods/chat.postMessage
type ChatPostMessageReq struct {
	Channel     string        `json:"channel"`
	IconEmoji   string        `json:"icon_emoji,omitempty"`
	Text        string        `json:"text"`
	Username    string        `json:"username,omitempty"`
	Attachments []*Attachment `json:"attachments"`
}

type chatPostMessageRes struct {
	OK    bool   `json:"ok"`
	Error string `json:"error"`
}

func (c *client) ChatPostMessage(ctx context.Context, r *ChatPostMessageReq) error {
	reqBody, err := json.Marshal(r)
	if err != nil {
		return xerrors.Errorf("failed to marshal slack chat.postMessage request: %w", err)
	}

	req, err := http.NewRequest("POST", "https://slack.com/api/chat.postMessage", bytes.NewReader(reqBody))
	if err != nil {
		return xerrors.Errorf("failed to build http request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.token)

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

	var sres chatPostMessageRes
	if err := json.Unmarshal(body, &sres); err != nil {
		return xerrors.Errorf("failed to unmarshal response body: %w", err)
	}

	if !sres.OK {
		return xerrors.Errorf("slack chat.postMessageMessage returned an error: %s", sres.Error)
	}

	return nil
}
