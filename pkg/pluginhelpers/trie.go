package pluginhelpers

import (
	"path/filepath"
	"strings"
	"sync"
)

type NearestTrie struct {
	sync.RWMutex
	Root *TrieNode
}

type TrieNode struct {
	Children map[string]*TrieNode
	Leaf     bool
}

func (t *NearestTrie) Insert(path string) {
	t.Lock()
	defer t.Unlock()

	node := t.Root

	path = filepath.Clean(strings.TrimLeft(path, "/"))
	pathlist := strings.Split(path, string([]byte{filepath.Separator}))
	for i := 0; i < len(pathlist); i++ {
		key := pathlist[i]
		if node.Children[key] == nil {
			node.Children[key] = new(TrieNode)
			node.Children[key].Children = map[string]*TrieNode{}
		}
		node = node.Children[key]
	}
	node.Leaf = true
}

func (t *NearestTrie) Search(path string) string {
	t.RLock()
	defer t.RUnlock()

	node := t.Root

	path = filepath.Clean(strings.TrimLeft(path, "/"))
	pathlist := strings.Split(path, string([]byte{filepath.Separator}))

	nearest := ""
	current := ""
	for i := 0; i < len(pathlist); i++ {
		key := pathlist[i]
		if node.Leaf {
			nearest = current
		}
		if node.Children[key] == nil {
			return nearest
		}
		current = filepath.Join(current, key)
		node = node.Children[key]
	}
	return nearest
}
