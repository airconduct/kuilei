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
	// Tide plugin should handle GitCommentEvent
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
	// Tide plugin should handle GitPREvent
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

// tideGitCommentPlugin is a plugin to handle tide related GitCommentEvent
type tideGitCommentPlugin struct {
	tidePlugin
}

func (p *tideGitCommentPlugin) Do(ctx context.Context, e plugins.GitCommentEvent) error {
	if e.IsPR {
		// If this comment from a PR, we should enqueue the PR to tide controller
		p.tidePlugin.enqueue(tidePRKey{GitRepo: e.Repo})
	}
	return nil
}

// tideGitPRPlugin is a plugin to handle tide related GitPREvent
type tideGitPRPlugin struct {
	tidePlugin
}

func (p *tideGitPRPlugin) Do(ctx context.Context, e plugins.GitPREvent) error {
	// Enqueue the PR to tide controller
	p.tidePlugin.enqueue(tidePRKey{GitRepo: e.Repo})
	return nil
}

// tidePlugin is a plugin to handle tide related GitEvent.
// Compared to the other two plugins, this plugin is used to handle more general events.
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
	// Set logger for tide controller
	setLoggerOnce.Do(func() {
		tide.logger = p.loggerClient.GetLogger().WithName("tide_controller")
	})
	// Ensure the tide controller has started
	tide.start(key, &tideContext{
		RepoClient:     p.repoClient,
		PRClient:       p.prClient,
		SearchClient:   p.searchClient,
		Repo:           key.GitRepo,
		RequiredLabels: p.required,
		MissingLabels:  p.missing,
		Log: p.loggerClient.GetLogger().WithName("tide_context").
			WithValues("repo", key.Name).WithValues("owner", key.Owner.Name),
	})
}

var tide *tideController = &tideController{
	queue: workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "tideContextQueue"),
}

// tidePRKey is the key used to identify a group of PRs
type tidePRKey struct {
	plugins.GitRepo
}

func (k tidePRKey) String() string {
	return fmt.Sprintf("repo: %s/%s", k.Owner.Name, k.Name)
}

// tideResult is the result of a tide operation.
type tideResult struct {
	Requeue      bool
	RequeueAfter time.Duration
}

// tideContext is the context used by tide controller.
// Every tideContext corresponds to one repo.
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

// tideController is the controller to handle tide related events.
// It will enqueue a tidePRKey when it receives a tide related event.
//
// The steps of the work process are:
//  1. Get the tidePRKey from the queue
//  2. Get the tide context from the context store via the tidePRKey
//  3. Get all prs that correspond to the tidePRKey
//  4. Handle the prs, create `tide` status for each pr
//  5. Check if each pr can be merged, if yes, merge it
//  6. Requeue the key if any error occurred or at least one pr is not merged
type tideController struct {
	startOnce sync.Once

	logger       logr.Logger
	contextStore tideContextStore
	queue        workqueue.RateLimitingInterface
}

func (c *tideController) start(key tidePRKey, tideCtx *tideContext) {
	tideCtx.Log.Info("Get tide event", "key", key)
	c.queue.Add(key)
	c.contextStore.Set(key.GitRepo, tideCtx)

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
				// Search all repos that need to be handled
				repos := []plugins.GitRepo{}
				c.contextStore.contexts.Range(func(key, value any) bool {
					repos = append(repos, key.(plugins.GitRepo))
					return true
				})
				for _, repo := range repos {
					// Get the tide context from the context store
					tideCtx := c.contextStore.Get(repo)
					// Skip the repo if the tide context is nil
					if tideCtx == nil {
						continue
					}
					c.logger.Info("Enqueue repo by search", "repo", repo)
					c.queue.Add(tidePRKey{GitRepo: repo})
				}
			}, TideSyncInterval)
		}()
	})
}

// work is the work process of tide controller.
func (c *tideController) work() bool {
	v, shutdown := c.queue.Get()
	if shutdown {
		return false
	}
	defer c.queue.Done(v)

	key := v.(tidePRKey)
	ctx, cancle := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancle()
	// Do pr sync
	result, err := c.syncOnce(ctx, key)
	if err != nil {
		// Requeue the pr if any error occurred
		c.logger.Error(err, "Failed to sync tide")
		c.queue.AddRateLimited(key)
		return true
	}
	// Requeue the pr if needed
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

// syncOnce syncs a group of prs in one repo.
func (c *tideController) syncOnce(ctx context.Context, key tidePRKey) (tideResult, error) {
	c.logger.Info("Start to handel pr", "key", key)
	// Get tide context from context cache
	tideCtx := c.contextStore.Get(key.GitRepo)
	if tideCtx == nil {
		return tideResult{}, fmt.Errorf("tide context not found, key: %s", key)
	}
	// Search all prs need to be handled
	results, err := tideCtx.SearchClient.SearchPR(ctx, key.GitRepo, plugins.PullRequestStateOpen)
	if err != nil {
		return tideResult{}, fmt.Errorf("failed to search prs, %w", err)
	}
	// Handle all prs
	requeue := false
	var after time.Duration
	for _, prResult := range results {
		// Get head commit of the pr
		commit, ok := getHeadCommit(prResult.GitPullRequest.Head, prResult.Commits)
		// If the commit not found, skip handling
		if !ok {
			tideCtx.Log.Info("Commit not found, skip handling", "pr_number", prResult.GitPullRequest.Number, "sha", prResult.GitPullRequest.Head.SHA)
			continue
		}
		// Handle the pr
		merged, err := c.syncPR(ctx, tideCtx, prResult.GitPullRequest, commit.Statuses, commit.Checks)
		if err != nil {
			return tideResult{}, fmt.Errorf("failed to sync pr, %w", err)
		}
		// Requeue the pr if not merged
		if !merged {
			requeue = true
			after = 30 * time.Second
		}
	}
	return tideResult{Requeue: requeue, RequeueAfter: after}, nil
}

// syncPR handle
func (c *tideController) syncPR(
	ctx context.Context, tideCtx *tideContext,
	pr plugins.GitPullRequest,
	statuses []plugins.GitCommitStatus,
	checks []plugins.GitCommitCheck,
) (merged bool, err error) {
	// Get `tide` status
	tideStatus, ok := getTideStatus(statuses)
	// Get tide status/description need to be set, and whether the pr can be merged.
	state, desc, merge := wantsStateAndDescription(
		pr, statuses, checks,
		tideCtx.RequiredLabels, tideCtx.MissingLabels,
	)
	// Decide whether to set the status
	//  1. If the status is not found, create the status
	//  2. If the status is not the same as the status need to be set, set the status
	//  3. If the status is the same as the status need to be set, do nothing
	if !ok || (tideStatus.State != state || tideStatus.Description != desc) {
		// Set the status
		err := tideCtx.RepoClient.CreateStatus(ctx, tideCtx.Repo, pr.Head.SHA, plugins.GitCommitStatus{
			State:       state,
			Context:     statusContext,
			Description: desc,
		})
		if err != nil {
			tideCtx.Log.Error(err, "Failed to create status", "pr", pr.Number)
			return false, err
		}
	}
	// If the pr can not be merged, return
	if !merge {
		tideCtx.Log.Info("No need to merge", "pr id", pr.Number)
		return false, nil
	}
	// Merge the pr
	if err := tideCtx.PRClient.MergePR(ctx, tideCtx.Repo, pr.Number, tideCtx.MergeMethod); err != nil {
		tideCtx.Log.Error(err, "Failed to merge pr", "pr", pr.Number)
		return false, err
	}
	return true, nil
}

// getHeadCommit get the head commit of the pr
func getHeadCommit(head plugins.GitBranch, commits []plugins.GitCommit) (plugins.GitCommit, bool) {
	for _, commit := range commits {
		if commit.Sha == head.SHA {
			return commit, true
		}
	}
	return plugins.GitCommit{}, false
}

func getTideStatus(statuses []plugins.GitCommitStatus) (plugins.GitCommitStatus, bool) {
	for _, status := range statuses {
		if status.Context == statusContext {
			return status, true
		}
	}
	return plugins.GitCommitStatus{}, false
}

func wantsStateAndDescription(
	pr plugins.GitPullRequest,
	statuses []plugins.GitCommitStatus,
	checks []plugins.GitCommitCheck,
	required, missing []string,
) (state string, desc string, merge bool) {
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
	for _, check := range checks {
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
	for _, status := range statuses {
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
