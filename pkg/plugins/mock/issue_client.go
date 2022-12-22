package mock

import (
	"context"

	"github.com/airconduct/kuilei/pkg/plugins"
)

func FakeGitIssueClient(
	createIssueComment func(context.Context, plugins.GitIssueComment) error,
	addLabel func(context.Context, []plugins.Label) error,
	removeLabel func(ctx context.Context, repo plugins.GitRepo, issue plugins.GitIssue, l plugins.Label) error,
) plugins.GitIssueClient {
	return &fakeIssueClient{
		createIssueComment: createIssueComment,
		addLabel:           addLabel,
		removeLabel:        removeLabel,
	}
}

type fakeIssueClient struct {
	createIssueComment func(context.Context, plugins.GitIssueComment) error
	addLabel           func(context.Context, []plugins.Label) error
	removeLabel        func(ctx context.Context, repo plugins.GitRepo, issue plugins.GitIssue, l plugins.Label) error
}

func (c *fakeIssueClient) CreateIssueComment(ctx context.Context, repo plugins.GitRepo, issue plugins.GitIssue, comment plugins.GitIssueComment) error {
	return c.createIssueComment(ctx, comment)
}

func (c *fakeIssueClient) AddLabel(ctx context.Context, repo plugins.GitRepo, issue plugins.GitIssue, labels []plugins.Label) error {
	return c.addLabel(ctx, labels)
}

func (c *fakeIssueClient) RemoveLabel(ctx context.Context, repo plugins.GitRepo, issue plugins.GitIssue, l plugins.Label) error {
	return c.removeLabel(ctx, repo, issue, l)
}
