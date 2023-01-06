package pluginhelpers

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	"github.com/google/go-github/v48/github"
	"sigs.k8s.io/yaml"

	"github.com/airconduct/go-probot"
	"github.com/airconduct/kuilei/pkg/pluginhelpers/syncer"
	"github.com/airconduct/kuilei/pkg/plugins"
)

func OwnersClientFromGithub(gh *probot.GithubClient, ownersFileName string, cache ConfigCache[plugins.OwnersConfiguration]) plugins.OwnersClient {
	c := &githubOwnersClient{
		ghClient:       gh,
		ownersFileName: ownersFileName,
		configCache:    cache,
	}

	return c
}

type githubOwnersClient struct {
	ghClient       *probot.GithubClient
	ownersFileName string

	configCache ConfigCache[plugins.OwnersConfiguration]
}

func (c *githubOwnersClient) GetOwners(owner, repo, file string) (plugins.OwnersConfiguration, error) {
	cfg := c.configCache.Get(owner, repo, file)
	if cfg == nil {
		if err := c.syncOwnersFromRemote(owner, repo); err != nil {
			return plugins.OwnersConfiguration{}, err
		}
		syncer.CacheSyncer.EnsureSync(syncer.CacheSyncFuncKey{
			Owner: owner, Repo: repo, Kind: "owners",
		}, c.syncOwnersFromRemote)
		cfg = c.configCache.Get(owner, repo, file)
		if cfg == nil {
			return plugins.OwnersConfiguration{}, nil
		}
	}
	return *cfg, nil
}

func (c *githubOwnersClient) syncOwnersFromRemote(owner, repo string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	query := fmt.Sprintf("repo:%s/%s filename:%s", owner, repo, c.ownersFileName)
	result, _, err := c.ghClient.Search.Code(ctx, query, &github.SearchOptions{})
	if err != nil {
		return err
	}
	for _, r := range result.CodeResults {
		file, _, _, err := c.ghClient.Repositories.GetContents(ctx, owner, repo, r.GetPath(), &github.RepositoryContentGetOptions{})
		if err != nil {
			return err
		}
		contents, err := file.GetContent()
		if err != nil {
			return err
		}
		cfg := &plugins.OwnersConfiguration{}
		if err := yaml.Unmarshal([]byte(contents), cfg); err != nil {
			return err
		}

		c.configCache.Save(owner, repo, filepath.Dir(r.GetPath()), cfg)
	}
	return nil
}
