package internal

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/go-logr/logr"
	"github.com/spf13/pflag"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/util/workqueue"

	"github.com/airconduct/kuilei/pkg/plugins"
)

var TideSyncInterval = time.Minute

const (
	statusContext = "tide"
	statusInPool  = "In merge pool."
	// statusNotInPool is a format string used when a PR is not in a tide pool.
	// The '%s' field is populated with the reason why the PR is not in a
	// tide pool or the empty string if the reason is unknown. See requirementDiff.
	statusNotInPool = "Not mergeable"
)

var setLoggerOnce sync.Once

func init() {
	plugins.RegisterGitCommentPlugin("tide", func(cs plugins.ClientSets) plugins.GitCommentPlugin {
		plugin := tidePlugin{
			issueClient:  cs.GitIssueClient,
			repoClient:   cs.GitRepoClient,
			prClient:     cs.GitPRClient,
			searchClient: cs.GitSearchClient,
			loggerClient: cs.LoggerClient,
		}
		return &tideGitCommentPlugin{tidePlugin: plugin}
	})

	plugins.RegisterGitPRPlugin("tide", func(cs plugins.ClientSets) plugins.GitPRPlugin {
		plugin := tidePlugin{
			issueClient:  cs.GitIssueClient,
			repoClient:   cs.GitRepoClient,
			prClient:     cs.GitPRClient,
			searchClient: cs.GitSearchClient,
			loggerClient: cs.LoggerClient,
		}
		return &tideGitPRPlugin{tidePlugin: plugin}
	})
}

type tideGitCommentPlugin struct {
	tidePlugin
}

func (p *tideGitCommentPlugin) Do(ctx context.Context, e plugins.GitCommentEvent) error {
	if e.IsPR {
		p.tidePlugin.enqueue(tidePRKey{Repo: e.Repo, Number: e.Number})
	}
	return nil
}

type tideGitPRPlugin struct {
	tidePlugin
}

func (p *tideGitPRPlugin) Do(ctx context.Context, e plugins.GitPREvent) error {
	p.tidePlugin.enqueue(tidePRKey{Repo: e.Repo, Number: e.Number})
	return nil
}

type tidePlugin struct {
	issueClient  plugins.GitIssueClient
	repoClient   plugins.GitRepoClient
	prClient     plugins.GitPRClient
	searchClient plugins.GitSearchClient
	loggerClient plugins.LoggerClient

	required    []string
	missing     []string
	mergeMethod string
}

func (p *tidePlugin) Name() string {
	return "tide"
}

func (lp *tidePlugin) Description() string {
	return "Managing a pool of GitHub PRs that match a given set of criteria. It will automatically retest PRs that meet the criteria (“tide comes in”) and automatically merge them when they have up-to-date passing test results (“tide goes out”)."
}

func (lp *tidePlugin) Usage() string {
	return "Add 'tide' plugin in configuration located under [.github/kuilei.yml](/.github/kuilei.yml)"
}

func (lp *tidePlugin) BindFlags(flags *pflag.FlagSet) {
	flags.StringSliceVar(&lp.required, "required-labels", []string{"lgtm", "approved"}, "Do not merge prs without required-labels")
	flags.StringSliceVar(&lp.missing, "missing-labels", []string{
		"needs-rebase", "do-not-merge/hold", "do-not-merge/work-in-progress", "do-not-merge/invalid-owners-file",
	}, "Do not merge prs with missing-labels")
	flags.StringVar(&lp.mergeMethod, "merge-method", "merge", "Merge method: merge | squash | rebase")
}

func (p *tidePlugin) enqueue(key tidePRKey) {
	setLoggerOnce.Do(func() {
		tide.logger = p.loggerClient.GetLogger().WithName("tide_controller")
	})
	tide.start(key, &tideContext{
		RepoClient:     p.repoClient,
		PRClient:       p.prClient,
		SearchClient:   p.searchClient,
		Repo:           key.Repo,
		RequiredLabels: p.required,
		MissingLabels:  p.missing,
		Log: p.loggerClient.GetLogger().WithName("tide_context").
			WithValues("repo", key.Repo.Name).WithValues("owner", key.Repo.Owner.Name),
	})
}

var tide *tideController = &tideController{
	queue: workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "tideContextQueue"),
}

type tidePRKey struct {
	Repo   plugins.GitRepo
	Number int
}

func (k tidePRKey) String() string {
	return fmt.Sprintf("id: %d, repo: %s/%s", k.Number, k.Repo.Owner.Name, k.Repo.Name)
}

type tideResult struct {
	Requeue      bool
	RequeueAfter time.Duration
}

type tideContext struct {
	RepoClient     plugins.GitRepoClient
	PRClient       plugins.GitPRClient
	SearchClient   plugins.GitSearchClient
	Repo           plugins.GitRepo
	Log            logr.Logger
	RequiredLabels []string
	MissingLabels  []string
	MergeMethod    string
}

type tideController struct {
	startOnce sync.Once

	logger       logr.Logger
	contextStore tideContextStore
	queue        workqueue.RateLimitingInterface
}

func (c *tideController) start(key tidePRKey, tideCtx *tideContext) {
	tideCtx.Log.Info("Get tide event", "key", key)
	c.queue.Add(key)
	c.contextStore.Set(key.Repo, tideCtx)

	c.startOnce.Do(func() {
		// Start work process
		go func() {
			wait.Forever(func() {
				for c.work() {
				}
			}, TideSyncInterval)
		}()
		// Start list process
		go func() {
			wait.Forever(func() {
				repos := []plugins.GitRepo{}
				c.contextStore.contexts.Range(func(key, value any) bool {
					repos = append(repos, key.(plugins.GitRepo))
					return true
				})
				for _, repo := range repos {
					tideCtx := c.contextStore.Get(repo)
					if tideCtx == nil {
						continue
					}
					prs, err := tideCtx.SearchClient.SearchPR(context.TODO(), repo, plugins.PullRequestStateOpen)
					if err != nil {
						c.logger.Error(err, "Failed to search pr", "repo", repo)
						continue
					}
					for _, pr := range prs {
						c.logger.Info("Enqueue pr by search", "repo", repo)
						c.queue.Add(tidePRKey{Repo: repo, Number: pr.Number})
					}
				}
			}, TideSyncInterval)
		}()
	})
}

func (c *tideController) work() bool {
	v, shutdown := c.queue.Get()
	if shutdown {
		return false
	}
	defer c.queue.Done(v)

	key := v.(tidePRKey)
	ctx, cancle := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancle()
	result, err := c.syncOnce(ctx, key)
	if err != nil {
		c.logger.Error(err, "Failed to sync tide")
		c.queue.AddRateLimited(key)
		return true
	}
	if result.Requeue {
		after := result.RequeueAfter
		if after == 0 {
			after = time.Second
		}
		c.logger.Info("Requeue after", "key", key, "after", after)
		c.queue.AddAfter(key, after)
	}
	return true
}

func (c *tideController) syncOnce(ctx context.Context, key tidePRKey) (tideResult, error) {
	c.logger.Info("Start to handel pr", "key", key)
	tideCtx := c.contextStore.Get(key.Repo)
	if tideCtx == nil {
		return tideResult{}, fmt.Errorf("tide context not found, key: %s", key)
	}
	pr, err := tideCtx.PRClient.GetPR(ctx, key.Repo, key.Number)
	if err != nil {
		return tideResult{}, fmt.Errorf("failed to get pr, %w", err)
	}
	if pr.State != plugins.PullRequestStateOpen {
		tideCtx.Log.Info("Skip to handel not open pr", "key", key)
		return tideResult{}, nil
	}
	return c.syncPR(ctx, tideCtx, pr)
}

func (c *tideController) syncPR(ctx context.Context, tideCtx *tideContext, pr plugins.GitPullRequest) (tideResult, error) {
	tideStatus, ok := getTideStatus(pr)
	state, desc, merge := wantsStateAndDescription(pr, tideCtx.RequiredLabels, tideCtx.MissingLabels)
	if !ok || (tideStatus.State != state || tideStatus.Description != desc) {
		err := tideCtx.RepoClient.CreateStatus(ctx, tideCtx.Repo, pr.Head.Sha, plugins.GitCommitStatus{
			State:       state,
			Context:     statusContext,
			Description: desc,
		})
		if err != nil {
			tideCtx.Log.Error(err, "Failed to create status", "pr", pr.Number)
			return tideResult{}, err
		}
	}
	if !merge {
		tideCtx.Log.Info("No need to merge", "pr id", pr.Number)
		return tideResult{Requeue: true}, nil
	}
	if err := tideCtx.PRClient.MergePR(ctx, tideCtx.Repo, pr.Number, tideCtx.MergeMethod); err != nil {
		tideCtx.Log.Error(err, "Failed to merge pr", "pr", pr.Number)
		return tideResult{}, err
	}
	return tideResult{}, nil
}

func getTideStatus(pr plugins.GitPullRequest) (plugins.GitCommitStatus, bool) {
	for _, commit := range pr.Commits {
		if commit.Sha == pr.Head.Sha {
			for _, status := range commit.Statuses {
				if status.Context == statusContext {
					return status, true
				}
			}
		}
	}
	return plugins.GitCommitStatus{}, false
}

func wantsStateAndDescription(pr plugins.GitPullRequest, required, missing []string) (state string, desc string, merge bool) {
	// Check labels
	currentLabels := sets.NewString()
	requiredSets := sets.NewString(required...)
	missingSets := sets.NewString(missing...)
	for _, label := range pr.Labels {
		currentLabels.Insert(label.Name)
	}
	if diff := requiredSets.Difference(currentLabels); diff.Len() > 0 {
		desc := fmt.Sprintf("%s. Needs %s label.", statusNotInPool, strings.Join(diff.List(), ", "))
		return plugins.GitStatusStatePending, desc, false
	}
	if intersec := missingSets.Intersection(currentLabels); intersec.Len() > 0 {
		desc := fmt.Sprintf("%s. Should not have %s label.", statusNotInPool, strings.Join(intersec.List(), ", "))
		return plugins.GitStatusStatePending, desc, false
	}

	// Check jobs
	tideSuccess := false
	unsuccessJobs := []string{}
	for _, commit := range pr.Commits {
		if commit.Sha == pr.Head.Sha {
			for _, check := range commit.Checks {
				// ignore empty check
				if check.Name == "" {
					continue
				}
				if check.Status != plugins.GitCheckStatusCompleted ||
					(check.Conclusion != plugins.GitCheckConclusionStateSuccess &&
						check.Conclusion != plugins.GitCheckConclusionStateNeutral) {
					unsuccessJobs = append(unsuccessJobs, check.Name)
				}
			}
			for _, status := range commit.Statuses {
				// ignore empty check
				if status.Context == "" {
					continue
				}
				// ignore tide context
				if status.Context == statusContext {
					tideSuccess = status.State == plugins.GitStatusStateSuccess
					continue
				}
				if status.State != plugins.GitStatusStateSuccess {
					unsuccessJobs = append(unsuccessJobs, status.Context)
				}
			}
			break
		}
	}
	if len(unsuccessJobs) > 0 {
		desc := fmt.Sprintf("%s. Job %s has not succeeded.", statusNotInPool, strings.Join(unsuccessJobs, ", "))
		return plugins.GitStatusStatePending, desc, false
	}
	// Check merge conflict
	if pr.Mergeable == plugins.GitMergeableStateConflicting {
		return plugins.GitStatusStateError, "PR has a merge conflict.", false
	}
	return plugins.GitStatusStateSuccess, statusInPool, tideSuccess
}

type tideContextStore struct {
	contexts sync.Map
}

func (s *tideContextStore) Get(repo plugins.GitRepo) *tideContext {
	v, ok := s.contexts.Load(repo)
	if !ok {
		return nil
	}
	return v.(*tideContext)
}

func (s *tideContextStore) Set(repo plugins.GitRepo, tctx *tideContext) {
	s.contexts.Store(repo, tctx)
}
