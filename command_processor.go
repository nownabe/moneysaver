package main

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/slack-go/slack"
)

type commandProcessor struct {
	channelRepo *channelRepo
}

func (p *commandProcessor) process(ctx context.Context, c slack.SlashCommand) (*slack.Msg, error) {
	args := strings.Split(c.Text, " ")
	if len(args) != 2 || args[0] != "set" {
		return &slack.Msg{Text: "Invalid command format. Usage: `/moneysaver set 1000"}, nil
	}

	budget, err := strconv.ParseInt(args[1], 10, 64)
	if err != nil {
		return &slack.Msg{Text: "Budget must be an integer."}, nil
	}

	if err := p.setBudget(ctx, c.ChannelID, budget); err != nil {
		return nil, wrap(http.StatusInternalServerError, "p.setBudget: %w", err)
	}

	return &slack.Msg{Text: "Set budget to #" + c.ChannelName}, nil
}

func (p *commandProcessor) setBudget(ctx context.Context, chID string, budget int64) error {
	ch := &channel{
		ID:     chID,
		Budget: budget,
	}

	if err := p.channelRepo.save(ctx, ch); err != nil {
		return fmt.Errorf("p.channelRepo.save: %w", err)
	}

	return nil
}
