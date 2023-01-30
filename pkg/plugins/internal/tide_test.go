package internal_test

import (
	"context"
	"sync"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/util/wait"

	"github.com/airconduct/kuilei/pkg/plugins"
	"github.com/airconduct/kuilei/pkg/plugins/internal"
	"github.com/airconduct/kuilei/pkg/plugins/mock"
)

var _ = Describe("Plugin tide", func() {
	internal.TideSyncInterval = time.Second

	When("Using git comment plugin", func() {
		globalLock := sync.RWMutex{}
		merged := false
		fakePR := &plugins.GitPullRequest{
			Number:    1,
			State:     plugins.PullRequestStateOpen,
			Head:      plugins.GitBranch{Sha: "foo"},
			Locked:    false,
			Title:     "foo",
			Body:      "foo",
			Mergeable: plugins.GitMergeableStateMergeable,
			Labels:    []plugins.Label{},
			Commits: []plugins.GitCommit{
				{Sha: "foo"},
			},
		}
		callCreateStatus := false

		tide := plugins.GetGitCommentPlugin("tide", plugins.ClientSets{
			GitPRClient: mock.FakeGitPRClient(
				nil, nil, func(ctx context.Context, repo plugins.GitRepo, number int, method string) error {
					globalLock.Lock()
					defer globalLock.Unlock()
					merged = true
					return nil
				},
			),
			GitSearchClient: mock.FakeSearchClient(map[string]interface{}{
				"SearchPR": func(ctx context.Context, repo plugins.GitRepo, state string) ([]plugins.GitPullRequest, error) {
					globalLock.RLock()
					defer globalLock.RUnlock()
					return []plugins.GitPullRequest{*fakePR}, nil
				},
			}),
			GitRepoClient: mock.FakeRepoClient(map[string]interface{}{
				"CreateStatus": func(ctx context.Context, repo plugins.GitRepo, ref string, status plugins.GitCommitStatus) error {
					globalLock.Lock()
					defer globalLock.Unlock()
					callCreateStatus = true

					var target *plugins.GitCommitStatus
					for idx, s := range fakePR.Commits[0].Statuses {
						if s.Context == status.Context {
							target = &fakePR.Commits[0].Statuses[idx]
						}
					}
					if target == nil {
						fakePR.Commits[0].Statuses = append(fakePR.Commits[0].Statuses, plugins.GitCommitStatus{Context: status.Context})
						target = &fakePR.Commits[0].Statuses[len(fakePR.Commits[0].Statuses)-1]
					}
					target.Description = status.Description
					target.State = status.State
					return nil
				},
			}),
			LoggerClient: mock.FakeLoggerClient(),
		}, []string{}...)

		It("Should create tide status", func() {
			Expect(tide.Do(context.TODO(), plugins.GitCommentEvent{
				Repo: plugins.GitRepo{Name: "foo_repo", Owner: plugins.GitUser{Name: "foo_owner"}},
			})).Should(Succeed())
			Expect(merged).Should(BeFalse())

			Eventually(func(g Gomega) {
				globalLock.RLock()
				defer globalLock.RUnlock()

				g.Expect(len(fakePR.Commits)).Should(Equal(1))
				g.Expect(len(fakePR.Commits[0].Statuses)).Should(Equal(1))
				g.Expect(fakePR.Commits[0].Statuses[0].Context).Should(Equal("tide"))
				g.Expect(fakePR.Commits[0].Statuses[0].State).Should(Equal("PENDING"))
				g.Expect(fakePR.Commits[0].Statuses[0].Description).Should(Equal("Not mergeable. Needs approved, lgtm label."))
			}, 5*time.Second, time.Second).Should(Succeed())
		})
		It("Should change desc", func() {
			globalLock.Lock()
			fakePR.Labels = []plugins.Label{{Name: "lgtm"}}
			globalLock.Unlock()

			Eventually(func(g Gomega) {
				globalLock.RLock()
				defer globalLock.RUnlock()
				g.Expect(fakePR.Commits[0].Statuses[0].Description).Should(Equal("Not mergeable. Needs approved label."))
			}, 5*time.Second, time.Second).Should(Succeed())
		})

		It("Should change state", func() {
			globalLock.Lock()
			fakePR.Labels = []plugins.Label{{Name: "lgtm"}, {Name: "approved"}}
			globalLock.Unlock()

			Eventually(func(g Gomega) {
				globalLock.RLock()
				defer globalLock.RUnlock()
				g.Expect(fakePR.Commits[0].Statuses[0].State).Should(Equal("SUCCESS"))
			}, 5*time.Second, time.Second).Should(Succeed())
		})

		It("Should not create more status", func() {
			globalLock.Lock()
			callCreateStatus = false
			globalLock.Unlock()

			Expect(wait.PollImmediate(time.Second, 5*time.Second, func() (done bool, err error) {
				globalLock.RLock()
				defer globalLock.RUnlock()
				return callCreateStatus == true, nil
			})).ShouldNot(BeNil())
		})

		It("Should merge", func() {
			Eventually(func(g Gomega) {
				globalLock.RLock()
				defer globalLock.RUnlock()
				g.Expect(merged).Should(BeTrue())
			}, 5*time.Second, time.Second).Should(Succeed())
		})
	})
})
