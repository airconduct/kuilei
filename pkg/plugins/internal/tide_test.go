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
		fakePR := &plugins.GitPullRequest{
			Number:    1,
			State:     plugins.PullRequestStateOpen,
			Head:      plugins.GitBranch{SHA: "foo"},
			Locked:    false,
			Title:     "foo",
			Body:      "foo",
			Mergeable: plugins.GitMergeableStateMergeable,
			Labels:    []plugins.Label{},
		}
		statuses := []plugins.GitCommitStatus{}
		checks := []plugins.GitCommitCheck{}
		callCreateStatus := false

		tide := plugins.GetGitCommentPlugin("tide", plugins.ClientSets{
			GitPRClient: mock.FakeGitPRClient(
				nil, func(ctx context.Context, repo plugins.GitRepo, number int) (plugins.GitPullRequest, error) {
					globalLock.Lock()
					defer globalLock.Unlock()
					return *fakePR, nil
				},
				func(ctx context.Context, repo plugins.GitRepo, number int, method string) error {
					globalLock.Lock()
					defer globalLock.Unlock()
					fakePR.State = plugins.PullRequestStateMerged
					return nil
				},
			),
			GitSearchClient: mock.FakeSearchClient(map[string]interface{}{
				"SearchPR": func(ctx context.Context, repo plugins.GitRepo, state string) ([]plugins.GitPullRequestSearchResult, error) {
					globalLock.RLock()
					defer globalLock.RUnlock()
					commit := plugins.GitCommit{Sha: "foo"}
					commit.Statuses = append(commit.Statuses, statuses...)
					commit.Checks = append(commit.Checks, checks...)
					return []plugins.GitPullRequestSearchResult{{
						GitPullRequest: *fakePR, Commits: []plugins.GitCommit{commit},
					}}, nil
				},
			}),
			GitRepoClient: mock.FakeRepoClient(map[string]interface{}{
				"ListStatuses": func(ctx context.Context, repo plugins.GitRepo, ref string) ([]plugins.GitCommitStatus, error) {
					globalLock.Lock()
					defer globalLock.Unlock()
					return statuses, nil
				},
				"ListChecks": func(ctx context.Context, repo plugins.GitRepo, ref string) ([]plugins.GitCommitCheck, error) {
					globalLock.Lock()
					defer globalLock.Unlock()
					return checks, nil
				},
				"CreateStatus": func(ctx context.Context, repo plugins.GitRepo, ref string, status plugins.GitCommitStatus) error {
					globalLock.Lock()
					defer globalLock.Unlock()
					callCreateStatus = true

					var target *plugins.GitCommitStatus
					for idx, s := range statuses {
						if s.Context == status.Context {
							target = &statuses[idx]
						}
					}
					if target == nil {
						statuses = append(statuses, plugins.GitCommitStatus{Context: status.Context})
						target = &statuses[len(statuses)-1]
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
				GitComment: plugins.GitComment{
					Number: 1, IsPR: true,
				},
				Repo: plugins.GitRepo{Name: "foo_repo", Owner: plugins.GitUser{Name: "foo_owner"}},
			})).Should(Succeed())
			Expect(fakePR.State).Should(Equal(plugins.PullRequestStateOpen))

			Eventually(func(g Gomega) {
				globalLock.RLock()
				defer globalLock.RUnlock()

				g.Expect(len(statuses)).Should(Equal(1))
				g.Expect(statuses[0].Context).Should(Equal("tide"))
				g.Expect(statuses[0].State).Should(Equal("PENDING"))
				g.Expect(statuses[0].Description).Should(Equal("Not mergeable. Needs approved, lgtm label."))
			}, 5*time.Second, time.Second).Should(Succeed())
		})
		It("Should change desc", func() {
			globalLock.Lock()
			fakePR.Labels = []plugins.Label{{Name: "lgtm"}}
			globalLock.Unlock()

			Eventually(func(g Gomega) {
				globalLock.RLock()
				defer globalLock.RUnlock()
				g.Expect(statuses).Should(HaveLen(1))
				g.Expect(statuses[0].Description).Should(Equal("Not mergeable. Needs approved label."))
			}, 5*time.Second, time.Second).Should(Succeed())
		})

		It("Should change state", func() {
			globalLock.Lock()
			fakePR.Labels = []plugins.Label{{Name: "lgtm"}, {Name: "approved"}}
			globalLock.Unlock()

			Eventually(func(g Gomega) {
				globalLock.RLock()
				defer globalLock.RUnlock()
				g.Expect(statuses).Should(HaveLen(1))
				g.Expect(statuses[0].State).Should(Equal("SUCCESS"))
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
				g.Expect(fakePR.State).Should(Equal(plugins.PullRequestStateMerged))
			}, 5*time.Second, time.Second).Should(Succeed())
		})
	})
})
