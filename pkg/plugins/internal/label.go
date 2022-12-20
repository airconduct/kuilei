package internal

import (
	"context"
	"regexp"
	"strings"

	"github.com/airconduct/kuilei/pkg/plugins"
)

var (
	commentRegex           = regexp.MustCompile(`(?s)<!--(.*?)-->`)
	customLabelRegex       = regexp.MustCompile(`(?m)^/label\s*(.*?)\s*$`)
	customRemoveLabelRegex = regexp.MustCompile(`(?m)^/remove-label\s*(.*?)\s*$`)
)

func init() {
	plugins.RegisterGitCommentPlugin("label", func(cs plugins.ClientSets, args ...string) plugins.GitCommentPlugin {
		return &labelPlugin{issueClient: cs.GitIssueClient, configClient: cs.PluginConfigClient}
	})
}

type labelPlugin struct {
	issueClient  plugins.GitIssueClient
	configClient plugins.PluginConfigClient
}

func (lp *labelPlugin) Name() string {
	return "label"
}

func (lp *labelPlugin) Do(ctx context.Context, e plugins.GitCommentEvent) error {
	if e.Action != plugins.GitCommentActionCreated {
		return nil
	}
	bodyClean := commentRegex.ReplaceAllString(e.Body, "")
	customLabelMatches := customLabelRegex.FindAllStringSubmatch(bodyClean, -1)
	customRemoveLabelMatches := customRemoveLabelRegex.FindAllStringSubmatch(bodyClean, -1)
	if len(customLabelMatches) == 0 && len(customRemoveLabelMatches) == 0 {
		return nil
	}
	var labelsToAdd []plugins.Label
	for _, match := range customLabelMatches {
		parts := strings.Split(strings.TrimSpace(match[0]), " ")
		if ((parts[0] != "/label") && (parts[0] != "/remove-label")) || len(parts) != 2 {
			continue
		}
		labelsToAdd = append(labelsToAdd, plugins.Label{Name: strings.ToLower(parts[1])})
	}
	if len(labelsToAdd) != 0 {
		if err := lp.issueClient.AddLabel(ctx, e.Repo, plugins.GitIssue{Number: e.Number}, labelsToAdd); err != nil {
			return err
		}
	}
	return nil
}
