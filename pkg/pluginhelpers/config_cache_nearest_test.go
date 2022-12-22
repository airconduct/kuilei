package pluginhelpers_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/airconduct/kuilei/pkg/pluginhelpers"
	"github.com/airconduct/kuilei/pkg/plugins"
)

var _ = Describe("ConfigNearestCache", func() {
	When("Using a single owner file in root dir", func() {
		cache := pluginhelpers.NewConfigNearestCache[plugins.OwnersConfiguration]()
		cache.Save("foo", "bar", "", &plugins.OwnersConfiguration{
			Owner: "foo", Repo: "bar",
			Reviewers: []string{"foouser"},
			Approvers: []string{"baruser"},
		})
		It("Should get root owner config", func() {
			cfg := cache.Get("foo", "bar", "pkg/xxxx/aaaa/1111")
			Expect(cfg).ShouldNot(BeNil())
			Expect(*cfg).Should(Equal(plugins.OwnersConfiguration{
				Owner: "foo", Repo: "bar",
				Reviewers: []string{"foouser"},
				Approvers: []string{"baruser"},
			}))
		})
		It("Should get none", func() {
			cfg := cache.Get("foo", "bar2", "pkg/xxxx/aaaa/1111")
			Expect(cfg).Should(BeNil())
		})
	})

	When("Using multiple owner file", func() {
		cache := pluginhelpers.NewConfigNearestCache[plugins.OwnersConfiguration]()
		// root file
		cache.Save("foo", "bar", "", &plugins.OwnersConfiguration{
			Owner: "foo", Repo: "bar",
			Reviewers: []string{"foouser"},
			Approvers: []string{"baruser"},
		})
		// subpath
		cache.Save("foo", "bar", "pkg", &plugins.OwnersConfiguration{
			Owner: "foo", Repo: "bar",
			Reviewers: []string{"foouser-pkg"},
			Approvers: []string{"baruser-pkg"},
		})
		cache.Save("foo", "bar", "pkg/xxxx/aaaa", &plugins.OwnersConfiguration{
			Owner: "foo", Repo: "bar",
			Reviewers: []string{"foouser-pkg/xxxx/aaaa"},
			Approvers: []string{"baruser-pkg/xxxx/aaaa"},
		})
		cache.Save("foo", "bar", "cmd/xxxx", &plugins.OwnersConfiguration{
			Owner: "foo", Repo: "bar",
			Reviewers: []string{"foouser-cmd/xxxx"},
			Approvers: []string{"baruser-cmd/xxxx"},
		})
		It("Should get parent owner file", func() {
			cfg := cache.Get("foo", "bar", "pkg/xxxx/aaaa/1111")
			Expect(cfg).ShouldNot(BeNil())
			Expect(cfg.Reviewers).Should(Equal([]string{"foouser-pkg/xxxx/aaaa"}))
		})
		It("Should get current owner file", func() {
			cfg := cache.Get("foo", "bar", "cmd/xxxx/foo.go")
			Expect(cfg).ShouldNot(BeNil())
			Expect(cfg.Reviewers).Should(Equal([]string{"foouser-cmd/xxxx"}))
		})
		It("Should get root owner file", func() {
			cfg := cache.Get("foo", "bar", "cmd/yyyy/foo.go")
			Expect(cfg).ShouldNot(BeNil())
			Expect(cfg.Reviewers).Should(Equal([]string{"foouser"}))
		})
	})

})
