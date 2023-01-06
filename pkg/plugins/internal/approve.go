package internal

import (
	"context"
	"regexp"
	"strings"

	"github.com/spf13/pflag"
	"k8s.io/apimachinery/pkg/util/sets"

	"github.com/airconduct/kuilei/pkg/plugins"
)

var (
	approveRegex       = regexp.MustCompile(`(?m)^/approve\s*(.*?)\s*$`)
	approveCancelRegex = regexp.MustCompile(`(?m)^/approve\s*(cancel)\s*$`)
)

func init() {
	plugins.RegisterGitCommentPlugin("approve", func(cs plugins.ClientSets) plugins.GitCommentPlugin {
		plugin := &approvePlugin{
			issueClient: cs.GitIssueClient,
			prClient:    cs.GitPRClient,
			ownerClient: cs.OwnersClient,
		}
		return plugin
	})
}

type approvePlugin struct {
	issueClient plugins.GitIssueClient
	prClient    plugins.GitPRClient
	ownerClient plugins.OwnersClient

	allowAuthor bool
}

func (lp *approvePlugin) Name() string {
	return "approve"
}

func (lp *approvePlugin) Description() string {
	return "Adds or removes the 'approved' label which is typically used to gate merging."
}

func (lp *approvePlugin) Usage() string {
	return "/approve [cancel]"
}

func (lp *approvePlugin) BindFlags(flags *pflag.FlagSet) {
	flags.BoolVar(&lp.allowAuthor, "allow-author", false, "Whether allow author to add approve")
}

func (lp *approvePlugin) Do(ctx context.Context, e plugins.GitCommentEvent) error {
	if !e.IsPR || e.Action != plugins.GitCommentActionCreated {
		return nil
	}
	// Check body
	bodyClean := commentRegex.ReplaceAllString(e.Body, "")
	approveMatch := approveRegex.MatchString(bodyClean)
	approveCancelMatch := approveCancelRegex.MatchString(bodyClean)
	if !approveMatch && !approveCancelMatch {
		return nil
	}

	// Check author
	if !lp.allowAuthor {
		pr, err := lp.prClient.GetPR(ctx, e.Repo, e.Number)
		if err != nil {
			return err
		}
		if pr.User.Name == e.User.Name {
			resp := "you cannot APPROVE your own PR."
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
		resp := "adding APPROVE is restricted to approvers and reviewers in OWNERS files."
		return lp.issueClient.CreateIssueComment(ctx, e.Repo, plugins.GitIssue{Number: e.Number}, plugins.GitIssueComment{
			Body: plugins.FormatResponseRaw(e.Body, e.HTMLURL, e.User.Name, resp),
		})
	}
	if approveCancelMatch {
		return lp.issueClient.RemoveLabel(ctx, e.Repo, plugins.GitIssue{Number: e.Number}, plugins.Label{Name: "approved"})
	}
	return lp.issueClient.AddLabel(ctx, e.Repo, plugins.GitIssue{Number: e.Number}, []plugins.Label{{Name: "approved"}})
}
