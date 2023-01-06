package plugins_test

import (
	"context"
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/spf13/pflag"

	"github.com/airconduct/kuilei/pkg/plugins"
	"github.com/airconduct/kuilei/pkg/plugins/mock"
)

var _ = Describe("Test Plugins interface", func() {
	When("Register fake plugin", func() {
		plugins.RegisterGitCommentPlugin("foo", func(cs plugins.ClientSets) plugins.GitCommentPlugin {
			return &fakePlugin{}
		})
		It("Should get correct output", func() {
			p := plugins.GetGitCommentPlugin("foo", plugins.ClientSets{}, []string{
				"--foo=xxxx", "--bar=3",
			}...)
			Expect(p.Do(context.TODO(), plugins.GitCommentEvent{})).Should(Succeed())
			Expect(p.(*fakePlugin).output).Should(Equal("xxxx-3"))
		})
		It("Should create help comment", func() {
			p := plugins.GetGitCommentPlugin("help", plugins.ClientSets{
				GitIssueClient: mock.FakeGitIssueClient(func(ctx context.Context, gic plugins.GitIssueComment) error {
					fmt.Println("======body", gic.Body)
					return nil
				}, nil, nil),
			})
			Expect(p.Do(context.Background(), plugins.GitCommentEvent{
				GitComment: plugins.GitComment{
					Body: `/help`,
				},
				Action: plugins.GitCommentActionCreated,
			})).Should(Succeed())
		})
	})
})

type fakePlugin struct {
	foo string
	bar int

	output string
}

func (p *fakePlugin) Name() string {
	return "foo"
}

func (p *fakePlugin) Usage() string {
	return "no usage"
}

func (p *fakePlugin) Description() string {
	return "just a fake plugin"
}

func (p *fakePlugin) BindFlags(flags *pflag.FlagSet) {
	flags.StringVar(&p.foo, "foo", "", "")
	flags.IntVar(&p.bar, "bar", 0, "")
}

func (p *fakePlugin) Do(context.Context, plugins.GitCommentEvent) error {
	p.output = fmt.Sprintf("%s-%d", p.foo, p.bar)
	return nil
}
