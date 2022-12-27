package internal

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/go-logr/logr"
	"github.com/spf13/pflag"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/wait"

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

func init() {
	plugins.RegisterGitCommentPlugin("tide", func(cs plugins.ClientSets, args ...string) plugins.GitCommentPlugin {
		plugin := tidePlugin{
			issueClient:  cs.GitIssueClient,
			repoClient:   cs.GitRepoClient,
			prClient:     cs.GitPRClient,
			searchClient: cs.GitSearchClient,
			loggerClient: cs.LoggerClient,
			flags:        pflag.NewFlagSet("tide", pflag.ContinueOnError),
		}
		plugin.flags.StringSliceVar(&plugin.required, "required-labels", []string{"lgtm", "approved"}, "")
		plugin.flags.StringSliceVar(&plugin.missing, "missing-labels", []string{
			"needs-rebase", "do-not-merge/hold", "do-not-merge/work-in-progress", "do-not-merge/invalid-owners-file",
		}, "")
		plugin.flags.Parse(args)
		return &tideGitCommentPlugin{tidePlugin: plugin}
	})

	plugins.RegisterGitPRPlugin("tide", func(cs plugins.ClientSets, args ...string) plugins.GitPRPlugin {
		plugin := tidePlugin{
			issueClient:  cs.GitIssueClient,
			repoClient:   cs.GitRepoClient,
			prClient:     cs.GitPRClient,
			searchClient: cs.GitSearchClient,
			loggerClient: cs.LoggerClient,
			flags:        pflag.NewFlagSet("tide", pflag.ContinueOnError),
		}
		plugin.flags.StringSliceVar(&plugin.required, "required-labels", []string{"lgtm", "approved"}, "")
		plugin.flags.StringSliceVar(&plugin.missing, "missing-labels", []string{
			"needs-rebase", "do-not-merge/hold", "do-not-merge/work-in-progress", "do-not-merge/invalid-owners-file",
		}, "")
		plugin.flags.StringVar(&plugin.mergeMethod, "merge-method", "merge", "merge, squash orrebase")
		plugin.flags.Parse(args)
		return &tideGitPRPlugin{tidePlugin: plugin}
	})
}

type tideGitCommentPlugin struct {
	tidePlugin
}

func (p *tideGitCommentPlugin) Do(ctx context.Context, e plugins.GitCommentEvent) error {
	return p.tidePlugin.enableRepo(e.Repo)
}

type tideGitPRPlugin struct {
	tidePlugin
}

func (p *tideGitPRPlugin) Do(ctx context.Context, e plugins.GitPREvent) error {
	return p.tidePlugin.enableRepo(e.Repo)
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
	flags       *pflag.FlagSet
}

func (p *tidePlugin) Name() string {
	return "tide"
}

func (p *tidePlugin) enableRepo(repo plugins.GitRepo) error {
	tide.start(repo, &tideContext{
		RepoClient:     p.repoClient,
		PRClient:       p.prClient,
		SearchClient:   p.searchClient,
		Repo:           repo,
		RequiredLabels: p.required,
		MissingLabels:  p.missing,
		Log: p.loggerClient.GetLogger().WithName("tide_context").
			WithValues("repo", repo.Name).WithValues("owner", repo.Owner.Name),
	})
	return nil
}

var tide *tideController = &tideController{
	contexts: sync.Map{},
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
	contexts sync.Map
}

func (c *tideController) start(repo plugins.GitRepo, tideCtx *tideContext) {
	var tideCtxVal *atomic.Value
	v, ok := c.contexts.Load(repo)
	if ok {
		tideCtxVal = v.(*atomic.Value)
		tideCtxVal.Store(tideCtx)
		return
	}

	tideCtx.Log.Info("Start tide controller")
	tideCtxVal = &atomic.Value{}
	tideCtxVal.Store(tideCtx)
	c.contexts.Store(repo, tideCtxVal)
	go c.syncRepo(tideCtxVal)
}

func (c *tideController) syncRepo(tideCtxVal *atomic.Value) {
	wait.Forever(func() {
		v := tideCtxVal.Load()
		tideCtx, ok := v.(*tideContext)
		if !ok {
			return
		}
		tideCtx.Log.Info("Start tide sync")
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		prs, err := tideCtx.SearchClient.SearchPR(ctx, tideCtx.Repo, plugins.PullRequestStateOpen)
		if err != nil {
			tideCtx.Log.Error(err, "Failed to search pr")
			return
		}
		for _, pr := range prs {
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
					continue
				}
			}
			if !merge {
				continue
			}
			if err := tideCtx.PRClient.MergePR(ctx, tideCtx.Repo, pr.Number, tideCtx.MergeMethod); err != nil {
				tideCtx.Log.Error(err, "Failed to merge pr", "pr", pr.Number)
			}
		}
	}, TideSyncInterval)
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
		desc := fmt.Sprintf("%s. Needs %s label.", statusNotInPool, strings.Join(diff.UnsortedList(), ", "))
		return plugins.GitStatusStatePending, desc, false
	}
	if intersec := missingSets.Intersection(currentLabels); intersec.Len() > 0 {
		desc := fmt.Sprintf("%s. Should not have %s label.", statusNotInPool, strings.Join(intersec.UnsortedList(), ", "))
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
