package pluginhelpers

import (
	"context"

	"github.com/airconduct/kuilei/pkg/plugins"
	"github.com/airconduct/kuilei/pkg/probot"
	"github.com/google/go-github/v48/github"
)

func GitIssueClientFromGithub(gh *probot.GithubClient) plugins.GitIssueClient {
	return &githubClientWrapper{ghClient: gh}
}

func GitPRClientFromGithub(gh *probot.GithubClient) plugins.GitPRClient {
	return &githubClientWrapper{ghClient: gh}
}

type githubClientWrapper struct {
	ghClient *probot.GithubClient
}

func (c *githubClientWrapper) CreateIssueComment(ctx context.Context, repo plugins.GitRepo, issue plugins.GitIssue, in plugins.GitIssueComment) error {
	_, _, err := c.ghClient.Issues.CreateComment(ctx, repo.Owner.Name, repo.Name, issue.Number, &github.IssueComment{
		Body: &in.Body,
	})
	return err
}

func (c *githubClientWrapper) AddLabel(ctx context.Context, repo plugins.GitRepo, issue plugins.GitIssue, labels []plugins.Label) error {
	var labelNames []string
	for _, l := range labels {
		labelNames = append(labelNames, l.Name)
	}
	_, _, err := c.ghClient.Issues.AddLabelsToIssue(ctx, repo.Owner.Name, repo.Name, issue.Number, labelNames)
	return err
}

func (c *githubClientWrapper) RemoveLabel(ctx context.Context, repo plugins.GitRepo, issue plugins.GitIssue, label plugins.Label) error {
	_, err := c.ghClient.Issues.RemoveLabelForIssue(ctx, repo.Owner.Name, repo.Name, issue.Number, label.Name)
	return err
}

func (c *githubClientWrapper) ListFiles(ctx context.Context, repo plugins.GitRepo, pr plugins.GitPullRequest) ([]plugins.GitCommitFile, error) {
	files, _, err := c.ghClient.PullRequests.ListFiles(ctx, repo.Owner.Name, repo.Name, pr.Number, &github.ListOptions{})
	if err != nil {
		return nil, err
	}
	var commitFiles []plugins.GitCommitFile
	for _, f := range files {
		commitFiles = append(commitFiles, plugins.GitCommitFile{Path: f.GetFilename()})
	}
	return commitFiles, nil
}

func (c *githubClientWrapper) GetPR(ctx context.Context, repo plugins.GitRepo, number int) (*plugins.GitPullRequest, error) {
	pr, _, err := c.ghClient.PullRequests.Get(ctx, repo.Owner.Name, repo.Name, number)
	if err != nil {
		return nil, err
	}
	return &plugins.GitPullRequest{
		ID:        int(pr.GetID()),
		Number:    pr.GetNumber(),
		State:     pr.GetState(),
		Locked:    pr.GetLocked(),
		Title:     pr.GetTitle(),
		Body:      pr.GetBody(),
		Labels:    GitLabelsFromGithub(pr.Labels),
		Assignees: GitUsersFromGithub(pr.Assignees),
		User:      GitUserFromGithub(pr.User),
	}, nil
}
