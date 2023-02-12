package app

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/airconduct/go-probot"
	"github.com/airconduct/kuilei/pkg/app"
	"github.com/airconduct/kuilei/pkg/signals"

	"github.com/airconduct/kuilei/pkg/app/github"
	_ "github.com/airconduct/kuilei/pkg/plugins/factory"
)

func NewHook() *cobra.Command {
	opts := &hookOptions{
		githubAppBuilder: github.New(),
	}

	cmd := &cobra.Command{
		Use:   "hook start hook server",
		Short: "start hook server",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := opts.Validate(args); err != nil {
				return err
			}
			return opts.Run()
		},
	}

	opts.AddFlags(cmd.Flags())
	return cmd
}

type hookOptions struct {
	githubAppBuilder app.Builder[probot.GitHubClient]
}

func (opts *hookOptions) AddFlags(flags *pflag.FlagSet) {
	opts.githubAppBuilder.BindFlags(flags)
}

func (opts *hookOptions) Validate(args []string) error {
	// TODO: validate args
	return nil
}

func (opts *hookOptions) Run() error {
	ctx := signals.SetupSignalContext()
	githubApp, err := opts.githubAppBuilder.Build()
	if err != nil {
		return fmt.Errorf("faield to build github app: %w", err)
	}
	return githubApp.Run(ctx)
}
