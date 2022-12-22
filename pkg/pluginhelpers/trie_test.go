package pluginhelpers_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/airconduct/kuilei/pkg/pluginhelpers"
)

var _ = Describe("Trie struct", func() {
	When("Insert fake data", func() {
		trie := pluginhelpers.NearestTrie{Root: &pluginhelpers.TrieNode{Children: make(map[string]*pluginhelpers.TrieNode), Leaf: true}}
		trie.Insert("pkg")
		trie.Insert("pkg/plugins/xxx")
		trie.Insert("pkg/plugins/yyy")
		trie.Insert("pkg/plugins/zzz")
		trie.Insert("cmd/aaa")
		trie.Insert("cmd/bbb")
		It("Should get correct key", func() {
			Expect(trie.Search("foo")).Should(Equal(""))
			Expect(trie.Search("pkg/plugins/zzz/111/222/333")).Should(Equal("pkg/plugins/zzz"))
			Expect(trie.Search("cmd/ccc/111/222")).Should(Equal(""))
			Expect(trie.Search("pkg/plugins/ddd")).Should(Equal("pkg"))
		})
	})
})
