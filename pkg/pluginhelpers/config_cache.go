package pluginhelpers

import (
	"fmt"
	"sync"

	"github.com/airconduct/kuilei/pkg/plugins"
)

func NewPluginConfigCache() *PluginConfigCache {
	return &PluginConfigCache{}
}

type PluginConfigCache struct {
	sync.Map
}

func (c PluginConfigCache) Get(owner, repo string) *plugins.Configuration {
	key := c.key(owner, repo)
	v, ok := c.Map.Load(key)
	if !ok {
		return nil
	}
	return v.(*plugins.Configuration)
}

func (c PluginConfigCache) Save(owner, repo string, cfg *plugins.Configuration) {
	key := c.key(owner, repo)
	c.Map.Store(key, cfg)
}

func (c PluginConfigCache) key(owner, repo string) string {
	return fmt.Sprintf("%s/%s", owner, repo)
}
