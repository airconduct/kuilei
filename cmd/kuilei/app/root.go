package app

import (
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func New() *cobra.Command {
	opts := &rootOptions{}
	cmd := &cobra.Command{
		Use: "kuilei",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}
	opts.AddFlags(pflag.CommandLine)
	cmd.AddCommand(
		NewHook(),
	)
	return cmd
}

type rootOptions struct{}

func (opts *rootOptions) AddFlags(flags *pflag.FlagSet) {}

func (opts *rootOptions) Validate(args []string) {}

func (opts *rootOptions) Run() error {
	return nil
}
