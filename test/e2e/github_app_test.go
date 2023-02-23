package e2e

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/airconduct/go-probot"
	"github.com/google/go-github/v48/github"
	"github.com/h2non/gock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

//go:embed testdata/kuilei.yaml
var testConfig string

//go:embed testdata/fake.OWNERS
var testOwners string

var _ = Describe("GitHub App E2E Test", func() {
	setupFixture()
	When("send github issue comment event", func() {
		It("mock help comment and execute help command", func() {
			lock := sync.RWMutex{}
			var body string
			// Mock create comment
			gock.New("https://api.github.com").
				Post("/repos/foo-owner/foo-repo/issues/1/comments").
				AddMatcher(func(r1 *http.Request, r2 *gock.Request) (bool, error) {
					// Get body value
					data := map[string]interface{}{}
					Expect(json.NewDecoder(r1.Body).Decode(&data)).Should(Succeed())
					lock.Lock()
					defer lock.Unlock()
					body = data["body"].(string)
					return true, nil
				}).
				Reply(200)

			Eventually(func(g Gomega) {
				// Send a mock event to this github app
				g.Expect(sendToGitHubApp(
					probot.GitHub.IssueComment.Created,
					github.IssueCommentEvent{
						Action: github.String("created"),
						Issue:  &github.Issue{Number: github.Int(1)},
						Comment: &github.IssueComment{
							ID:   github.Int64(1),
							Body: github.String("/help"),
						},
						Repo: &github.Repository{
							ID:    github.Int64(1),
							Name:  github.String("foo-repo"),
							Owner: &github.User{Login: github.String("foo-owner")},
						},
						Installation: &github.Installation{ID: github.Int64(1)},
					}),
				).Should(Succeed())
			}, 5*time.Second, time.Second).Should(Succeed())

			Eventually(func(g Gomega) {
				// Check if body is correct
				lock.RLock()
				defer lock.RUnlock()
				g.Expect(body).Should(ContainSubstring("Command: **help**"))
			}, 5*time.Second, time.Second).Should(Succeed())
		})

		It("mock label comment and execute label command", func() {
			lock := sync.RWMutex{}
			var labels []string
			// Mock add label
			gock.New("https://api.github.com").
				Post("/repos/foo-owner/foo-repo/issues/1/labels").
				AddMatcher(func(r1 *http.Request, r2 *gock.Request) (bool, error) {
					// Get labels value
					lock.Lock()
					defer lock.Unlock()
					Expect(json.NewDecoder(r1.Body).Decode(&labels)).Should(Succeed())
					return true, nil
				}).
				Reply(200)

			Eventually(func(g Gomega) {
				// Send a mock event to this github app
				g.Expect(sendToGitHubApp(
					probot.GitHub.IssueComment,
					github.IssueCommentEvent{
						Action: github.String("created"),
						Issue: &github.Issue{
							Number:           github.Int(1),
							PullRequestLinks: &github.PullRequestLinks{},
						},
						Comment: &github.IssueComment{
							ID:   github.Int64(1),
							Body: github.String("/label foo"),
						},
						Repo: &github.Repository{
							ID:    github.Int64(1),
							Name:  github.String("foo-repo"),
							Owner: &github.User{Login: github.String("foo-owner")},
						},
						Installation: &github.Installation{ID: github.Int64(1)},
					}),
				).Should(Succeed())
			}, 5*time.Second, time.Second).Should(Succeed())

			Eventually(func(g Gomega) {
				// Check if labels is correct
				lock.RLock()
				defer lock.RUnlock()
				g.Expect(labels).Should(ContainElements("foo"))
			}, 5*time.Second, time.Second).Should(Succeed())
		})

		It("mock lgtm comment and execute lgtm command", func() {
			lock := sync.RWMutex{}
			var labels []string
			// Mock add label
			gock.New("https://api.github.com").
				Post("/repos/foo-owner/foo-repo/issues/1/labels").
				AddMatcher(func(r1 *http.Request, r2 *gock.Request) (bool, error) {
					// Get labels value
					lock.Lock()
					defer lock.Unlock()
					Expect(json.NewDecoder(r1.Body).Decode(&labels)).Should(Succeed())
					return true, nil
				}).
				Reply(200)
			// Mock list pr files
			gock.New("https://api.github.com").
				Get("/repos/foo-owner/foo-repo/pulls/1/files").
				Reply(200).
				JSON([]github.CommitFile{{Filename: github.String("README.md")}})

			Eventually(func(g Gomega) {
				// Send a mock event to this github app
				g.Expect(sendToGitHubApp(
					probot.GitHub.IssueComment,
					github.IssueCommentEvent{
						Action: github.String("created"),
						Issue: &github.Issue{
							Number:           github.Int(1),
							PullRequestLinks: &github.PullRequestLinks{},
						},
						Comment: &github.IssueComment{
							ID:   github.Int64(1),
							Body: github.String("/lgtm"),
							User: &github.User{Login: github.String("foo-user")},
						},
						Repo: &github.Repository{
							ID:    github.Int64(1),
							Name:  github.String("foo-repo"),
							Owner: &github.User{Login: github.String("foo-owner")},
						},
						Installation: &github.Installation{ID: github.Int64(1)},
					}),
				).Should(Succeed())
			}, 5*time.Second, time.Second).Should(Succeed())

			Eventually(func(g Gomega) {
				// Check if labels is correct
				lock.RLock()
				defer lock.RUnlock()
				g.Expect(labels).Should(ContainElements("lgtm"))
			}, 5*time.Second, time.Second).Should(Succeed())
		})

		It("mock approve comment and execute approve command", func() {
			lock := sync.RWMutex{}
			var labels []string
			// Mock add label
			gock.New("https://api.github.com").
				Post("/repos/foo-owner/foo-repo/issues/1/labels").
				AddMatcher(func(r1 *http.Request, r2 *gock.Request) (bool, error) {
					// Get labels value
					lock.Lock()
					defer lock.Unlock()
					Expect(json.NewDecoder(r1.Body).Decode(&labels)).Should(Succeed())
					return true, nil
				}).
				Reply(200)
			// Mock list pr files
			gock.New("https://api.github.com").
				Get("/repos/foo-owner/foo-repo/pulls/1/files").
				Reply(200).
				JSON([]github.CommitFile{{Filename: github.String("README.md")}})

			Eventually(func(g Gomega) {
				// Send a mock event to this github app
				g.Expect(sendToGitHubApp(
					probot.GitHub.IssueComment,
					github.IssueCommentEvent{
						Action: github.String("created"),
						Issue: &github.Issue{
							Number:           github.Int(1),
							PullRequestLinks: &github.PullRequestLinks{},
						},
						Comment: &github.IssueComment{
							ID:   github.Int64(1),
							Body: github.String("/approve"),
							User: &github.User{Login: github.String("foo-user")},
						},
						Repo: &github.Repository{
							ID:    github.Int64(1),
							Name:  github.String("foo-repo"),
							Owner: &github.User{Login: github.String("foo-owner")},
						},
						Installation: &github.Installation{ID: github.Int64(1)},
					}),
				).Should(Succeed())
			}, 5*time.Second, time.Second).Should(Succeed())

			Eventually(func(g Gomega) {
				// Check if labels is correct
				lock.RLock()
				defer lock.RUnlock()
				g.Expect(labels).Should(ContainElements("approved"))
			}, 5*time.Second, time.Second).Should(Succeed())
		})

		It("test tide plugin", func() {
			lock := sync.RWMutex{}
			var tideStatus = map[string]interface{}{}
			merged := false
			// prSearchResults := pluginhelpers.PRSearchQuery{}
			// Mock GraphQL query
			data := map[string]interface{}{
				"data": map[string]interface{}{
					"repository": map[string]interface{}{
						"pullRequests": map[string]interface{}{
							"nodes": []interface{}{
								map[string]interface{}{
									"number":      1,
									"state":       "OPEN",
									"headRefName": "fooref",
									"headRefOid":  "foosha",
									"mergeable":   "MERGEABLE",
									"labels": map[string]interface{}{
										"nodes": []interface{}{
											map[string]interface{}{
												"name": "lgtm",
											},
											map[string]interface{}{
												"name": "approved",
											},
										},
									},
									"commits": map[string]interface{}{
										"nodes": []interface{}{
											map[string]interface{}{
												"commit": map[string]interface{}{
													"oid": "foosha",
													// "statusCheckRollup": ,
													"status": map[string]interface{}{
														"contexts": []interface{}{
															tideStatus,
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			}
			// Mock GraphQL query
			gock.New("https://api.github.com").
				Post("/graphql").
				Reply(200).JSON(data)
			// Mock create status
			gock.New("https://api.github.com").
				Post("/repos/foo-owner/foo-repo/statuses/foosha").
				Times(100).
				AddMatcher(func(r1 *http.Request, r2 *gock.Request) (bool, error) {
					// Get status value
					lock.Lock()
					defer lock.Unlock()
					json.NewDecoder(r1.Body).Decode(&tideStatus)
					delete(tideStatus, "target_url")
					tideStatus["state"] = strings.ToUpper(tideStatus["state"].(string))
					// Re-add mock GraphQL query to return refresh the tide status
					gock.New("https://api.github.com").
						Post("/graphql").
						Reply(200).JSON(data)
					return true, nil
				}).
				Reply(200)
			// Mock merge pr
			gock.New("https://api.github.com").
				Put("/repos/foo-owner/foo-repo/pulls/1/merge").
				AddMatcher(func(r1 *http.Request, r2 *gock.Request) (bool, error) {
					lock.Lock()
					defer lock.Unlock()
					merged = true
					return true, nil
				}).
				Reply(200)

			Eventually(func(g Gomega) {
				// Send a mock event to this github app
				g.Expect(sendToGitHubApp(
					probot.GitHub.PullRequest.Opened,
					github.PullRequestEvent{
						Action: github.String("opened"),
						Number: github.Int(1),
						PullRequest: &github.PullRequest{
							ID:     github.Int64(1),
							Number: github.Int(1),
							State:  github.String("open"),
						},
						Repo: &github.Repository{
							ID:    github.Int64(1),
							Name:  github.String("foo-repo"),
							Owner: &github.User{Login: github.String("foo-owner")},
						},
						Installation: &github.Installation{ID: github.Int64(1)},
					}),
				).Should(Succeed())
			}, 5*time.Second, time.Second).Should(Succeed())

			Eventually(func(g Gomega) {
				lock.RLock()
				defer lock.RUnlock()
				// Check if tide status has been already created
				g.Expect(tideStatus["context"]).Should(Equal("tide"))
				// Check if pr has been merged
				g.Expect(merged).Should(BeTrue())
			}, 5*time.Second, time.Second).Should(Succeed())
		})
	})
})

func setupFixture() {
	installationID := 1
	// Mock get token
	gock.New("https://api.github.com").
		Post(fmt.Sprintf("/app/installations/%d/access_tokens", installationID)).
		Persist().
		Reply(200).JSON(map[string]interface{}{
		"token": "test",
		"permissions": map[string]interface{}{
			"issues": "write",
		}})
	// Mock get configuration
	gock.New("https://api.github.com").
		Get("/repos/foo-owner/foo-repo/contents/.github/kuilei.yml").
		Persist().
		Reply(200).JSON(map[string]interface{}{
		"name":    "kuilei.yml",
		"path":    ".github/kuilei.yml",
		"sha":     "51ff363192be5925c5c24d96a2941c271c5060d0",
		"size":    186,
		"type":    "file",
		"content": testConfig,
	})
	// Mock get owners configuration
	gock.New("https://api.github.com").
		Get("/repos/foo-owner/foo-repo/contents/OWNERS").
		Persist().
		Reply(200).JSON(map[string]interface{}{
		"name":    "OWNERS",
		"path":    "OWNERS",
		"type":    "file",
		"content": testOwners,
	})
	// Mock search owners file
	gock.New("https://api.github.com").
		Get("/search/code").
		Persist().
		MatchParams(map[string]string{
			// The encoding result of this query is:
			// q=repo%3Afoo-owner%2Ffoo-repo+filename%3AOWNERS
			// The space(" ") is encoded to "+".
			//
			// refence: https://stackoverflow.com/questions/2678551/when-should-space-be-encoded-to-plus-or-20
			"q": "repo:foo-owner/foo-repo filename:OWNERS",
		}).
		Reply(200).JSON(map[string]interface{}{
		"total_count": 1,
		"items": []map[string]interface{}{
			{
				"name": "OWNERS",
				"path": "OWNERS",
			},
		},
	})
}
