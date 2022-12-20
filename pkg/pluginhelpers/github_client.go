package pluginhelpers

import (
	"context"

	"github.com/airconduct/kuilei/pkg/plugins"
	"github.com/airconduct/kuilei/pkg/probot"
	"github.com/google/go-github/v48/github"
)

func GitIssueClientFromGithub(gh *probot.GithubClient) plugins.GitIssueClient {
	return &githubIssueClient{ghClient: gh}
}

type githubIssueClient struct {
	ghClient *probot.GithubClient
}

func (c *githubIssueClient) CreateIssueComment(ctx context.Context, repo plugins.GitRepo, issue plugins.GitIssue, in plugins.GitIssueComment) error {
	_, _, err := c.ghClient.Issues.CreateComment(ctx, repo.Owner.Name, repo.Name, issue.Number, &github.IssueComment{
		Body: &in.Body,
	})
	return err
}

func (c *githubIssueClient) AddLabel(ctx context.Context, repo plugins.GitRepo, issue plugins.GitIssue, labels []plugins.Label) error {
	var labelNames []string
	for _, l := range labels {
		labelNames = append(labelNames, l.Name)
	}
	_, _, err := c.ghClient.Issues.AddLabelsToIssue(ctx, repo.Owner.Name, repo.Name, issue.Number, labelNames)
	return err
}
