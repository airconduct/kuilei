package app

import (
	"github.com/airconduct/go-probot"
	"github.com/spf13/pflag"
)

type Builder[GT probot.GitClientType] interface {
	BindFlags(flags *pflag.FlagSet)
	Build() (probot.App[GT], error)
}
