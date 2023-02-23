package internal

import (
	"context"
	"regexp"
	"strings"

	"github.com/go-logr/logr"
	"github.com/spf13/pflag"
	"k8s.io/apimachinery/pkg/util/sets"

	"github.com/airconduct/kuilei/pkg/plugins"
)

var (
	lgtmRegex       = regexp.MustCompile(`(?m)^/lgtm\s*(.*?)\s*$`)
	lgtmCancelRegex = regexp.MustCompile(`(?m)^/lgtm\s*(cancel)\s*$`)
)

func init() {
	plugins.RegisterGitCommentPlugin("lgtm", func(cs plugins.ClientSets) plugins.GitCommentPlugin {
		plugin := &lgtmPlugin{
			issueClient: cs.GitIssueClient,
			prClient:    cs.GitPRClient,
			ownerClient: cs.OwnersClient,
			logger:      cs.LoggerClient.GetLogger().WithName("lgtm_plugin"),
		}
		return plugin
	})
}

type lgtmPlugin struct {
	issueClient plugins.GitIssueClient
	prClient    plugins.GitPRClient
	ownerClient plugins.OwnersClient
	logger      logr.Logger

	allowAuthor bool
}

func (lp *lgtmPlugin) Name() string {
	return "lgtm"
}

func (lp *lgtmPlugin) Description() string {
	return "Adds or removes the 'lgtm' label which is typically used to gate merging."
}

func (lp *lgtmPlugin) Usage() string {
	return "/lgtm [cancel]"
}

func (lp *lgtmPlugin) BindFlags(flags *pflag.FlagSet) {
	flags.BoolVar(&lp.allowAuthor, "allow-author", false, "Whether allow author to add lgtm")
}

func (lp *lgtmPlugin) Do(ctx context.Context, e plugins.GitCommentEvent) error {
	if !e.IsPR || e.Action != plugins.GitCommentActionCreated {
		return nil
	}
	// Check body
	bodyClean := plugins.CleanMarkdownComments(e.Body)
	lgtmMatch := lgtmRegex.MatchString(bodyClean)
	lgtmCancelMatch := lgtmCancelRegex.MatchString(bodyClean)
	if !lgtmMatch && !lgtmCancelMatch {
		return nil
	}

	// Check author
	if !lp.allowAuthor {
		pr, err := lp.prClient.GetPR(ctx, e.Repo, e.Number)
		if err != nil {
			return err
		}
		if pr.User.Name == e.User.Name {
			resp := "you cannot LGTM your own PR."
			return lp.issueClient.CreateIssueComment(ctx, e.Repo, plugins.GitIssue{Number: e.Number}, plugins.GitIssueComment{
				Body: plugins.FormatResponseRaw(e.Body, e.HTMLURL, e.User.Name, resp),
			})
		}
	}

	// Check owners config
	reviewers := sets.NewString()
	files, err := lp.prClient.ListFiles(ctx, e.Repo, plugins.GitPullRequest{Number: e.Number})
	if err != nil {
		return err
	}
	for _, file := range files {
		owner, err := lp.ownerClient.GetOwners(e.Repo.Owner.Name, e.Repo.Name, file.Path)
		if err != nil {
			return err
		}
		for _, name := range owner.Reviewers {
			reviewers.Insert(strings.ToLower(name))
		}
		for _, name := range owner.Approvers {
			reviewers.Insert(strings.ToLower(name))
		}
	}
	if !reviewers.Has(strings.ToLower(e.User.Name)) {
		lp.logger.Info("No reviewer matched", "user", e.User.Name, "reviewers", reviewers)
		resp := "adding LGTM is restricted to approvers and reviewers in OWNERS files."
		return lp.issueClient.CreateIssueComment(ctx, e.Repo, plugins.GitIssue{Number: e.Number}, plugins.GitIssueComment{
			Body: plugins.FormatResponseRaw(e.Body, e.HTMLURL, e.User.Name, resp),
		})
	}
	if lgtmCancelMatch {
		return lp.issueClient.RemoveLabel(ctx, e.Repo, plugins.GitIssue{Number: e.Number}, plugins.Label{Name: "lgtm"})
	}
	return lp.issueClient.AddLabel(ctx, e.Repo, plugins.GitIssue{Number: e.Number}, []plugins.Label{{Name: "lgtm"}})
}
