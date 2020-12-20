package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"golang.org/x/xerrors"
)

var (
	token   string = os.Getenv("SLACK_BOT_TOKEN")
	channel string = os.Getenv("SLACK_CHANNEL")
)

type slackChatPostMessage struct {
	Channel     string            `json:"channel"`
	IconEmoji   string            `json:"icon_emoji,omitempty"`
	Text        string            `json:"text"`
	Username    string            `json:"username,omitempty"`
	Attachments []slackAttachment `json:"attachments"`
}

type slackAttachment struct {
	Fields []slackAttachmentField `json:"fields"`
}

type slackAttachmentField struct {
	Title string `json:"title"`
	Value string `json:"value"`
	Short bool   `json:"short"`
}

type slackResponse struct {
	OK    bool   `json:"ok"`
	Error string `json:"error"`
}

func reply(ctx context.Context, amount int) error {
	m := buildSlackChatPostMessage(amount)

	reqBody, err := json.Marshal(m)
	if err != nil {
		return xerrors.Errorf("failed to marshal slack chat.postMessage request: %w", err)
	}

	req, err := http.NewRequest("POST", "https://slack.com/api/chat.postMessage", bytes.NewReader(reqBody))
	if err != nil {
		return xerrors.Errorf("failed to build http request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{}

	resp, err := client.Do(req.WithContext(ctx))
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

	var sres slackResponse
	if err := json.Unmarshal(body, &sres); err != nil {
		return xerrors.Errorf("failed to unmarshal response body: %w", err)
	}

	if !sres.OK {
		return xerrors.Errorf("slack chat.postMessageMessage returned an error: %s", sres.Error)
	}

	return nil
}

func buildSlackChatPostMessage(amount int) slackChatPostMessage {
	return slackChatPostMessage{
		Channel:  channel,
		Text:     "カード利用を登録しました",
		Username: "MoneySaver",
		Attachments: []slackAttachment{
			slackAttachment{
				Fields: []slackAttachmentField{
					slackAttachmentField{
						Title: "利用額",
						Value: humanize(amount),
						Short: true,
					},
					slackAttachmentField{
						Title: "今月の利用可能残額",
						Value: "xxxx円",
						Short: true,
					},
					slackAttachmentField{
						Title: "今月の設定上限額",
						Value: "xxxx円",
						Short: true,
					},
					slackAttachmentField{
						Title: "今月の合計利用額",
						Value: "xxxx円",
						Short: true,
					},
				},
			},
		},
	}
}

func humanize(n int) string {
	s := fmt.Sprint(n)
	l := (len(s) + 3 - 1) / 3
	parts := make([]string, l)

	for i := 0; i < l; i++ {
		start := len(s) - (l-i)*3
		end := start + 3
		if start < 0 {
			start = 0
		}
		parts[i] = s[start:end]
	}
	return "¥" + strings.Join(parts, ",")
}
