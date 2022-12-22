package pluginhelpers

import (
	"context"
	"time"

	"github.com/google/go-github/v48/github"
	"sigs.k8s.io/yaml"

	"github.com/airconduct/kuilei/pkg/pluginhelpers/syncer"
	"github.com/airconduct/kuilei/pkg/plugins"
	"github.com/airconduct/kuilei/pkg/probot"
)

func PluginConfigClientFromGithub(gh *probot.GithubClient, configPath string, cache ConfigCache[plugins.Configuration]) plugins.PluginConfigClient {
	c := &githubPluginConfigClient{
		ghClient: gh, configPath: configPath,
		configCache: cache,
	}

	return c
}

type githubPluginConfigClient struct {
	ghClient   *probot.GithubClient
	configPath string

	configCache ConfigCache[plugins.Configuration]
}

func (c *githubPluginConfigClient) GetConfig(owner, repo string) (plugins.Configuration, error) {
	cfg := c.getConfigFromCache(owner, repo)
	if cfg == nil {
		err := c.syncConfigFromRemote(owner, repo)
		if err != nil {
			return plugins.Configuration{}, err

		}
		syncer.CacheSyncer.EnsureSync(syncer.CacheSyncFuncKey{
			Owner: owner, Repo: repo, Kind: "plugin-config",
		}, c.syncConfigFromRemote)
		cfg = c.getConfigFromCache(owner, repo)
		if cfg == nil {
			return plugins.Configuration{}, nil
		}
	}
	return *cfg, nil
}

func (c *githubPluginConfigClient) getConfigFromCache(owner, repo string) *plugins.Configuration {
	return c.configCache.Get(owner, repo, c.configPath)
}

func (c *githubPluginConfigClient) syncConfigFromRemote(owner, repo string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	file, _, _, err := c.ghClient.Repositories.GetContents(ctx, owner, repo, c.configPath, &github.RepositoryContentGetOptions{})
	if err != nil {
		return err
	}
	contents, err := file.GetContent()
	if err != nil {
		return err
	}

	cfg := new(plugins.Configuration)
	if err := yaml.Unmarshal([]byte(contents), cfg); err != nil {
		return err
	}
	c.configCache.Save(owner, repo, c.configPath, cfg)
	return nil
}
