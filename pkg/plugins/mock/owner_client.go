package mock

import (
	"github.com/airconduct/kuilei/pkg/plugins"
)

func FakeOwnerClient(
	getOwners func(owner, repo, file string) (plugins.OwnersConfiguration, error),
) plugins.OwnersClient {
	return &fakeOwnerClient{
		getOwners: getOwners,
	}
}

type fakeOwnerClient struct {
	getOwners func(owner, repo, file string) (plugins.OwnersConfiguration, error)
}

func (c *fakeOwnerClient) GetOwners(owner, repo, file string) (plugins.OwnersConfiguration, error) {
	return c.getOwners(owner, repo, file)
}
