package plugins

import "context"

// GitComment plugin
type GitCommentPluginBuilder func(ClientSets, ...string) GitCommentPlugin

type GitCommentPlugin interface {
	Name() string
	Do(context.Context, GitCommentEvent) error
}

// GitPR plugin
type GitPRPluginBuilder func(ClientSets, ...string) GitPRPlugin

type GitPRPlugin interface {
	Name() string
	Do(context.Context, GitPREvent) error
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

var gitPRPlugins = map[string]GitPRPluginBuilder{}

func RegisterGitPRPlugin(name string, builder GitPRPluginBuilder) {
	gitPRPlugins[name] = builder
}

func GetGitPRPlugin(name string, clientSets ClientSets, arg ...string) GitPRPlugin {
	if builder, ok := gitPRPlugins[name]; ok {
		return builder(clientSets, arg...)
	}
	return nil
}
