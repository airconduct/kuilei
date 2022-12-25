package mock

import (
	"context"

	"github.com/airconduct/kuilei/pkg/plugins"
)

func FakeRepoClient(funcs map[string]interface{}) plugins.GitRepoClient {
	return &fakeRepoClient{funcs: funcs}
}

type fakeRepoClient struct {
	funcs map[string]interface{}
}

func (c *fakeRepoClient) CreateStatus(ctx context.Context, repo plugins.GitRepo, ref string, status plugins.GitCommitStatus) error {
	return c.funcs["CreateStatus"].(func(ctx context.Context, repo plugins.GitRepo, ref string, status plugins.GitCommitStatus) error)(
		ctx, repo, ref, status,
	)
}
