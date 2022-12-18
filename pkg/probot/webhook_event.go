package probot

import (
	"encoding/json"
)

type WebhookEvent interface {
	Type() string
	Action() string
}

func MakeWebhookEvent(e, a string) WebhookEventIdentifier {
	return func() (event string, action string) {
		return e, a
	}
}

type WebhookEventIdentifier func() (event, action string)

func (fn WebhookEventIdentifier) Type() string {
	tp, _ := fn()
	return tp
}

func (fn WebhookEventIdentifier) Action() string {
	_, act := fn()
	return act
}

type Handler interface {
	handlerIdentyfier()
}

type handleWith interface {
	Handle(Handler) error
}

type handleWithFunc func(Handler) error

func (fn handleWithFunc) Handle(h Handler) error {
	return fn(h)
}

type EventHandlerFunc[GT GitClientType, PT gitEventType] func(ctx ProbotContext[GT, PT])

func (EventHandlerFunc[GT, PT]) handlerIdentyfier() {}

type gitEventType interface {
	any
	// github.BranchProtectionRuleEvent |
	// 	github.CheckRunEvent |
	// 	github.CheckSuiteEvent |
	// 	github.CodeScanningAlertEvent |
	// 	github.CommitCommentEvent |
	// 	github.ContentReferenceEvent |
	// 	github.CreateEvent |
	// 	github.DeleteEvent |
	// 	github.DeployKeyEvent |
	// 	github.DeploymentEvent |
	// 	github.DeploymentStatusEvent |
	// 	github.DiscussionEvent |
	// 	github.ForkEvent |
	// 	github.GitHubAppAuthorizationEvent |
	// 	github.GollumEvent |
	// 	github.InstallationEvent |
	// 	github.InstallationRepositoriesEvent |
	// 	github.IssueCommentEvent |
	// 	github.IssueEvent |
	// 	github.LabelEvent |
	// 	github.MarketplacePurchaseEvent |
	// 	github.MemberEvent |
	// 	github.MembershipEvent |
	// 	github.MergeGroupEvent |
	// 	github.MetaEvent |
	// 	github.MilestoneEvent |
	// 	github.OrganizationEvent |
	// 	github.OrgBlockEvent |
	// 	github.PackageEvent |
	// 	github.PageBuildEvent |
	// 	github.PingEvent |
	// 	github.ProjectEvent |
	// 	github.ProjectCardEvent |
	// 	github.ProjectColumnEvent |
	// 	github.PublicEvent |
	// 	github.PullRequestEvent |
	// 	github.PullRequestReviewEvent |
	// 	github.PullRequestReviewCommentEvent |
	// 	github.PullRequestReviewThreadEvent |
	// 	github.PullRequestTargetEvent |
	// 	github.PushEvent |
	// 	github.RepositoryEvent |
	// 	github.RepositoryDispatchEvent |
	// 	github.RepositoryImportEvent |
	// 	github.RepositoryVulnerabilityAlertEvent |
	// 	github.ReleaseEvent |
	// 	github.SecretScanningAlertEvent |
	// 	github.StarEvent |
	// 	github.StatusEvent |
	// 	github.TeamEvent |
	// 	github.TeamAddEvent |
	// 	github.UserEvent |
	// 	github.WatchEvent |
	// 	github.WorkflowDispatchEvent |
	// 	github.WorkflowJobEvent |
	// 	github.WorkflowRunEvent |
}

func parseWebHook[T gitEventType](messageType string, payload []byte, out *T) (err error) {
	return json.Unmarshal(payload, out)
}
