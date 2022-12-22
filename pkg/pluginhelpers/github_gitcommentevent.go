package pluginhelpers

import (
	"github.com/airconduct/kuilei/pkg/plugins"
	"github.com/google/go-github/v48/github"
)

func GitCommentEventFromGithubIssueCommentEvent(event *github.IssueCommentEvent) plugins.GitCommentEvent {
	var assignees []plugins.GitUser
	for _, assignee := range event.Issue.Assignees {
		assignees = append(assignees, plugins.GitUser{Name: assignee.GetLogin()})
	}
	return plugins.GitCommentEvent{
		GitComment: plugins.GitComment{
			ID:           int(event.Issue.GetID()),
			NodeID:       event.Issue.GetNodeID(),
			CommentID:    int(event.Comment.GetID()),
			IsPR:         event.Issue.PullRequestLinks != nil,
			Body:         event.Comment.GetBody(),
			HTMLURL:      event.Comment.GetHTMLURL(),
			Number:       event.Issue.GetNumber(),
			User:         plugins.GitUser{Name: event.Comment.User.GetLogin()},
			IssueAuthor:  plugins.GitUser{Name: event.Issue.User.GetLogin()},
			Assignees:    assignees,
			IssueState:   event.Issue.GetState(),
			IssueTitle:   event.Issue.GetTitle(),
			IssueBody:    event.Issue.GetBody(),
			IssueHTMLURL: event.Issue.GetHTMLURL(),
		},
		Action: plugins.GitCommentEventAction(event.GetAction()),
		Repo: plugins.GitRepo{
			Name:  event.Repo.GetName(),
			Owner: plugins.GitUser{Name: event.Repo.Owner.GetLogin()},
		},
	}
}

func GitCommentEventFromGithubIssuesEvent(event *github.IssuesEvent) plugins.GitCommentEvent {
	var assignees []plugins.GitUser
	for _, assignee := range event.Issue.Assignees {
		assignees = append(assignees, plugins.GitUser{Name: assignee.GetLogin()})
	}
	return plugins.GitCommentEvent{
		GitComment: plugins.GitComment{
			ID:           int(event.Issue.GetID()),
			NodeID:       event.Issue.GetNodeID(),
			IsPR:         event.Issue.PullRequestLinks != nil,
			Body:         event.Issue.GetBody(),
			Number:       event.Issue.GetNumber(),
			User:         plugins.GitUser{Name: event.Issue.User.GetLogin()},
			IssueAuthor:  plugins.GitUser{Name: event.Issue.User.GetLogin()},
			Assignees:    assignees,
			IssueState:   event.Issue.GetState(),
			IssueTitle:   event.Issue.GetTitle(),
			IssueBody:    event.Issue.GetBody(),
			IssueHTMLURL: event.Issue.GetHTMLURL(),
		},
		Action: plugins.GitCommentEventAction(event.GetAction()),
		Repo: plugins.GitRepo{
			Name:  event.Repo.GetName(),
			Owner: plugins.GitUser{Name: event.Repo.Owner.GetLogin()},
		},
	}
}

func GitCommentEventFromGithubPullRequestEvent(event *github.PullRequestEvent) plugins.GitCommentEvent {
	var assignees []plugins.GitUser
	for _, assignee := range event.PullRequest.Assignees {
		assignees = append(assignees, plugins.GitUser{Name: assignee.GetLogin()})
	}
	return plugins.GitCommentEvent{
		GitComment: plugins.GitComment{
			ID:           int(event.PullRequest.GetID()),
			NodeID:       event.PullRequest.GetNodeID(),
			IsPR:         true,
			Body:         event.PullRequest.GetBody(),
			Number:       event.PullRequest.GetNumber(),
			User:         plugins.GitUser{Name: event.PullRequest.User.GetLogin()},
			IssueAuthor:  plugins.GitUser{Name: event.PullRequest.User.GetLogin()},
			Assignees:    assignees,
			IssueState:   event.PullRequest.GetState(),
			IssueTitle:   event.PullRequest.GetTitle(),
			IssueBody:    event.PullRequest.GetBody(),
			IssueHTMLURL: event.PullRequest.GetHTMLURL(),
		},
		Action: plugins.GitCommentEventAction(event.GetAction()),
		Repo: plugins.GitRepo{
			Name:  event.Repo.GetName(),
			Owner: plugins.GitUser{Name: event.Repo.Owner.GetLogin()},
		},
	}
}

func GitCommentEventFromGithubPullRequestReviewEvent(event *github.PullRequestReviewEvent) plugins.GitCommentEvent {
	var assignees []plugins.GitUser
	for _, assignee := range event.PullRequest.Assignees {
		assignees = append(assignees, plugins.GitUser{Name: assignee.GetLogin()})
	}
	return plugins.GitCommentEvent{
		GitComment: plugins.GitComment{
			ID:           int(event.PullRequest.GetID()),
			NodeID:       event.PullRequest.GetNodeID(),
			CommentID:    int(event.Review.GetID()),
			IsPR:         true,
			Body:         event.Review.GetBody(),
			Number:       event.PullRequest.GetNumber(),
			User:         plugins.GitUser{Name: event.Review.User.GetLogin()},
			IssueAuthor:  plugins.GitUser{Name: event.PullRequest.User.GetLogin()},
			Assignees:    assignees,
			IssueState:   event.PullRequest.GetState(),
			IssueTitle:   event.PullRequest.GetTitle(),
			IssueBody:    event.PullRequest.GetBody(),
			IssueHTMLURL: event.PullRequest.GetHTMLURL(),
		},
		Action: plugins.GitCommentEventAction(event.GetAction()),
		Repo: plugins.GitRepo{
			Name:  event.Repo.GetName(),
			Owner: plugins.GitUser{Name: event.Repo.Owner.GetLogin()},
		},
	}
}

func GitCommentEventFromGithubPullRequestReviewCommentEvent(event *github.PullRequestReviewCommentEvent) plugins.GitCommentEvent {
	var assignees []plugins.GitUser
	for _, assignee := range event.PullRequest.Assignees {
		assignees = append(assignees, plugins.GitUser{Name: assignee.GetLogin()})
	}
	return plugins.GitCommentEvent{
		GitComment: plugins.GitComment{
			ID:           int(event.PullRequest.GetID()),
			NodeID:       event.PullRequest.GetNodeID(),
			CommentID:    int(event.Comment.GetID()),
			IsPR:         true,
			Body:         event.Comment.GetBody(),
			Number:       event.PullRequest.GetNumber(),
			User:         plugins.GitUser{Name: event.Comment.User.GetLogin()},
			IssueAuthor:  plugins.GitUser{Name: event.PullRequest.User.GetLogin()},
			Assignees:    assignees,
			IssueState:   event.PullRequest.GetState(),
			IssueTitle:   event.PullRequest.GetTitle(),
			IssueBody:    event.PullRequest.GetBody(),
			IssueHTMLURL: event.PullRequest.GetHTMLURL(),
		},
		Action: plugins.GitCommentEventAction(event.GetAction()),
		Repo: plugins.GitRepo{
			Name:  event.Repo.GetName(),
			Owner: plugins.GitUser{Name: event.Repo.Owner.GetLogin()},
		},
	}
}
