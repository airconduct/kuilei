package plugins

import "context"

// GitComment plugin
type GitCommentPluginBuilder func(ClientSets, ...string) GitCommentPlugin

type GitCommentPlugin interface {
	Name() string
	Do(context.Context, GitCommentEvent) error
}

// GitReviewComment plugin
type GitReviewCommentPluginBuilder func(ClientSets) GitReviewCommentPlugin

type GitReviewCommentPlugin interface {
	Name() string
	Do(context.Context, GitCommentEvent) error
}

var gitCommentPlugins = map[string]GitCommentPluginBuilder{}

func RegisterGitCommentPlugin(name string, builder GitCommentPluginBuilder) {
	gitCommentPlugins[name] = builder
}

func GetGitCommentPlugin(name string, clientSets ClientSets, arg ...string) GitCommentPlugin {
	if builder, ok := gitCommentPlugins[name]; ok {
		return builder(clientSets, arg...)
	}
	return nil
}
