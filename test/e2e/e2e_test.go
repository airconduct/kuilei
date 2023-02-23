package e2e

import (
	"context"
	"testing"

	"github.com/h2non/gock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/format"
	"github.com/spf13/pflag"

	"github.com/airconduct/go-probot"
	probotmock "github.com/airconduct/go-probot/mock"
	"github.com/airconduct/kuilei/pkg/app/github"
	_ "github.com/airconduct/kuilei/pkg/plugins/factory"
	"github.com/airconduct/kuilei/pkg/signals"
)

func TestPluginhelpers(t *testing.T) {
	RegisterFailHandler(Fail)
	format.MaxLength = 1 << 20
	RunSpecs(t, "Kuilei E2E Suite")
}

var (
	githubApp probot.App[probot.GitHubClient]
	ctx       context.Context
	cancel    func()
)

var _ = BeforeSuite(func() {
	flags := pflag.NewFlagSet("tess-cluster-e2e", pflag.ExitOnError)
	// pflag.CommandLine = flags
	b := github.New()
	b.BindFlags(flags)
	flags.Parse([]string{
		"--github.hmac-token-file=testdata/hmac_token",
		"--github.private-key-file=testdata/tls.key",
		"--github.appid=1",
		"--address=127.0.0.1",
		"--port=7771",
		"--path=/hook",
	})

	var err error
	githubApp, err = b.Build()
	Expect(err).Should(BeNil())

	ctx, cancel = context.WithCancel(signals.SetupSignalContext())
	go func() {
		Expect(githubApp.Run(ctx)).Should(Succeed())
	}()
})

var _ = AfterSuite(func() {
	cancel()
	gock.Off()
})

func sendToGitHubApp(e probot.WebhookEvent, v interface{}) error {
	return probotmock.Send(githubApp.(probotmock.AppMock[probot.GitHubClient]), e, v)
}
