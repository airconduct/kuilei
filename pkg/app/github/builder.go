package github

import (
	"github.com/airconduct/go-probot"
	"github.com/airconduct/kuilei/pkg/app"
	"github.com/airconduct/kuilei/pkg/pluginhelpers"
	"github.com/airconduct/kuilei/pkg/plugins"
	"github.com/go-logr/logr"
	"github.com/spf13/pflag"
)

func New() app.Builder[probot.GitHubClient] {
	return &githubAppBuilder{
		githubApp:         probot.NewGitHubAPP(),
		pluginConfigCache: pluginhelpers.NewConfigCache[plugins.Configuration](),
		ownersConfigCache: pluginhelpers.NewConfigNearestCache[plugins.OwnersConfiguration](),
	}
}

type githubAppBuilder struct {
	githubApp         probot.App[probot.GitHubClient]
	configPath        string
	ownersFile        string
	pluginConfigCache pluginhelpers.ConfigCache[plugins.Configuration]
	ownersConfigCache pluginhelpers.ConfigCache[plugins.OwnersConfiguration]
}

var _ app.Builder[probot.GitHubClient] = &githubAppBuilder{}

func (b *githubAppBuilder) BindFlags(flags *pflag.FlagSet) {
	b.githubApp.AddFlags(flags)
	flags.StringVar(&b.configPath, "config-path", ".github/kuilei.yml", "config path for kuilei App in git repo")
	flags.StringVar(&b.ownersFile, "owners-file", "OWNERS", "owners file name")
}
func (b *githubAppBuilder) Build() (probot.App[probot.GitHubClient], error) {
	return b.complete(
		b.githubApp, b.configPath, b.ownersFile,
		b.pluginConfigCache, b.ownersConfigCache,
	), nil
}

func (*githubAppBuilder) complete(
	githubApp probot.App[probot.GitHubClient],
	configPath string,
	ownersFile string,
	pluginConfigCache pluginhelpers.ConfigCache[plugins.Configuration],
	ownersConfigCache pluginhelpers.ConfigCache[plugins.OwnersConfiguration],
) probot.App[probot.GitHubClient] {
	// Listen for GitHub issues events
	githubApp.On(probot.GitHub.Issues).WithHandler(probot.GitHub.Issues.Handler(func(ctx probot.GitHubIssuesContext) {
		payload := ctx.Payload()
		// Get app config
		pluginClient := pluginhelpers.PluginConfigClientFromGithub(
			ctx.Client(), configPath, pluginConfigCache,
		)
		cfg, err := pluginClient.GetConfig(payload.Repo.Owner.GetLogin(), payload.Repo.GetName())
		ctx.Must(err)
		// Execute all plugins in config
		for _, p := range cfg.Plugins {
			plugin := plugins.GetGitCommentPlugin(p.Name, getClientSets(ownersFile, ownersConfigCache, ctx, pluginClient), p.Args...)
			if plugin == nil {
				ctx.Logger().Info("Plugin not found", "name", p.Name)
				continue
			}
			// Execute plugin
			if err := plugin.Do(ctx, pluginhelpers.GitCommentEventFromGithubIssuesEvent(payload)); err != nil {
				ctx.Logger().Error(err, "Failed to execute plugin", "name", plugin.Name())
			}
		}
	}))
	// Listen for GitHub issue comment events
	githubApp.On(probot.GitHub.IssueComment).WithHandler(probot.GitHub.IssueComment.Handler(func(ctx probot.GitHubIssueCommentContext) {
		payload := ctx.Payload()
		pluginClient := pluginhelpers.PluginConfigClientFromGithub(
			ctx.Client(), configPath, pluginConfigCache,
		)
		// Get app config
		cfg, err := pluginClient.GetConfig(payload.Repo.Owner.GetLogin(), payload.Repo.GetName())
		ctx.Must(err)
		// Execute all plugins in config
		for _, p := range cfg.Plugins {
			plugin := plugins.GetGitCommentPlugin(p.Name, getClientSets(ownersFile, ownersConfigCache, ctx, pluginClient), p.Args...)
			if plugin == nil {
				ctx.Logger().Info("Plugin not found", "name", p.Name)
				continue
			}
			// Execute plugin
			if err := plugin.Do(ctx, pluginhelpers.GitCommentEventFromGithubIssueCommentEvent(payload)); err != nil {
				ctx.Logger().Error(err, "Failed to execute plugin", "name", plugin.Name())
			}
		}
	}))
	// Listen for GitHub pull request events
	githubApp.On(probot.GitHub.PullRequest).WithHandler(probot.GitHub.PullRequest.Handler(func(ctx probot.GitHubPullRequestContext) {
		payload := ctx.Payload()
		pluginClient := pluginhelpers.PluginConfigClientFromGithub(
			ctx.Client(), configPath, pluginConfigCache,
		)
		cfg, err := pluginClient.GetConfig(payload.Repo.Owner.GetLogin(), payload.Repo.GetName())
		ctx.Must(err)
		// Execute all plugins in config
		for _, p := range cfg.Plugins {
			plugin := plugins.GetGitCommentPlugin(p.Name, getClientSets(ownersFile, ownersConfigCache, ctx, pluginClient), p.Args...)
			if plugin == nil {
				ctx.Logger().Info("Plugin not found", "name", p.Name)
				continue
			}
			// Execute plugin
			if err := plugin.Do(ctx, pluginhelpers.GitCommentEventFromGithubPullRequestEvent(payload)); err != nil {
				ctx.Logger().Error(err, "Failed to execute plugin", "name", plugin.Name())
			}
		}
	}))
	// Listen for GitHub pull request review events
	githubApp.On(
		probot.GitHub.PullRequestReview,
	).WithHandler(probot.GitHub.PullRequestReview.Handler(func(ctx probot.GitHubPullRequestReviewContext) {
		payload := ctx.Payload()
		pluginClient := pluginhelpers.PluginConfigClientFromGithub(
			ctx.Client(), configPath, pluginConfigCache,
		)
		cfg, err := pluginClient.GetConfig(payload.Repo.Owner.GetLogin(), payload.Repo.GetName())
		ctx.Must(err)

		for _, p := range cfg.Plugins {
			plugin := plugins.GetGitCommentPlugin(p.Name, getClientSets(ownersFile, ownersConfigCache, ctx, pluginClient), p.Args...)
			if plugin == nil {
				ctx.Logger().Info("Plugin not found", "name", p.Name)
				continue
			}
			// Execute plugin
			if err := plugin.Do(ctx, pluginhelpers.GitCommentEventFromGithubPullRequestReviewEvent(payload)); err != nil {
				ctx.Logger().Error(err, "Failed to execute plugin", "name", plugin.Name())
			}
		}
	}))
	// Listen for GitHub pull request review comment events
	githubApp.On(
		probot.GitHub.PullRequestReviewComment,
	).WithHandler(probot.GitHub.PullRequestReviewComment.Handler(func(ctx probot.GitHubPullRequestReviewCommentContext) {
		payload := ctx.Payload()
		pluginClient := pluginhelpers.PluginConfigClientFromGithub(
			ctx.Client(), configPath, pluginConfigCache,
		)
		cfg, err := pluginClient.GetConfig(payload.Repo.Owner.GetLogin(), payload.Repo.GetName())
		ctx.Must(err)

		for _, p := range cfg.Plugins {
			plugin := plugins.GetGitCommentPlugin(p.Name, getClientSets(ownersFile, ownersConfigCache, ctx, pluginClient), p.Args...)
			if plugin == nil {
				ctx.Logger().Info("Plugin not found", "name", p.Name)
				continue
			}
			// Execute plugin
			if err := plugin.Do(ctx, pluginhelpers.GitCommentEventFromGithubPullRequestReviewCommentEvent(payload)); err != nil {
				ctx.Logger().Error(err, "Failed to execute plugin", "name", plugin.Name())
			}
		}
	}))
	// Listen for GitHub push events
	githubApp.On(probot.GitHub.Push).WithHandler(probot.GitHub.Push.Handler(func(ctx probot.GitHubPushContext) {
		// TODO: Implement
	}))
	// Listen for GitHub status events
	githubApp.On(probot.GitHub.Status).WithHandler(probot.GitHub.Status.Handler(func(ctx probot.GitHubStatusContext) {
		// TODO: Implement
	}))

	return githubApp
}

func getClientSets[PT any](
	ownersFile string,
	ownersConfigCache pluginhelpers.ConfigCache[plugins.OwnersConfiguration],
	ctx probot.ProbotContext[probot.GitHubClient, PT],
	pluginClient plugins.PluginConfigClient,
) plugins.ClientSets {
	logger := ctx.Logger()
	return plugins.ClientSets{
		GitIssueClient:     pluginhelpers.GitIssueClientFromGithub(ctx.Client()),
		GitPRClient:        pluginhelpers.GitPRClientFromGithub(ctx.Client()),
		PluginConfigClient: pluginClient,
		OwnersClient:       pluginhelpers.OwnersClientFromGithub(ctx.Client(), ownersFile, ownersConfigCache),
		GitRepoClient:      pluginhelpers.GitRepoClientFromGithub(ctx.Client()),
		GitSearchClient:    pluginhelpers.GitSearchClientFromGithub(ctx.GraphQL()),
		LoggerClient: pluginhelpers.MakeLoggerClient(func() logr.Logger {
			return logger
		}),
	}
}
