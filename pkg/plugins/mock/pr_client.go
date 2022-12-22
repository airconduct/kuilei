package mock

import (
	"context"

	"github.com/airconduct/kuilei/pkg/plugins"
)

func FakeGitPRClient(
	listFiles func(context.Context, plugins.GitRepo, plugins.GitPullRequest) ([]plugins.GitCommitFile, error),
	getPR func(ctx context.Context, repo plugins.GitRepo, number int) (*plugins.GitPullRequest, error),
) plugins.GitPRClient {
	return &fakePRClient{
		listFiles: listFiles,
		getPR:     getPR,
	}
}

type fakePRClient struct {
	listFiles func(context.Context, plugins.GitRepo, plugins.GitPullRequest) ([]plugins.GitCommitFile, error)
	getPR     func(ctx context.Context, repo plugins.GitRepo, number int) (*plugins.GitPullRequest, error)
}

func (c *fakePRClient) ListFiles(ctx context.Context, repo plugins.GitRepo, pr plugins.GitPullRequest) ([]plugins.GitCommitFile, error) {
	return c.listFiles(ctx, repo, pr)
}

func (c *fakePRClient) GetPR(ctx context.Context, repo plugins.GitRepo, number int) (*plugins.GitPullRequest, error) {
	return c.getPR(ctx, repo, number)
}
