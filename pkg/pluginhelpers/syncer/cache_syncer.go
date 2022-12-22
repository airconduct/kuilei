package syncer

import (
	"fmt"
	"log"
	"sync"
	"time"

	"k8s.io/apimachinery/pkg/util/wait"
)

var CacheSyncer = &cacheSyncer{
	syncFuncs: make(map[CacheSyncFuncKey]func(owner string, repo string) error),
}

type CacheSyncFuncKey struct {
	Owner string
	Repo  string
	Kind  string
}

func (k CacheSyncFuncKey) String() string {
	return fmt.Sprintf("{owner:%s, repo:%s, kind:%s}", k.Owner, k.Repo, k.Kind)
}

type cacheSyncer struct {
	mutex     sync.RWMutex
	startOnce sync.Once

	syncFuncs map[CacheSyncFuncKey]func(owner, repo string) error
}

func (c *cacheSyncer) EnsureSync(key CacheSyncFuncKey, fn func(owner, repo string) error) {
	c.mutex.Lock()
	if _, ok := c.syncFuncs[key]; !ok {
		c.syncFuncs[key] = fn
	}
	c.mutex.Unlock()

	c.startOnce.Do(c.startToSync)
}

func (c *cacheSyncer) startToSync() {
	go wait.Forever(func() {
		c.mutex.RLock()
		defer c.mutex.RUnlock()

		for key, fn := range c.syncFuncs {
			if err := fn(key.Owner, key.Repo); err != nil {
				log.Printf("Failed to sync function for key %v, error: %v", key, err)
			}
		}
	}, 5*time.Minute)
}
