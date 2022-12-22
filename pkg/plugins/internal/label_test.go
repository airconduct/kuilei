package internal_test

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/airconduct/kuilei/pkg/plugins"
	_ "github.com/airconduct/kuilei/pkg/plugins/internal"
	"github.com/airconduct/kuilei/pkg/plugins/mock"
)

var _ = Describe("Plugin label", func() {
	var actualLabels []plugins.Label

	plugin := plugins.GetGitCommentPlugin("label", plugins.ClientSets{
		GitIssueClient: mock.FakeGitIssueClient(
			func(ctx context.Context, gic plugins.GitIssueComment) error {
				return nil
			},
			func(ctx context.Context, l []plugins.Label) error {
				actualLabels = l
				return nil
			},
			func(ctx context.Context, repo plugins.GitRepo, issue plugins.GitIssue, l plugins.Label) error {
				return nil
			},
		),
		PluginConfigClient: mock.FakeConfigClient(
			func(owner, repo string) (plugins.Configuration, error) {
				return plugins.Configuration{
					Plugins: []plugins.PluginConfiguration{
						{Name: "label", Args: []string{"--forbidden=lgtm,approved"}},
					},
				}, nil
			},
		),
	}, []string{"--forbidden=lgtm,approved"}...)

	It("Should add ok-to-test label", func() {
		actualLabels = []plugins.Label{}
		Expect(plugin.Do(context.Background(), plugins.GitCommentEvent{
			Action: plugins.GitCommentActionCreated,
			GitComment: plugins.GitComment{
				Body: "/label ok-to-test",
			},
		})).To(Succeed())
		Expect(actualLabels).Should(Equal([]plugins.Label{{Name: "ok-to-test"}}))
	})
	It("Should add multiple labels", func() {
		actualLabels = []plugins.Label{}
		Expect(plugin.Do(context.Background(), plugins.GitCommentEvent{
			Action: plugins.GitCommentActionCreated,
			GitComment: plugins.GitComment{
				Body: "/label ok-to-test\r\n/label foo",
			},
		})).To(Succeed())
		Expect(actualLabels).Should(Equal([]plugins.Label{{Name: "ok-to-test"}, {Name: "foo"}}))
	})
	It("Should not add lgtm label", func() {
		actualLabels = []plugins.Label{}
		Expect(plugin.Do(context.Background(), plugins.GitCommentEvent{
			Action: plugins.GitCommentActionCreated,
			GitComment: plugins.GitComment{
				Body: "/label foo\r\n/label lgtm",
			},
		})).To(Succeed())
		Expect(actualLabels).Should(Equal([]plugins.Label{{Name: "foo"}}))
	})
	It("Should not add approved label", func() {
		actualLabels = []plugins.Label{}
		Expect(plugin.Do(context.Background(), plugins.GitCommentEvent{
			Action: plugins.GitCommentActionCreated,
			GitComment: plugins.GitComment{
				Body: "/label approved",
			},
		})).To(Succeed())
		Expect(actualLabels).Should(Equal([]plugins.Label{}))
	})
	It("Should not add any label", func() {
		actualLabels = []plugins.Label{}
		Expect(plugin.Do(context.Background(), plugins.GitCommentEvent{
			Action: plugins.GitCommentActionCreated,
			GitComment: plugins.GitComment{
				Body: "/label",
			},
		})).To(Succeed())
		Expect(actualLabels).Should(Equal([]plugins.Label{}))
	})
})
