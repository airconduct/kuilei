package plugins

import (
	"bytes"
	"context"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"text/template"

	"github.com/spf13/pflag"
	"k8s.io/apimachinery/pkg/util/sets"
)

const (
	helpPluginName = "help"
)

const helpCommentTemplateText = `Command: **{{.Name}}**
> {{.Description}}
>
> Usage:
>>{{.Usage}}
>
{{if eq .Name "help"}}>Available Commands:
>> | Name | Description | Usage |
>> | ---- | ---- | ---- |
{{.AvailableCommands}}{{end}}
`

var (
	helpRegex = regexp.MustCompile(`(?m)^/help\s*(.*?)\s*$`)

	helpCommentTemplate = template.Must(template.New("help").Parse(helpCommentTemplateText))
)

func init() {
	RegisterGitCommentPlugin(helpPluginName, func(cs ClientSets) GitCommentPlugin {
		return &helpPlugin{clientsets: cs}
	})
}

type helpPlugin struct {
	clientsets ClientSets

	AvailableCommands string
}

func (p *helpPlugin) Name() string {
	return helpPluginName
}

func (p *helpPlugin) Description() string {
	return "Help provides help for any command."
}

func (p *helpPlugin) Usage() string {
	return "/help [command]"
}

func (p *helpPlugin) BindFlags(flags *pflag.FlagSet) {}

func (p *helpPlugin) Do(ctx context.Context, e GitCommentEvent) error {
	if e.Action != GitCommentActionCreated {
		return nil
	}
	bodyClean := CleanMarkdownComments(e.Body)
	helpMatches := helpRegex.FindAllStringSubmatch(bodyClean, -1)
	if len(helpMatches) == 0 {
		return nil
	}
	commandSets := sets.NewString()
	for _, match := range helpMatches {
		parts := strings.Split(strings.TrimSpace(match[0]), " ")
		if parts[0] != "/help" {
			continue
		}
		if len(parts) != 2 {
			commandSets.Insert(helpPluginName)
			continue
		}
		commandSets.Insert(parts[1])
	}
	body, err := p.buildCommentBody(commandSets.List())
	if err != nil {
		return err
	}
	return p.clientsets.GitIssueClient.CreateIssueComment(ctx, e.Repo, GitIssue{Number: e.Number}, GitIssueComment{
		Body: body,
	})
}

func (p *helpPlugin) buildCommentBody(commands []string) (string, error) {
	// List all plugins
	var availableCommands []string
	allPlugins := map[string]Plugin{}
	for name, builder := range gitCommentPlugins {
		plugin := builder(p.clientsets)
		allPlugins[name] = plugin
		availableCommands = append(availableCommands, fmt.Sprintf(
			">> | %s | %s | %s |", plugin.Name(), plugin.Description(), plugin.Usage(),
		))
	}
	// Build AvailableCommands description
	sort.Strings(availableCommands)
	p.AvailableCommands = strings.Join(availableCommands, "\r\n")

	// Build body
	buf := new(bytes.Buffer)
	for _, name := range commands {
		plugin, ok := allPlugins[name]
		if !ok {
			fmt.Fprintf(buf, "Unknown command %s \r\n. Use \"/help\" for all commands usage.", name)
			continue
		}
		if err := helpCommentTemplate.Execute(buf, map[string]string{
			"Name":              plugin.Name(),
			"Description":       plugin.Description(),
			"Usage":             plugin.Usage(),
			"AvailableCommands": p.AvailableCommands,
		}); err != nil {
			return "", err
		}
	}
	return buf.String(), nil
}
