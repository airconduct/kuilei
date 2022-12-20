package main

import (
	"os"

	"github.com/spf13/pflag"

	"github.com/airconduct/kuilei/cmd/kuilei/app"
)

func main() {
	flags := pflag.NewFlagSet("kuilei", pflag.PanicOnError)
	pflag.CommandLine = flags

	if err := app.New().Execute(); err != nil {
		os.Exit(1)
	}
}
