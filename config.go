package main

import (
	"github.com/kelseyhightower/envconfig"
	"golang.org/x/xerrors"
)

type config struct {
	ProjectID          string `required:"true" split_words:"true"`
	SlackBotToken      string `required:"true" split_words:"true"`
	SlackSigningSecret string `required:"true" split_words:"true"`
}

func newConfig() (*config, error) {
	var c config
	if err := envconfig.Process("", &c); err != nil {
		return nil, xerrors.Errorf("failed to process config: %w", err)
	}
	return &c, nil
}
