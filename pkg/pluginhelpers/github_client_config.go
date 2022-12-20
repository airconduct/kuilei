package pluginhelpers

import (
	"context"

	"github.com/google/go-github/v48/github"
	"sigs.k8s.io/yaml"

	"github.com/airconduct/kuilei/pkg/plugins"
	"github.com/airconduct/kuilei/pkg/probot"
)

func PluginConfigClientFromGithub(gh *probot.GithubClient, configPath string, cache *PluginConfigCache) plugins.PluginConfigClient {
	c := &githubPluginConfigClient{
		ghClient: gh, configPath: configPath,
		configCache: cache,
	}

	return c
}

type githubPluginConfigClient struct {
	ghClient   *probot.GithubClient
	configPath string

	configCache *PluginConfigCache
}

func (c *githubPluginConfigClient) GetConfig(owner, repo string) (plugins.Configuration, error) {
	cfg := c.getConfigFromCache(owner, repo)
	if cfg != nil {
		return *cfg, nil
	}

	cfg, err := c.getConfigFromRemote(context.TODO(), owner, repo)
	if err != nil {
		return plugins.Configuration{}, err
	}
	c.cacheConfig(owner, repo, cfg)
	return *cfg, nil
}

func (c *githubPluginConfigClient) getConfigFromCache(owner, repo string) *plugins.Configuration {
	return c.configCache.Get(owner, repo)
}

func (c *githubPluginConfigClient) getConfigFromRemote(ctx context.Context, owner, repo string) (*plugins.Configuration, error) {
	file, _, _, err := c.ghClient.Repositories.GetContents(ctx, owner, repo, c.configPath, &github.RepositoryContentGetOptions{})
	if err != nil {
		return nil, err
	}
	contents, err := file.GetContent()
	if err != nil {
		return nil, err
	}

	cfg := new(plugins.Configuration)
	if err := yaml.Unmarshal([]byte(contents), cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (c *githubPluginConfigClient) cacheConfig(owner, repo string, cfg *plugins.Configuration) {
	c.configCache.Save(owner, repo, cfg)
}
