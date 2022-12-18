package probot

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/spf13/pflag"
)

type App[GT GitClientType] interface {
	AddFlags(flags *pflag.FlagSet)
	On(events ...WebhookEvent) handleWith
	Run(ctx context.Context) error
}

type ProbotContext[GT GitClientType, PT gitEventType] interface {
	context.Context

	Payload() *PT
	Client() *GT
	Logger() logr.Logger
	Must(...interface{})
}
