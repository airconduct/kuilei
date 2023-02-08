package pluginhelpers

import (
	"context"
	"strings"

	"github.com/airconduct/go-probot"
	"github.com/airconduct/kuilei/pkg/plugins"
	"github.com/google/go-github/v48/github"
)

func GitIssueClientFromGithub(gh *probot.GitHubClient) plugins.GitIssueClient {
	return &githubClientWrapper{ghClient: gh}
}

func GitPRClientFromGithub(gh *probot.GitHubClient) plugins.GitPRClient {
	return &githubClientWrapper{ghClient: gh}
}

func GitRepoClientFromGithub(gh *probot.GitHubClient) plugins.GitRepoClient {
	return &githubClientWrapper{ghClient: gh}
}

type githubClientWrapper struct {
	ghClient *probot.GitHubClient
}

func (c *githubClientWrapper) CreateIssueComment(ctx context.Context, repo plugins.GitRepo, issue plugins.GitIssue, in plugins.GitIssueComment) error {
	_, _, err := c.ghClient.Issues.CreateComment(ctx, repo.Owner.Name, repo.Name, issue.Number, &github.IssueComment{
		Body: github.String(in.Body),
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

func (c *githubClientWrapper) GetPR(ctx context.Context, repo plugins.GitRepo, number int) (plugins.GitPullRequest, error) {
	pr, _, err := c.ghClient.PullRequests.Get(ctx, repo.Owner.Name, repo.Name, number)
	if err != nil {
		return plugins.GitPullRequest{}, err
	}
	return plugins.GitPullRequest{
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

func (c *githubClientWrapper) MergePR(ctx context.Context, repo plugins.GitRepo, number int, method string) error {
	_, _, err := c.ghClient.PullRequests.Merge(ctx, repo.Owner.Name, repo.Name, number, "", &github.PullRequestOptions{
		MergeMethod: method,
	})
	return err
}

func (c *githubClientWrapper) CreateStatus(
	ctx context.Context, repo plugins.GitRepo, ref string, status plugins.GitCommitStatus,
) error {
	_, _, err := c.ghClient.Repositories.CreateStatus(ctx, repo.Owner.Name, repo.Name, ref, &github.RepoStatus{
		State:       github.String(strings.ToLower(status.State)),
		Context:     github.String(status.Context),
		Description: github.String(status.Description),
		TargetURL:   github.String(status.TargetURL),
	})
	return err
}
