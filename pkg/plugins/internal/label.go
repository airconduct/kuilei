package internal

import (
	"context"
	"regexp"
	"strings"

	"github.com/airconduct/kuilei/pkg/plugins"
	"github.com/spf13/pflag"
	"k8s.io/apimachinery/pkg/util/sets"
)

var (
	customLabelRegex       = regexp.MustCompile(`(?m)^/label\s*(.*?)\s*$`)
	customRemoveLabelRegex = regexp.MustCompile(`(?m)^/remove-label\s*(.*?)\s*$`)
)

func init() {
	plugins.RegisterGitCommentPlugin("label", func(cs plugins.ClientSets) plugins.GitCommentPlugin {
		plugin := &labelPlugin{
			issueClient: cs.GitIssueClient,
		}
		return plugin
	})
}

type labelPlugin struct {
	issueClient plugins.GitIssueClient

	forbiddenLabels []string
}

func (lp *labelPlugin) Name() string {
	return "label"
}

func (lp *labelPlugin) Description() string {
	return "Applies or removes a label from one of the recognized types of labels."
}

func (lp *labelPlugin) Usage() string {
	return "/[remove-]label [name]"
}

func (lp *labelPlugin) BindFlags(flags *pflag.FlagSet) {
	flags.StringSliceVar(&lp.forbiddenLabels, "forbidden", []string{}, "A group of labels that is forbidden to add")
}

func (lp *labelPlugin) Do(ctx context.Context, e plugins.GitCommentEvent) error {
	if e.Action != plugins.GitCommentActionCreated {
		return nil
	}
	forbiddenLabelSets := sets.NewString(lp.forbiddenLabels...)

	bodyClean := commentRegex.ReplaceAllString(e.Body, "")
	customLabelMatches := customLabelRegex.FindAllStringSubmatch(bodyClean, -1)
	customRemoveLabelMatches := customRemoveLabelRegex.FindAllStringSubmatch(bodyClean, -1)
	if len(customLabelMatches) == 0 && len(customRemoveLabelMatches) == 0 {
		return nil
	}
	var labelsToAdd []plugins.Label
	for _, match := range customLabelMatches {
		parts := strings.Split(strings.TrimSpace(match[0]), " ")
		if (parts[0] != "/label") || len(parts) != 2 {
			continue
		}
		if labelName := strings.ToLower(parts[1]); !forbiddenLabelSets.Has(labelName) {
			labelsToAdd = append(labelsToAdd, plugins.Label{Name: labelName})
		}
	}
	for _, match := range customRemoveLabelMatches {
		parts := strings.Split(strings.TrimSpace(match[0]), " ")
		if (parts[0] != "/remove-label") || len(parts) != 2 {
			continue
		}
		if labelName := strings.ToLower(parts[1]); !forbiddenLabelSets.Has(labelName) {
			lp.issueClient.RemoveLabel(ctx, e.Repo, plugins.GitIssue{Number: e.Number}, plugins.Label{Name: labelName})
		}
	}
	if len(labelsToAdd) != 0 {
		if err := lp.issueClient.AddLabel(ctx, e.Repo, plugins.GitIssue{Number: e.Number}, labelsToAdd); err != nil {
			return err
		}
	}
	return nil
}
