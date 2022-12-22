package pluginhelpers

import (
	"fmt"
	"path/filepath"
	"strings"
	"sync"
)

func NewConfigNearestCache[T any]() ConfigCache[T] {
	return &nearestConfigCache[T]{
		trie: &NearestTrie{Root: &TrieNode{Children: make(map[string]*TrieNode)}},
	}
}

type nearestConfigCache[T any] struct {
	sync.Map

	trie *NearestTrie
}

func (c *nearestConfigCache[T]) Get(owner, repo, path string) *T {
	key := c.key(owner, repo, path)
	nearestKey := c.trie.Search(key)
	v, ok := c.Map.Load(nearestKey)
	if !ok {
		return nil
	}
	return v.(*T)
}

func (c *nearestConfigCache[T]) Save(owner, repo, path string, cfg *T) {
	key := c.key(owner, repo, path)
	c.trie.Insert(key)
	c.Map.Store(key, cfg)
}

func (c *nearestConfigCache[T]) key(owner, repo, path string) string {
	path = strings.TrimLeft(path, "/")
	return filepath.Clean(fmt.Sprintf("%s/%s/%s", owner, repo, path))
}
