package mock

import (
	"context"

	"github.com/airconduct/kuilei/pkg/plugins"
)

func FakeSearchClient(funcs map[string]interface{}) plugins.GitSearchClient {
	return &fakeSearchClient{funcs: funcs}
}

type fakeSearchClient struct {
	funcs map[string]interface{}
}

func (c *fakeSearchClient) SearchPR(ctx context.Context, repo plugins.GitRepo, state string) ([]plugins.GitPullRequestSearchResult, error) {
	return c.funcs["SearchPR"].(func(ctx context.Context, repo plugins.GitRepo, state string) ([]plugins.GitPullRequestSearchResult, error))(
		ctx, repo, state,
	)
}
