package mock

import "github.com/airconduct/kuilei/pkg/plugins"

func FakeConfigClient(getConfig func(owner, repo string) (plugins.Configuration, error)) plugins.PluginConfigClient {
	return &fakeConfigClient{getConfig: getConfig}
}

type fakeConfigClient struct {
	getConfig func(owner, repo string) (plugins.Configuration, error)
}

func (c *fakeConfigClient) GetConfig(owner, repo string) (plugins.Configuration, error) {
	return c.getConfig(owner, repo)
}
