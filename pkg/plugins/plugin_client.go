package plugins

import (
	"context"

	"github.com/go-logr/logr"
)

type ClientSets struct {
	GitIssueClient
	GitPRClient
	GitRepoClient
	GitSearchClient

	PluginConfigClient
	OwnersClient
	LoggerClient
}

type GitIssueClient interface {
	CreateIssueComment(context.Context, GitRepo, GitIssue, GitIssueComment) error
	AddLabel(context.Context, GitRepo, GitIssue, []Label) error
	RemoveLabel(context.Context, GitRepo, GitIssue, Label) error
}

type GitPRClient interface {
	ListFiles(context.Context, GitRepo, GitPullRequest) ([]GitCommitFile, error)
	GetPR(ctx context.Context, repo GitRepo, number int) (GitPullRequest, error)
	MergePR(ctx context.Context, repo GitRepo, number int, method string) error
}

type GitRepoClient interface {
	CreateStatus(ctx context.Context, repo GitRepo, ref string, status GitCommitStatus) error
	ListStatuses(ctx context.Context, repo GitRepo, ref string) ([]GitCommitStatus, error)
	ListChecks(ctx context.Context, repo GitRepo, ref string) ([]GitCommitCheck, error)
}

type GitSearchClient interface {
	SearchPR(ctx context.Context, repo GitRepo, state string) ([]GitPullRequestSearchResult, error)
}

type PluginConfigClient interface {
	GetConfig(owner, repo string) (Configuration, error)
}

type OwnersClient interface {
	GetOwners(owner, repo, file string) (OwnersConfiguration, error)
}

type LoggerClient interface {
	GetLogger() logr.Logger
}
