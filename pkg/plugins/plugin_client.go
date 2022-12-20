package plugins

import "context"

type ClientSets struct {
	GitIssueClient
	PluginConfigClient
}

type GitIssueClient interface {
	CreateIssueComment(context.Context, GitRepo, GitIssue, GitIssueComment) error
	AddLabel(context.Context, GitRepo, GitIssue, []Label) error
}

type PluginConfigClient interface {
	GetConfig(owner, repo string) (Configuration, error)
}
