package app

import (
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/airconduct/kuilei/pkg/pluginhelpers"
	"github.com/airconduct/kuilei/pkg/plugins"
	"github.com/airconduct/kuilei/pkg/probot"
	"github.com/airconduct/kuilei/pkg/probot/github"
	"github.com/airconduct/kuilei/pkg/signals"

	_ "github.com/airconduct/kuilei/pkg/plugins/factory"
)

func NewHook() *cobra.Command {
	opts := &hookOptions{
		githubApp: probot.NewGithubAPP(),

		pluginConfigCache: pluginhelpers.NewPluginConfigCache(),
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
	githubApp probot.App[probot.GithubClient]

	configPath        string
	pluginConfigCache *pluginhelpers.PluginConfigCache
}

func (opts *hookOptions) AddFlags(flags *pflag.FlagSet) {
	opts.githubApp.AddFlags(flags)
	flags.StringVar(&opts.configPath, "config-path", ".github/kuilei.yml", "config path for kuilei App in git repo")
}

func (opts *hookOptions) Validate(args []string) error {
	opts.githubApp.On(github.Event.Issues).
		WithHandler(github.IssuesHandler(func(ctx github.IssuesContext) {
			payload := ctx.Payload()
			pluginClient := pluginhelpers.PluginConfigClientFromGithub(
				ctx.Client(), opts.configPath, opts.pluginConfigCache,
			)
			cfg, err := pluginClient.GetConfig(payload.Repo.Owner.GetLogin(), payload.Repo.GetName())
			ctx.Must(err)

			for _, p := range cfg.Plugins {
				plugin := plugins.GetGitCommentPlugin(p.Name, plugins.ClientSets{
					GitIssueClient:     pluginhelpers.GitIssueClientFromGithub(ctx.Client()),
					PluginConfigClient: pluginClient,
				}, p.Args...)
				if err := plugin.Do(ctx, pluginhelpers.GitCommentEventFromGithubIssuesEvent(payload)); err != nil {
					ctx.Logger().Error(err, "Failed to execute plugin", "name", plugin.Name())
				}
			}
		}))
	opts.githubApp.On(github.Event.IssueComment).
		WithHandler(github.IssueCommentHandler(func(ctx github.IssueCommentContext) {
			payload := ctx.Payload()
			pluginClient := pluginhelpers.PluginConfigClientFromGithub(
				ctx.Client(), opts.configPath, opts.pluginConfigCache,
			)
			cfg, err := pluginClient.GetConfig(payload.Repo.Owner.GetLogin(), payload.Repo.GetName())
			ctx.Must(err)

			for _, p := range cfg.Plugins {
				plugin := plugins.GetGitCommentPlugin(p.Name, plugins.ClientSets{
					GitIssueClient:     pluginhelpers.GitIssueClientFromGithub(ctx.Client()),
					PluginConfigClient: pluginClient,
				}, p.Args...)
				if err := plugin.Do(ctx, pluginhelpers.GitCommentEventFromGithubIssueCommentEvent(payload)); err != nil {
					ctx.Logger().Error(err, "Failed to execute plugin", "name", plugin.Name())
				}
			}
		}))
	opts.githubApp.On(github.Event.PullRequest).
		WithHandler(github.PullRequestHandler(func(ctx github.PullRequestContext) {
			payload := ctx.Payload()
			pluginClient := pluginhelpers.PluginConfigClientFromGithub(
				ctx.Client(), opts.configPath, opts.pluginConfigCache,
			)
			cfg, err := pluginClient.GetConfig(payload.Repo.Owner.GetLogin(), payload.Repo.GetName())
			ctx.Must(err)

			for _, p := range cfg.Plugins {
				plugin := plugins.GetGitCommentPlugin(p.Name, plugins.ClientSets{
					GitIssueClient:     pluginhelpers.GitIssueClientFromGithub(ctx.Client()),
					PluginConfigClient: pluginClient,
				}, p.Args...)
				if err := plugin.Do(ctx, pluginhelpers.GitCommentEventFromGithubPullRequestEvent(payload)); err != nil {
					ctx.Logger().Error(err, "Failed to execute plugin", "name", plugin.Name())
				}
			}
		}))
	opts.githubApp.On(github.Event.PullRequestReview).
		WithHandler(github.PullRequestReviewHandler(func(ctx github.PullRequestReviewContext) {
			payload := ctx.Payload()
			pluginClient := pluginhelpers.PluginConfigClientFromGithub(
				ctx.Client(), opts.configPath, opts.pluginConfigCache,
			)
			cfg, err := pluginClient.GetConfig(payload.Repo.Owner.GetLogin(), payload.Repo.GetName())
			ctx.Must(err)

			for _, p := range cfg.Plugins {
				plugin := plugins.GetGitCommentPlugin(p.Name, plugins.ClientSets{
					GitIssueClient:     pluginhelpers.GitIssueClientFromGithub(ctx.Client()),
					PluginConfigClient: pluginClient,
				}, p.Args...)
				if err := plugin.Do(ctx, pluginhelpers.GitCommentEventFromGithubPullRequestReviewEvent(payload)); err != nil {
					ctx.Logger().Error(err, "Failed to execute plugin", "name", plugin.Name())
				}
			}
		}))
	opts.githubApp.On(github.Event.PullRequestReviewComment).
		WithHandler(github.PullRequestReviewCommentHandler(func(ctx github.PullRequestReviewCommentContext) {
			payload := ctx.Payload()
			pluginClient := pluginhelpers.PluginConfigClientFromGithub(
				ctx.Client(), opts.configPath, opts.pluginConfigCache,
			)
			cfg, err := pluginClient.GetConfig(payload.Repo.Owner.GetLogin(), payload.Repo.GetName())
			ctx.Must(err)

			for _, p := range cfg.Plugins {
				plugin := plugins.GetGitCommentPlugin(p.Name, plugins.ClientSets{
					GitIssueClient:     pluginhelpers.GitIssueClientFromGithub(ctx.Client()),
					PluginConfigClient: pluginClient,
				}, p.Args...)
				if err := plugin.Do(ctx, pluginhelpers.GitCommentEventFromGithubPullRequestReviewCommentEvent(payload)); err != nil {
					ctx.Logger().Error(err, "Failed to execute plugin", "name", plugin.Name())
				}
			}
		}))
	opts.githubApp.On(github.Event.Push).
		WithHandler(github.PullRequestHandler(func(ctx github.PullRequestContext) {}))
	opts.githubApp.On(github.Event.Status).
		WithHandler(github.PullRequestHandler(func(ctx github.PullRequestContext) {}))
	return nil
}

func (opts *hookOptions) Run() error {
	ctx := signals.SetupSignalContext()
	return opts.githubApp.Run(ctx)
}
