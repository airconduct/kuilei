package app

import (
	"github.com/go-logr/logr"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/airconduct/go-probot"
	"github.com/airconduct/kuilei/pkg/pluginhelpers"
	"github.com/airconduct/kuilei/pkg/plugins"
	"github.com/airconduct/kuilei/pkg/signals"

	_ "github.com/airconduct/kuilei/pkg/plugins/factory"
)

func NewHook() *cobra.Command {
	opts := &hookOptions{
		githubApp: probot.NewGitHubAPP(),

		pluginConfigCache: pluginhelpers.NewConfigCache[plugins.Configuration](),
		ownersConfigCache: pluginhelpers.NewConfigNearestCache[plugins.OwnersConfiguration](),
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
	githubApp probot.App[probot.GitHubClient]

	configPath        string
	ownersFile        string
	pluginConfigCache pluginhelpers.ConfigCache[plugins.Configuration]
	ownersConfigCache pluginhelpers.ConfigCache[plugins.OwnersConfiguration]
}

func (opts *hookOptions) AddFlags(flags *pflag.FlagSet) {
	opts.githubApp.AddFlags(flags)
	flags.StringVar(&opts.configPath, "config-path", ".github/kuilei.yml", "config path for kuilei App in git repo")
	flags.StringVar(&opts.ownersFile, "owners-file", "OWNERS", "owners file name")
}

func (opts *hookOptions) Validate(args []string) error {
	opts.githubApp.On(probot.GitHub.Issues).
		WithHandler(probot.GitHub.Issues.Handler(func(ctx probot.GitHubIssuesContext) {
			payload := ctx.Payload()
			pluginClient := pluginhelpers.PluginConfigClientFromGithub(
				ctx.Client(), opts.configPath, opts.pluginConfigCache,
			)
			cfg, err := pluginClient.GetConfig(payload.Repo.Owner.GetLogin(), payload.Repo.GetName())
			ctx.Must(err)

			for _, p := range cfg.Plugins {
				plugin := plugins.GetGitCommentPlugin(p.Name, getClientSets(opts, ctx, pluginClient), p.Args...)
				if err := plugin.Do(ctx, pluginhelpers.GitCommentEventFromGithubIssuesEvent(payload)); err != nil {
					ctx.Logger().Error(err, "Failed to execute plugin", "name", plugin.Name())
				}
			}
		}))
	opts.githubApp.On(probot.GitHub.IssueComment).
		WithHandler(probot.GitHub.IssueComment.Handler(func(ctx probot.GitHubIssueCommentContext) {
			payload := ctx.Payload()
			pluginClient := pluginhelpers.PluginConfigClientFromGithub(
				ctx.Client(), opts.configPath, opts.pluginConfigCache,
			)
			cfg, err := pluginClient.GetConfig(payload.Repo.Owner.GetLogin(), payload.Repo.GetName())
			ctx.Must(err)

			for _, p := range cfg.Plugins {
				plugin := plugins.GetGitCommentPlugin(p.Name, getClientSets(opts, ctx, pluginClient), p.Args...)
				if err := plugin.Do(ctx, pluginhelpers.GitCommentEventFromGithubIssueCommentEvent(payload)); err != nil {
					ctx.Logger().Error(err, "Failed to execute plugin", "name", plugin.Name())
				}
			}
		}))
	opts.githubApp.On(probot.GitHub.PullRequest).
		WithHandler(probot.GitHub.PullRequest.Handler(func(ctx probot.GitHubPullRequestContext) {
			payload := ctx.Payload()
			pluginClient := pluginhelpers.PluginConfigClientFromGithub(
				ctx.Client(), opts.configPath, opts.pluginConfigCache,
			)
			cfg, err := pluginClient.GetConfig(payload.Repo.Owner.GetLogin(), payload.Repo.GetName())
			ctx.Must(err)

			for _, p := range cfg.Plugins {
				plugin := plugins.GetGitCommentPlugin(p.Name, getClientSets(opts, ctx, pluginClient), p.Args...)
				if err := plugin.Do(ctx, pluginhelpers.GitCommentEventFromGithubPullRequestEvent(payload)); err != nil {
					ctx.Logger().Error(err, "Failed to execute plugin", "name", plugin.Name())
				}
			}
		}))
	opts.githubApp.On(probot.GitHub.PullRequestReview).
		WithHandler(probot.GitHub.PullRequestReview.Handler(func(ctx probot.GitHubPullRequestReviewContext) {
			payload := ctx.Payload()
			pluginClient := pluginhelpers.PluginConfigClientFromGithub(
				ctx.Client(), opts.configPath, opts.pluginConfigCache,
			)
			cfg, err := pluginClient.GetConfig(payload.Repo.Owner.GetLogin(), payload.Repo.GetName())
			ctx.Must(err)

			for _, p := range cfg.Plugins {
				plugin := plugins.GetGitCommentPlugin(p.Name, getClientSets(opts, ctx, pluginClient), p.Args...)
				if err := plugin.Do(ctx, pluginhelpers.GitCommentEventFromGithubPullRequestReviewEvent(payload)); err != nil {
					ctx.Logger().Error(err, "Failed to execute plugin", "name", plugin.Name())
				}
			}
		}))
	opts.githubApp.On(probot.GitHub.PullRequestReviewComment).
		WithHandler(probot.GitHub.PullRequestReviewComment.Handler(func(ctx probot.GitHubPullRequestReviewCommentContext) {
			payload := ctx.Payload()
			pluginClient := pluginhelpers.PluginConfigClientFromGithub(
				ctx.Client(), opts.configPath, opts.pluginConfigCache,
			)
			cfg, err := pluginClient.GetConfig(payload.Repo.Owner.GetLogin(), payload.Repo.GetName())
			ctx.Must(err)

			for _, p := range cfg.Plugins {
				plugin := plugins.GetGitCommentPlugin(p.Name, getClientSets(opts, ctx, pluginClient), p.Args...)
				if err := plugin.Do(ctx, pluginhelpers.GitCommentEventFromGithubPullRequestReviewCommentEvent(payload)); err != nil {
					ctx.Logger().Error(err, "Failed to execute plugin", "name", plugin.Name())
				}
			}
		}))
	opts.githubApp.On(probot.GitHub.Push).
		WithHandler(probot.GitHub.Push.Handler(func(ctx probot.GitHubPushContext) {}))
	opts.githubApp.On(probot.GitHub.Status).
		WithHandler(probot.GitHub.Status.Handler(func(ctx probot.GitHubStatusContext) {}))
	return nil
}

func (opts *hookOptions) Run() error {
	ctx := signals.SetupSignalContext()
	return opts.githubApp.Run(ctx)
}

func getClientSets[PT any](
	opts *hookOptions,
	ctx probot.ProbotContext[probot.GitHubClient, PT], pluginClient plugins.PluginConfigClient,
) plugins.ClientSets {
	logger := ctx.Logger()
	return plugins.ClientSets{
		GitIssueClient:     pluginhelpers.GitIssueClientFromGithub(ctx.Client()),
		GitPRClient:        pluginhelpers.GitPRClientFromGithub(ctx.Client()),
		PluginConfigClient: pluginClient,
		OwnersClient:       pluginhelpers.OwnersClientFromGithub(ctx.Client(), opts.ownersFile, opts.ownersConfigCache),
		GitRepoClient:      pluginhelpers.GitRepoClientFromGithub(ctx.Client()),
		GitSearchClient:    pluginhelpers.GitSearchClientFromGithub(ctx.GraphQL()),
		LoggerClient: pluginhelpers.MakeLoggerClient(func() logr.Logger {
			return logger
		}),
	}
}
