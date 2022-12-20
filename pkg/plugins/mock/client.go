package mock

import (
	"context"

	"github.com/airconduct/kuilei/pkg/plugins"
)

func FakeGitIssueClient(
	createIssueComment func(context.Context, plugins.GitIssueComment) error,
	addLabel func(context.Context, []plugins.Label) error,
) plugins.GitIssueClient {
	return &fakeIssueClient{
		createIssueComment: createIssueComment,
		addLabel:           addLabel,
	}
}

type fakeIssueClient struct {
	createIssueComment func(context.Context, plugins.GitIssueComment) error
	addLabel           func(context.Context, []plugins.Label) error
}

func (c *fakeIssueClient) CreateIssueComment(ctx context.Context, repo plugins.GitRepo, issue plugins.GitIssue, comment plugins.GitIssueComment) error {
	return c.createIssueComment(ctx, comment)
}
func (c *fakeIssueClient) AddLabel(ctx context.Context, repo plugins.GitRepo, issue plugins.GitIssue, labels []plugins.Label) error {
	return c.addLabel(ctx, labels)
}
