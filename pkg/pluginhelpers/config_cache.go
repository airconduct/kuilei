package pluginhelpers

import (
	"fmt"
	"sync"
)

type ConfigCache[T any] interface {
	Get(owner, repo, path string) *T
	Save(owner, repo, path string, cfg *T)
}

func NewConfigCache[T any]() ConfigCache[T] {
	return &configCache[T]{}
}

type configCache[T any] struct {
	sync.Map
}

func (c *configCache[T]) Get(owner, repo, path string) *T {
	key := c.key(owner, repo, path)
	v, ok := c.Map.Load(key)
	if !ok {
		return nil
	}
	return v.(*T)
}

func (c *configCache[T]) Save(owner, repo, path string, cfg *T) {
	key := c.key(owner, repo, path)
	c.Map.Store(key, cfg)
}

func (c *configCache[T]) key(owner, repo, path string) string {
	return fmt.Sprintf("%s/%s/%s", owner, repo, path)
}
