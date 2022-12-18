package probot

import (
	"context"

	"github.com/go-logr/logr"
)

func newProbotContext[GT GitClientType, PT gitEventType](
	parent context.Context, logger logr.Logger,
	client *GT, payload *PT,
) *probotContext[GT, PT] {
	return &probotContext[GT, PT]{
		Context: parent,
		logger:  logger,
		client:  client,
		payload: payload,
	}
}

type probotContext[GT GitClientType, PT gitEventType] struct {
	context.Context
	// Input
	logger  logr.Logger
	client  *GT
	payload *PT
}

func (ctx *probotContext[GT, PT]) Payload() *PT {
	return ctx.payload
}
func (ctx *probotContext[GT, PT]) Client() *GT {
	return ctx.client
}

func (ctx *probotContext[GT, PT]) Logger() logr.Logger {
	return ctx.logger
}

func (ctx *probotContext[GT, PT]) Must(errors ...interface{}) {
	for _, v := range errors {
		if v != nil {
			if err, ok := v.(error); ok {
				panic(err)
			}
		}
	}
}
