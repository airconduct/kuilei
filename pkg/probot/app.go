package probot

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/bradleyfalzon/ghinstallation"
	"github.com/google/go-github/v48/github"
	"github.com/spf13/pflag"
)

func NewGithubAPP() App[GithubClient] {
	return &githubApp{
		handlers: make(map[string]Handler),
		clients:  make(map[int64]*github.Client),
	}
}

type githubApp struct {
	handlers map[string]Handler
	clients  map[int64]*github.Client

	appID          int64
	privateKeyFile string
	hmacTokenFile  string
	baseURL        string
	uploadURL      string
	serverOpts     ServerOptions

	hmacToken  []byte
	privateKey []byte
	loggerOptions
}

var _ App[GithubClient] = &githubApp{}

func (app *githubApp) GetServerOptions() ServerOptions {
	return app.serverOpts
}

func (app *githubApp) GetHTTPHandler() http.Handler {
	return http.HandlerFunc(app.handle)
}

func (app *githubApp) GetSecretToken() []byte {
	return app.hmacToken
}

func (app *githubApp) AddFlags(flags *pflag.FlagSet) {
	app.serverOpts.AddFlags(flags)
	app.loggerOptions.AddFlags(flags)

	flags.Int64Var(&app.appID, "github.appid", 0, "github App id")
	flags.StringVar(&app.privateKeyFile, "github.private-key-file", "", "github App private-key file")
	flags.StringVar(&app.hmacTokenFile, "github.hmac-token-file", "", "github App hmac token file")
	flags.StringVar(&app.baseURL, "github.base-url", "https://api.github.com", "github base URL")
	flags.StringVar(&app.uploadURL, "github.upload-url", "https://upload.github.com", "github base URL")
}

func (app *githubApp) On(events ...WebhookEvent) handleWith {
	return handleWithFunc(func(h Handler) error {
		for _, event := range events {
			key := event.Type()
			if action := event.Action(); action != "" {
				key = key + "." + action
			}
			if _, ok := app.handlers[key]; ok {
				return fmt.Errorf("event type %s already exists", key)
			}
			app.handlers[key] = h
		}
		return nil
	})
}

func (app *githubApp) Run(ctx context.Context) error {
	if err := app.initialize(); err != nil {
		return err
	}

	mux := http.NewServeMux()
	mux.HandleFunc(app.serverOpts.Path, app.handle)
	server := &http.Server{Addr: fmt.Sprintf("%s:%d", app.serverOpts.Address, app.serverOpts.Port), Handler: mux}
	server.RegisterOnShutdown(app.shutdown)
	return server.ListenAndServe()
}

func (app *githubApp) shutdown() {}

func (app *githubApp) initialize() error {
	rawToken, err := os.ReadFile(app.hmacTokenFile)
	if err != nil {
		return err
	}
	app.hmacToken = rawToken

	rawPrivateKey, err := os.ReadFile(app.privateKeyFile)
	if err != nil {
		return err
	}
	app.privateKey = rawPrivateKey
	return nil
}

func (app *githubApp) handle(w http.ResponseWriter, r *http.Request) {
	var handlerKey string
	defer func() {
		if e := recover(); e != nil {
			err, ok := e.(error)
			if ok {
				app.handleError(w, err, http.StatusBadRequest)
				app.logger.Error(
					err, "Failed handle event",
					"handler_key", handlerKey,
				)
			}
		}
	}()

	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	if r.Method != http.MethodPost {
		app.handleError(w, fmt.Errorf("method %s not found", r.Method), http.StatusNotFound)
		return
	}
	event := github.WebHookType(r)
	rawPayload, err := github.ValidatePayload(r, []byte(app.hmacToken))
	if err != nil {
		app.handleError(w, err, http.StatusBadRequest)
		return
	}

	handlerKey = event
	switch event {
	case "branch_protection_rule":
		payload := new(github.BranchProtectionRuleEvent)
		err = parseWebHook(event, rawPayload, payload)
		if err != nil {
			app.handleError(w, err, http.StatusBadRequest)
			return
		}
		cli, err := app.getClient(*payload.Installation.ID)
		if err != nil {
			app.handleError(w, err, http.StatusBadRequest)
			return
		}
		protoCtx := newProbotContext(ctx, app.logger, cli, payload)
		if action := protoCtx.Payload().Action; action != nil && *action != "" {
			handlerKey = handlerKey + "." + *action
		}
		handler := app.handlers[handlerKey].(EventHandlerFunc[GithubClient, github.BranchProtectionRuleEvent])
		handler(protoCtx)
	case "check_run":
		payload := new(github.CheckRunEvent)
		err = parseWebHook(event, rawPayload, payload)
		if err != nil {
			app.handleError(w, err, http.StatusBadRequest)
			return
		}
		cli, err := app.getClient(*payload.Installation.ID)
		if err != nil {
			app.handleError(w, err, http.StatusBadRequest)
			return
		}
		protoCtx := newProbotContext(ctx, app.logger, cli, payload)
		if action := protoCtx.Payload().Action; action != nil && *action != "" {
			handlerKey = handlerKey + "." + *action
		}
		handler := app.handlers[handlerKey].(EventHandlerFunc[GithubClient, github.CheckRunEvent])
		handler(protoCtx)
	case "check_suite":
		payload := new(github.CheckSuiteEvent)
		err = parseWebHook(event, rawPayload, payload)
	case "code_scanning_alert":
		payload := new(github.CodeScanningAlertEvent)
		err = parseWebHook(event, rawPayload, payload)
	case "commit_comment":
		payload := new(github.CommitCommentEvent)
		err = parseWebHook(event, rawPayload, payload)
	case "content_reference":
		payload := new(github.ContentReferenceEvent)
		err = parseWebHook(event, rawPayload, payload)
	case "create":
		payload := new(github.CreateEvent)
		err = parseWebHook(event, rawPayload, payload)
	case "delete":
		payload := new(github.DeleteEvent)
		err = parseWebHook(event, rawPayload, payload)
	case "deploy_key":
		payload := new(github.DeployKeyEvent)
		err = parseWebHook(event, rawPayload, payload)
	case "deployment":
		payload := new(github.DeploymentEvent)
		err = parseWebHook(event, rawPayload, payload)
	case "deployment_status":
		payload := new(github.DeploymentStatusEvent)
		err = parseWebHook(event, rawPayload, payload)
	case "discussion":
		payload := new(github.DiscussionEvent)
		err = parseWebHook(event, rawPayload, payload)
	case "fork":
		payload := new(github.ForkEvent)
		err = parseWebHook(event, rawPayload, payload)
	case "github_app_authorization":
		payload := new(github.GitHubAppAuthorizationEvent)
		err = parseWebHook(event, rawPayload, payload)
	case "gollum":
		payload := new(github.GollumEvent)
		err = parseWebHook(event, rawPayload, payload)
	case "installation":
		payload := new(github.InstallationEvent)
		err = parseWebHook(event, rawPayload, payload)
	case "installation_repositories":
		payload := new(github.InstallationRepositoriesEvent)
		err = parseWebHook(event, rawPayload, payload)
	case "issue_comment":
		payload := new(github.IssueCommentEvent)
		err = parseWebHook(event, rawPayload, payload)
		if err != nil {
			app.handleError(w, err, http.StatusBadRequest)
			return
		}
		cli, err := app.getClient(*payload.Installation.ID)
		if err != nil {
			app.handleError(w, err, http.StatusBadRequest)
			return
		}
		protoCtx := newProbotContext(ctx, app.logger, cli, payload)
		if action := protoCtx.Payload().Action; action != nil && *action != "" {
			handlerKey = handlerKey + "." + *action
		}
		handler := app.handlers[handlerKey].(EventHandlerFunc[GithubClient, github.IssueCommentEvent])
		handler(protoCtx)
	case "issues":
		payload := new(github.IssueEvent)
		err = parseWebHook(event, rawPayload, payload)
	case "label":
		payload := new(github.LabelEvent)
		err = parseWebHook(event, rawPayload, payload)
	case "marketplace_purchase":
		payload := new(github.MarketplacePurchaseEvent)
		err = parseWebHook(event, rawPayload, payload)
	case "member":
		payload := new(github.MemberEvent)
		err = parseWebHook(event, rawPayload, payload)
	case "membership":
		payload := new(github.MembershipEvent)
		err = parseWebHook(event, rawPayload, payload)
	case "merge_group":
		payload := new(github.MergeGroupEvent)
		err = parseWebHook(event, rawPayload, payload)
	case "meta":
		payload := new(github.MetaEvent)
		err = parseWebHook(event, rawPayload, payload)
	case "milestone":
		payload := new(github.MilestoneEvent)
		err = parseWebHook(event, rawPayload, payload)
	case "organization":
		payload := new(github.OrganizationEvent)
		err = parseWebHook(event, rawPayload, payload)
	case "org_block":
		payload := new(github.OrgBlockEvent)
		err = parseWebHook(event, rawPayload, payload)
	case "package":
		payload := new(github.PackageEvent)
		err = parseWebHook(event, rawPayload, payload)
	case "page_build":
		payload := new(github.PageBuildEvent)
		err = parseWebHook(event, rawPayload, payload)
	case "ping":
		payload := new(github.PingEvent)
		err = parseWebHook(event, rawPayload, payload)
	case "project":
		payload := new(github.ProjectEvent)
		err = parseWebHook(event, rawPayload, payload)
	case "project_card":
		payload := new(github.ProjectCardEvent)
		err = parseWebHook(event, rawPayload, payload)
	case "project_column":
		payload := new(github.ProjectColumnEvent)
		err = parseWebHook(event, rawPayload, payload)
	case "public":
		payload := new(github.PublicEvent)
		err = parseWebHook(event, rawPayload, payload)
	case "pull_request":
		payload := new(github.PullRequestEvent)
		err = parseWebHook(event, rawPayload, payload)
	case "pull_request_review":
		payload := new(github.PullRequestReviewEvent)
		err = parseWebHook(event, rawPayload, payload)
	case "pull_request_review_comment":
		payload := new(github.PullRequestReviewCommentEvent)
		err = parseWebHook(event, rawPayload, payload)
	case "pull_request_review_thread":
		payload := new(github.PullRequestReviewThreadEvent)
		err = parseWebHook(event, rawPayload, payload)
	case "pull_request_target":
		payload := new(github.PullRequestTargetEvent)
		err = parseWebHook(event, rawPayload, payload)
	case "push":
		payload := new(github.PushEvent)
		err = parseWebHook(event, rawPayload, payload)
	case "repository":
		payload := new(github.RepositoryEvent)
		err = parseWebHook(event, rawPayload, payload)
	case "repository_dispatch":
		payload := new(github.RepositoryDispatchEvent)
		err = parseWebHook(event, rawPayload, payload)
	case "repository_import":
		payload := new(github.RepositoryImportEvent)
		err = parseWebHook(event, rawPayload, payload)
	case "repository_vulnerability_alert":
		payload := new(github.RepositoryVulnerabilityAlertEvent)
		err = parseWebHook(event, rawPayload, payload)
	case "release":
		payload := new(github.ReleaseEvent)
		err = parseWebHook(event, rawPayload, payload)
	case "secret_scanning_alert":
		payload := new(github.SecretScanningAlertEvent)
		err = parseWebHook(event, rawPayload, payload)
	case "star":
		payload := new(github.StarEvent)
		err = parseWebHook(event, rawPayload, payload)
	case "status":
		payload := new(github.StatusEvent)
		err = parseWebHook(event, rawPayload, payload)
	case "team":
		payload := new(github.TeamEvent)
		err = parseWebHook(event, rawPayload, payload)
	case "team_add":
		payload := new(github.TeamAddEvent)
		err = parseWebHook(event, rawPayload, payload)
	case "user":
		payload := new(github.UserEvent)
		err = parseWebHook(event, rawPayload, payload)
	case "watch":
		payload := new(github.WatchEvent)
		err = parseWebHook(event, rawPayload, payload)
	case "workflow_dispatch":
		payload := new(github.WorkflowDispatchEvent)
		err = parseWebHook(event, rawPayload, payload)
	case "workflow_job":
		payload := new(github.WorkflowJobEvent)
		err = parseWebHook(event, rawPayload, payload)
	case "workflow_run":
		payload := new(github.WorkflowRunEvent)
		err = parseWebHook(event, rawPayload, payload)
	}

	if err != nil {
		app.handleError(w, err, http.StatusBadRequest)
		return
	}
}

func (app *githubApp) handleError(w http.ResponseWriter, err error, status int) {
	w.WriteHeader(status)
	fmt.Fprint(w, err.Error())
}

func (app *githubApp) getClient(installID int64) (*github.Client, error) {
	if cli, ok := app.clients[installID]; ok {
		return cli, nil
	}

	tr, err := ghinstallation.New(http.DefaultTransport, app.appID, installID, app.privateKey)
	if err != nil {
		return nil, err
	}
	cli, err := github.NewEnterpriseClient(app.baseURL, app.uploadURL, &http.Client{Transport: tr})
	if err != nil {
		return nil, err
	}
	app.clients[installID] = cli
	return cli, nil
}
