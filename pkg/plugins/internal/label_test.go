package internal_test

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/airconduct/kuilei/pkg/plugins"
	_ "github.com/airconduct/kuilei/pkg/plugins/internal"
	"github.com/airconduct/kuilei/pkg/plugins/mock"
)

var _ = Describe("Label Plugin", func() {
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
		),
		PluginConfigClient: nil,
	})
	It("Should add ok-to-test label", func() {
		Expect(plugin.Do(context.Background(), plugins.GitCommentEvent{
			Action: plugins.GitCommentActionCreated,
			GitComment: plugins.GitComment{
				Body: "/label ok-to-test",
			},
		})).To(Succeed())
		Expect(actualLabels).Should(Equal([]plugins.Label{{Name: "ok-to-test"}}))
	})
})
