package plugins

import "context"

type ClientSets struct {
	GitIssueClient
	GitPRClient

	PluginConfigClient
	OwnersClient
}

type GitIssueClient interface {
	CreateIssueComment(context.Context, GitRepo, GitIssue, GitIssueComment) error
	AddLabel(context.Context, GitRepo, GitIssue, []Label) error
	RemoveLabel(context.Context, GitRepo, GitIssue, Label) error
}

type GitPRClient interface {
	ListFiles(context.Context, GitRepo, GitPullRequest) ([]GitCommitFile, error)
	GetPR(ctx context.Context, repo GitRepo, number int) (*GitPullRequest, error)
}

type PluginConfigClient interface {
	GetConfig(owner, repo string) (Configuration, error)
}

type OwnersClient interface {
	GetOwners(owner, repo, file string) (OwnersConfiguration, error)
}
