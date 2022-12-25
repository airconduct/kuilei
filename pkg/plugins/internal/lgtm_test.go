package internal_test

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/airconduct/kuilei/pkg/plugins"
	"github.com/airconduct/kuilei/pkg/plugins/mock"
)

var _ = Describe("Plugin lgtm", func() {
	var addedLabels []plugins.Label
	var removedLabel plugins.Label
	plugin := plugins.GetGitCommentPlugin("lgtm", plugins.ClientSets{
		GitIssueClient: mock.FakeGitIssueClient(
			func(ctx context.Context, gic plugins.GitIssueComment) error {
				return nil
			},
			func(ctx context.Context, l []plugins.Label) error {
				addedLabels = l
				return nil
			},
			func(ctx context.Context, repo plugins.GitRepo, issue plugins.GitIssue, l plugins.Label) error {
				removedLabel = l
				return nil
			},
		),
		GitPRClient: mock.FakeGitPRClient(
			func(ctx context.Context, gr plugins.GitRepo, gpr plugins.GitPullRequest) ([]plugins.GitCommitFile, error) {
				return []plugins.GitCommitFile{
					{Path: "pkg/xxxx/aaaa/1111"}, {Path: "pkg/yyyy/aaaa"}, {Path: "foo"}, {Path: "bar/xxxx/bbbb"},
				}, nil
			},
			func(ctx context.Context, repo plugins.GitRepo, number int) (*plugins.GitPullRequest, error) {
				return &plugins.GitPullRequest{Number: 11, User: plugins.GitUser{Name: "foouser"}}, nil
			},
			func(ctx context.Context, repo plugins.GitRepo, number int, method string) error {
				return nil
			},
		),
		PluginConfigClient: mock.FakeConfigClient(
			func(owner, repo string) (plugins.Configuration, error) {
				return plugins.Configuration{
					Plugins: []plugins.PluginConfiguration{
						{Name: "lgtm", Args: []string{"--allow-author=true"}},
					},
				}, nil
			},
		),
		OwnersClient: mock.FakeOwnerClient(func(owner, repo, file string) (plugins.OwnersConfiguration, error) {
			switch file {
			case "pkg/yyyy/aaaa":
				return plugins.OwnersConfiguration{Approvers: []string{"foo", "bar"}}, nil
			}
			return plugins.OwnersConfiguration{Reviewers: []string{"foouser"}}, nil
		}),
	}, []string{"--allow-author=true"}...)

	It("Should not add lgtm", func() {
		addedLabels = []plugins.Label{}
		Expect(plugin.Do(context.TODO(), plugins.GitCommentEvent{
			GitComment: plugins.GitComment{
				IsPR: true, Number: 11, Body: "/lgtdm",
				User: plugins.GitUser{Name: "foo"},
			},
			Action: plugins.GitCommentActionCreated,
		})).Should(Succeed())
		Expect(addedLabels).Should(Equal([]plugins.Label{}))
	})
	It("Should not add lgtm", func() {
		addedLabels = []plugins.Label{}
		Expect(plugin.Do(context.TODO(), plugins.GitCommentEvent{
			GitComment: plugins.GitComment{
				IsPR: true, Number: 11, Body: "/lgtm",
				User: plugins.GitUser{Name: "foobar"},
			},
			Action: plugins.GitCommentActionCreated,
		})).Should(Succeed())
		Expect(addedLabels).Should(Equal([]plugins.Label{}))
	})
	It("Should add lgtm", func() {
		addedLabels = []plugins.Label{}
		Expect(plugin.Do(context.TODO(), plugins.GitCommentEvent{
			GitComment: plugins.GitComment{
				IsPR: true, Number: 11, Body: "/lgtm",
				User: plugins.GitUser{Name: "foo"},
			},
			Action: plugins.GitCommentActionCreated,
		})).Should(Succeed())
		Expect(addedLabels).Should(Equal([]plugins.Label{{Name: "lgtm"}}))
	})
	It("Should add lgtm", func() {
		addedLabels = []plugins.Label{}
		Expect(plugin.Do(context.TODO(), plugins.GitCommentEvent{
			GitComment: plugins.GitComment{
				IsPR: true, Number: 11, Body: "/lgtm dsadf",
				User: plugins.GitUser{Name: "foo"},
			},
			Action: plugins.GitCommentActionCreated,
		})).Should(Succeed())
		Expect(addedLabels).Should(Equal([]plugins.Label{{Name: "lgtm"}}))
	})
	It("Should remove label lgtm", func() {
		addedLabels = []plugins.Label{}
		Expect(plugin.Do(context.TODO(), plugins.GitCommentEvent{
			GitComment: plugins.GitComment{
				IsPR: true, Number: 11, Body: "/lgtm cancel",
				User: plugins.GitUser{Name: "foo"},
			},
			Action: plugins.GitCommentActionCreated,
		})).Should(Succeed())
		Expect(addedLabels).Should(Equal([]plugins.Label{}))
		Expect(removedLabel).Should(Equal(plugins.Label{Name: "lgtm"}))
	})
})
