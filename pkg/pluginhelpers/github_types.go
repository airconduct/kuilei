package pluginhelpers

import (
	"github.com/google/go-github/v48/github"

	"github.com/airconduct/kuilei/pkg/plugins"
)

func GitLabelsFromGithub(labels []*github.Label) (output []plugins.Label) {
	for _, l := range labels {
		output = append(output, GitLabelFromGithub(l))
	}
	return
}

func GitLabelFromGithub(l *github.Label) plugins.Label {
	return plugins.Label{
		ID:    l.GetID(),
		Name:  l.GetName(),
		Color: l.GetColor(),
	}
}

func GitUsersFromGithub(users []*github.User) (output []plugins.GitUser) {
	for _, u := range users {
		output = append(output, GitUserFromGithub(u))
	}
	return
}

func GitUserFromGithub(u *github.User) plugins.GitUser {
	return plugins.GitUser{
		Name: u.GetLogin(),
	}
}
