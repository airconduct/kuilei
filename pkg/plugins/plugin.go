package plugins

import (
	"context"

	"github.com/spf13/pflag"
)

type Plugin interface {
	Name() string
	Description() string
	Usage() string
	BindFlags(flags *pflag.FlagSet)
}

// GitComment plugin
type GitCommentPluginBuilder func(ClientSets) GitCommentPlugin

type GitCommentPlugin interface {
	Plugin
	Do(context.Context, GitCommentEvent) error
}

// GitPR plugin
type GitPRPluginBuilder func(ClientSets) GitPRPlugin

type GitPRPlugin interface {
	Plugin
	Do(context.Context, GitPREvent) error
}

var gitCommentPlugins = map[string]GitCommentPluginBuilder{}

func RegisterGitCommentPlugin(name string, builder GitCommentPluginBuilder) {
	gitCommentPlugins[name] = builder
}

func GetGitCommentPlugin(name string, clientSets ClientSets, args ...string) GitCommentPlugin {
	if builder, ok := gitCommentPlugins[name]; ok {
		p := builder(clientSets)
		flags := pflag.NewFlagSet(p.Name(), pflag.ContinueOnError)
		p.BindFlags(flags)
		flags.Parse(args)
		return p
	}
	return nil
}

var gitPRPlugins = map[string]GitPRPluginBuilder{}

func RegisterGitPRPlugin(name string, builder GitPRPluginBuilder) {
	gitPRPlugins[name] = builder
}

func GetGitPRPlugin(name string, clientSets ClientSets, args ...string) GitPRPlugin {
	if builder, ok := gitPRPlugins[name]; ok {
		p := builder(clientSets)
		flags := pflag.NewFlagSet(p.Name(), pflag.ContinueOnError)
		p.BindFlags(flags)
		flags.Parse(args)
		return p
	}
	return nil
}
