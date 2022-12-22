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

func (app *githubApp) On(events ...WebhookEvent) handlerLoader {
	return handlerLoadFunc(func(h Handler) error {
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
	if app.serverOpts.Path == "" {
		app.serverOpts.Path = "/"
	}
	mux.HandleFunc(app.serverOpts.Path, app.handle)
	server := &http.Server{Addr: fmt.Sprintf("%s:%d", app.serverOpts.Address, app.serverOpts.Port), Handler: mux}
	server.RegisterOnShutdown(app.shutdown)
	app.logger.Info("Kuilei hook is serving", "addr", server.Addr)
	return server.ListenAndServe()
}

func (app *githubApp) shutdown() {}

func (app *githubApp) initialize() error {
	rawToken, err := os.ReadFile(app.hmacTokenFile)
	if err != nil {
		return fmt.Errorf("failed to read hmac token file, %w", err)
	}
	app.hmacToken = rawToken

	rawPrivateKey, err := os.ReadFile(app.privateKeyFile)
	if err != nil {
		return fmt.Errorf("failed to read private key file, %w", err)
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
	app.logger.Info("Handle event", "event", event)

	switch event {
	case "branch_protection_rule":
		if err := genericHandleFunc(
			ctx, app.logger, event, rawPayload,
			func(payload *github.BranchProtectionRuleEvent) (*github.Client, error) {
				handlerKey = getHandlerKey(event, payload)
				return app.getClient(payload.GetInstallation().GetID())
			},
			app.handlers,
		); err != nil {
			app.handleError(w, err, http.StatusBadRequest)
			return
		}
	case "check_run":
		if err := genericHandleFunc(
			ctx, app.logger, event, rawPayload,
			func(payload *github.CheckRunEvent) (*github.Client, error) {
				handlerKey = getHandlerKey(event, payload)
				return app.getClient(payload.GetInstallation().GetID())
			},
			app.handlers,
		); err != nil {
			app.handleError(w, err, http.StatusBadRequest)
			return
		}
	case "check_suite":
		if err := genericHandleFunc(
			ctx, app.logger, event, rawPayload,
			func(payload *github.CheckSuiteEvent) (*github.Client, error) {
				handlerKey = getHandlerKey(event, payload)
				return app.getClient(payload.GetInstallation().GetID())
			},
			app.handlers,
		); err != nil {
			app.handleError(w, err, http.StatusBadRequest)
			return
		}
	case "code_scanning_alert":
		// TODO: CodeScanningAlertEvent do not contain Installation
		//
		// 	payload := new(github.CodeScanningAlertEvent)
		//	payload.Installation (wrong)
	case "commit_comment":
		if err := genericHandleFunc(
			ctx, app.logger, event, rawPayload,
			func(payload *github.CommitCommentEvent) (*github.Client, error) {
				handlerKey = getHandlerKey(event, payload)
				return app.getClient(payload.GetInstallation().GetID())
			},
			app.handlers,
		); err != nil {
			app.handleError(w, err, http.StatusBadRequest)
			return
		}
	case "content_reference":
		if err := genericHandleFunc(
			ctx, app.logger, event, rawPayload,
			func(payload *github.ContentReferenceEvent) (*github.Client, error) {
				handlerKey = getHandlerKey(event, payload)
				return app.getClient(payload.GetInstallation().GetID())
			},
			app.handlers,
		); err != nil {
			app.handleError(w, err, http.StatusBadRequest)
			return
		}
	case "create":
		if err := genericHandleFunc(
			ctx, app.logger, event, rawPayload,
			func(payload *github.CreateEvent) (*github.Client, error) {
				handlerKey = getHandlerKey(event, payload)
				return app.getClient(payload.GetInstallation().GetID())
			},
			app.handlers,
		); err != nil {
			app.handleError(w, err, http.StatusBadRequest)
			return
		}
	case "delete":
		if err := genericHandleFunc(
			ctx, app.logger, event, rawPayload,
			func(payload *github.DeleteEvent) (*github.Client, error) {
				handlerKey = getHandlerKey(event, payload)
				return app.getClient(payload.GetInstallation().GetID())
			},
			app.handlers,
		); err != nil {
			app.handleError(w, err, http.StatusBadRequest)
			return
		}
	case "deploy_key":
		if err := genericHandleFunc(
			ctx, app.logger, event, rawPayload,
			func(payload *github.DeployKeyEvent) (*github.Client, error) {
				handlerKey = getHandlerKey(event, payload)
				return app.getClient(payload.GetInstallation().GetID())
			},
			app.handlers,
		); err != nil {
			app.handleError(w, err, http.StatusBadRequest)
			return
		}
	case "deployment":
		if err := genericHandleFunc(
			ctx, app.logger, event, rawPayload,
			func(payload *github.DeploymentEvent) (*github.Client, error) {
				handlerKey = getHandlerKey(event, payload)
				return app.getClient(payload.GetInstallation().GetID())
			},
			app.handlers,
		); err != nil {
			app.handleError(w, err, http.StatusBadRequest)
			return
		}
	case "deployment_status":
		if err := genericHandleFunc(
			ctx, app.logger, event, rawPayload,
			func(payload *github.DeploymentStatusEvent) (*github.Client, error) {
				handlerKey = getHandlerKey(event, payload)
				return app.getClient(payload.GetInstallation().GetID())
			},
			app.handlers,
		); err != nil {
			app.handleError(w, err, http.StatusBadRequest)
			return
		}
	case "discussion":
		if err := genericHandleFunc(
			ctx, app.logger, event, rawPayload,
			func(payload *github.DiscussionEvent) (*github.Client, error) {
				handlerKey = getHandlerKey(event, payload)
				return app.getClient(payload.GetInstallation().GetID())
			},
			app.handlers,
		); err != nil {
			app.handleError(w, err, http.StatusBadRequest)
			return
		}
	case "fork":
		if err := genericHandleFunc(
			ctx, app.logger, event, rawPayload,
			func(payload *github.ForkEvent) (*github.Client, error) {
				handlerKey = getHandlerKey(event, payload)
				return app.getClient(payload.GetInstallation().GetID())
			},
			app.handlers,
		); err != nil {
			app.handleError(w, err, http.StatusBadRequest)
			return
		}
	case "github_app_authorization":
		if err := genericHandleFunc(
			ctx, app.logger, event, rawPayload,
			func(payload *github.GitHubAppAuthorizationEvent) (*github.Client, error) {
				handlerKey = getHandlerKey(event, payload)
				return app.getClient(payload.GetInstallation().GetID())
			},
			app.handlers,
		); err != nil {
			app.handleError(w, err, http.StatusBadRequest)
			return
		}
	case "gollum":
		if err := genericHandleFunc(
			ctx, app.logger, event, rawPayload,
			func(payload *github.GollumEvent) (*github.Client, error) {
				handlerKey = getHandlerKey(event, payload)
				return app.getClient(payload.GetInstallation().GetID())
			},
			app.handlers,
		); err != nil {
			app.handleError(w, err, http.StatusBadRequest)
			return
		}
	case "installation":
		if err := genericHandleFunc(
			ctx, app.logger, event, rawPayload,
			func(payload *github.InstallationEvent) (*github.Client, error) {
				handlerKey = getHandlerKey(event, payload)
				return app.getClient(payload.GetInstallation().GetID())
			},
			app.handlers,
		); err != nil {
			app.handleError(w, err, http.StatusBadRequest)
			return
		}
	case "installation_repositories":
		if err := genericHandleFunc(
			ctx, app.logger, event, rawPayload,
			func(payload *github.InstallationRepositoriesEvent) (*github.Client, error) {
				handlerKey = getHandlerKey(event, payload)
				return app.getClient(payload.GetInstallation().GetID())
			},
			app.handlers,
		); err != nil {
			app.handleError(w, err, http.StatusBadRequest)
			return
		}
	case "issue_comment":
		if err := genericHandleFunc(
			ctx, app.logger, event, rawPayload,
			func(payload *github.IssueCommentEvent) (*github.Client, error) {
				handlerKey = getHandlerKey(event, payload)
				return app.getClient(payload.GetInstallation().GetID())
			},
			app.handlers,
		); err != nil {
			app.handleError(w, err, http.StatusBadRequest)
			return
		}
	case "issues":
		if err := genericHandleFunc(
			ctx, app.logger, event, rawPayload,
			func(payload *github.IssuesEvent) (*github.Client, error) {
				handlerKey = getHandlerKey(event, payload)
				return app.getClient(payload.GetInstallation().GetID())
			},
			app.handlers,
		); err != nil {
			app.handleError(w, err, http.StatusBadRequest)
			return
		}
	case "label":
		if err := genericHandleFunc(
			ctx, app.logger, event, rawPayload,
			func(payload *github.LabelEvent) (*github.Client, error) {
				handlerKey = getHandlerKey(event, payload)
				return app.getClient(payload.GetInstallation().GetID())
			},
			app.handlers,
		); err != nil {
			app.handleError(w, err, http.StatusBadRequest)
			return
		}
	case "marketplace_purchase":
		if err := genericHandleFunc(
			ctx, app.logger, event, rawPayload,
			func(payload *github.MarketplacePurchaseEvent) (*github.Client, error) {
				handlerKey = getHandlerKey(event, payload)
				return app.getClient(payload.GetInstallation().GetID())
			},
			app.handlers,
		); err != nil {
			app.handleError(w, err, http.StatusBadRequest)
			return
		}
	case "member":
		if err := genericHandleFunc(
			ctx, app.logger, event, rawPayload,
			func(payload *github.MemberEvent) (*github.Client, error) {
				handlerKey = getHandlerKey(event, payload)
				return app.getClient(payload.GetInstallation().GetID())
			},
			app.handlers,
		); err != nil {
			app.handleError(w, err, http.StatusBadRequest)
			return
		}
	case "membership":
		if err := genericHandleFunc(
			ctx, app.logger, event, rawPayload,
			func(payload *github.MembershipEvent) (*github.Client, error) {
				handlerKey = getHandlerKey(event, payload)
				return app.getClient(payload.GetInstallation().GetID())
			},
			app.handlers,
		); err != nil {
			app.handleError(w, err, http.StatusBadRequest)
			return
		}
	case "merge_group":
		if err := genericHandleFunc(
			ctx, app.logger, event, rawPayload,
			func(payload *github.MergeGroupEvent) (*github.Client, error) {
				handlerKey = getHandlerKey(event, payload)
				return app.getClient(payload.GetInstallation().GetID())
			},
			app.handlers,
		); err != nil {
			app.handleError(w, err, http.StatusBadRequest)
			return
		}
	case "meta":
		if err := genericHandleFunc(
			ctx, app.logger, event, rawPayload,
			func(payload *github.MetaEvent) (*github.Client, error) {
				handlerKey = getHandlerKey(event, payload)
				return app.getClient(payload.GetInstallation().GetID())
			},
			app.handlers,
		); err != nil {
			app.handleError(w, err, http.StatusBadRequest)
			return
		}
	case "milestone":
		if err := genericHandleFunc(
			ctx, app.logger, event, rawPayload,
			func(payload *github.MilestoneEvent) (*github.Client, error) {
				handlerKey = getHandlerKey(event, payload)
				return app.getClient(payload.GetInstallation().GetID())
			},
			app.handlers,
		); err != nil {
			app.handleError(w, err, http.StatusBadRequest)
			return
		}
	case "organization":
		if err := genericHandleFunc(
			ctx, app.logger, event, rawPayload,
			func(payload *github.OrganizationEvent) (*github.Client, error) {
				handlerKey = getHandlerKey(event, payload)
				return app.getClient(payload.GetInstallation().GetID())
			},
			app.handlers,
		); err != nil {
			app.handleError(w, err, http.StatusBadRequest)
			return
		}
	case "org_block":
		if err := genericHandleFunc(
			ctx, app.logger, event, rawPayload,
			func(payload *github.OrgBlockEvent) (*github.Client, error) {
				handlerKey = getHandlerKey(event, payload)
				return app.getClient(payload.GetInstallation().GetID())
			},
			app.handlers,
		); err != nil {
			app.handleError(w, err, http.StatusBadRequest)
			return
		}
	case "package":
		if err := genericHandleFunc(
			ctx, app.logger, event, rawPayload,
			func(payload *github.PackageEvent) (*github.Client, error) {
				handlerKey = getHandlerKey(event, payload)
				return app.getClient(payload.GetInstallation().GetID())
			},
			app.handlers,
		); err != nil {
			app.handleError(w, err, http.StatusBadRequest)
			return
		}
	case "page_build":
		if err := genericHandleFunc(
			ctx, app.logger, event, rawPayload,
			func(payload *github.PageBuildEvent) (*github.Client, error) {
				handlerKey = getHandlerKey(event, payload)
				return app.getClient(payload.GetInstallation().GetID())
			},
			app.handlers,
		); err != nil {
			app.handleError(w, err, http.StatusBadRequest)
			return
		}
	case "ping":
		if err := genericHandleFunc(
			ctx, app.logger, event, rawPayload,
			func(payload *github.PingEvent) (*github.Client, error) {
				handlerKey = getHandlerKey(event, payload)
				return app.getClient(payload.GetInstallation().GetID())
			},
			app.handlers,
		); err != nil {
			app.handleError(w, err, http.StatusBadRequest)
			return
		}
	case "project":
		if err := genericHandleFunc(
			ctx, app.logger, event, rawPayload,
			func(payload *github.ProjectEvent) (*github.Client, error) {
				handlerKey = getHandlerKey(event, payload)
				return app.getClient(payload.GetInstallation().GetID())
			},
			app.handlers,
		); err != nil {
			app.handleError(w, err, http.StatusBadRequest)
			return
		}
	case "project_card":
		if err := genericHandleFunc(
			ctx, app.logger, event, rawPayload,
			func(payload *github.ProjectCardEvent) (*github.Client, error) {
				handlerKey = getHandlerKey(event, payload)
				return app.getClient(payload.GetInstallation().GetID())
			},
			app.handlers,
		); err != nil {
			app.handleError(w, err, http.StatusBadRequest)
			return
		}
	case "project_column":
		if err := genericHandleFunc(
			ctx, app.logger, event, rawPayload,
			func(payload *github.ProjectColumnEvent) (*github.Client, error) {
				handlerKey = getHandlerKey(event, payload)
				return app.getClient(payload.GetInstallation().GetID())
			},
			app.handlers,
		); err != nil {
			app.handleError(w, err, http.StatusBadRequest)
			return
		}
	case "public":
		if err := genericHandleFunc(
			ctx, app.logger, event, rawPayload,
			func(payload *github.PublicEvent) (*github.Client, error) {
				handlerKey = getHandlerKey(event, payload)
				return app.getClient(payload.GetInstallation().GetID())
			},
			app.handlers,
		); err != nil {
			app.handleError(w, err, http.StatusBadRequest)
			return
		}
	case "pull_request":
		if err := genericHandleFunc(
			ctx, app.logger, event, rawPayload,
			func(payload *github.PullRequestEvent) (*github.Client, error) {
				handlerKey = getHandlerKey(event, payload)
				return app.getClient(payload.GetInstallation().GetID())
			},
			app.handlers,
		); err != nil {
			app.handleError(w, err, http.StatusBadRequest)
			return
		}
	case "pull_request_review":
		if err := genericHandleFunc(
			ctx, app.logger, event, rawPayload,
			func(payload *github.PullRequestReviewEvent) (*github.Client, error) {
				handlerKey = getHandlerKey(event, payload)
				return app.getClient(payload.GetInstallation().GetID())
			},
			app.handlers,
		); err != nil {
			app.handleError(w, err, http.StatusBadRequest)
			return
		}
	case "pull_request_review_comment":
		if err := genericHandleFunc(
			ctx, app.logger, event, rawPayload,
			func(payload *github.PullRequestReviewCommentEvent) (*github.Client, error) {
				handlerKey = getHandlerKey(event, payload)
				return app.getClient(payload.GetInstallation().GetID())
			},
			app.handlers,
		); err != nil {
			app.handleError(w, err, http.StatusBadRequest)
			return
		}
	case "pull_request_review_thread":
		if err := genericHandleFunc(
			ctx, app.logger, event, rawPayload,
			func(payload *github.PullRequestReviewThreadEvent) (*github.Client, error) {
				handlerKey = getHandlerKey(event, payload)
				return app.getClient(payload.GetInstallation().GetID())
			},
			app.handlers,
		); err != nil {
			app.handleError(w, err, http.StatusBadRequest)
			return
		}
	case "pull_request_target":
		if err := genericHandleFunc(
			ctx, app.logger, event, rawPayload,
			func(payload *github.PullRequestTargetEvent) (*github.Client, error) {
				handlerKey = getHandlerKey(event, payload)
				return app.getClient(payload.GetInstallation().GetID())
			},
			app.handlers,
		); err != nil {
			app.handleError(w, err, http.StatusBadRequest)
			return
		}
	case "push":
		if err := genericHandleFunc(
			ctx, app.logger, event, rawPayload,
			func(payload *github.PushEvent) (*github.Client, error) {
				handlerKey = getHandlerKey(event, payload)
				return app.getClient(payload.GetInstallation().GetID())
			},
			app.handlers,
		); err != nil {
			app.handleError(w, err, http.StatusBadRequest)
			return
		}
	case "repository":
		if err := genericHandleFunc(
			ctx, app.logger, event, rawPayload,
			func(payload *github.RepositoryEvent) (*github.Client, error) {
				handlerKey = getHandlerKey(event, payload)
				return app.getClient(payload.GetInstallation().GetID())
			},
			app.handlers,
		); err != nil {
			app.handleError(w, err, http.StatusBadRequest)
			return
		}
	case "repository_dispatch":
		if err := genericHandleFunc(
			ctx, app.logger, event, rawPayload,
			func(payload *github.RepositoryDispatchEvent) (*github.Client, error) {
				handlerKey = getHandlerKey(event, payload)
				return app.getClient(payload.GetInstallation().GetID())
			},
			app.handlers,
		); err != nil {
			app.handleError(w, err, http.StatusBadRequest)
			return
		}
	case "repository_import":
		// TODO
	case "repository_vulnerability_alert":
		if err := genericHandleFunc(
			ctx, app.logger, event, rawPayload,
			func(payload *github.RepositoryVulnerabilityAlertEvent) (*github.Client, error) {
				handlerKey = getHandlerKey(event, payload)
				return app.getClient(payload.GetInstallation().GetID())
			},
			app.handlers,
		); err != nil {
			app.handleError(w, err, http.StatusBadRequest)
			return
		}
	case "release":
		if err := genericHandleFunc(
			ctx, app.logger, event, rawPayload,
			func(payload *github.ReleaseEvent) (*github.Client, error) {
				handlerKey = getHandlerKey(event, payload)
				return app.getClient(payload.GetInstallation().GetID())
			},
			app.handlers,
		); err != nil {
			app.handleError(w, err, http.StatusBadRequest)
			return
		}
	case "secret_scanning_alert":
		if err := genericHandleFunc(
			ctx, app.logger, event, rawPayload,
			func(payload *github.SecretScanningAlertEvent) (*github.Client, error) {
				handlerKey = getHandlerKey(event, payload)
				return app.getClient(payload.GetInstallation().GetID())
			},
			app.handlers,
		); err != nil {
			app.handleError(w, err, http.StatusBadRequest)
			return
		}
	case "star":
		if err := genericHandleFunc(
			ctx, app.logger, event, rawPayload,
			func(payload *github.StarEvent) (*github.Client, error) {
				handlerKey = getHandlerKey(event, payload)
				return app.getClient(payload.GetInstallation().GetID())
			},
			app.handlers,
		); err != nil {
			app.handleError(w, err, http.StatusBadRequest)
			return
		}
	case "status":
		if err := genericHandleFunc(
			ctx, app.logger, event, rawPayload,
			func(payload *github.StatusEvent) (*github.Client, error) {
				handlerKey = getHandlerKey(event, payload)
				return app.getClient(payload.GetInstallation().GetID())
			},
			app.handlers,
		); err != nil {
			app.handleError(w, err, http.StatusBadRequest)
			return
		}
	case "team":
		if err := genericHandleFunc(
			ctx, app.logger, event, rawPayload,
			func(payload *github.TeamEvent) (*github.Client, error) {
				handlerKey = getHandlerKey(event, payload)
				return app.getClient(payload.GetInstallation().GetID())
			},
			app.handlers,
		); err != nil {
			app.handleError(w, err, http.StatusBadRequest)
			return
		}
	case "team_add":
		if err := genericHandleFunc(
			ctx, app.logger, event, rawPayload,
			func(payload *github.TeamAddEvent) (*github.Client, error) {
				handlerKey = getHandlerKey(event, payload)
				return app.getClient(payload.GetInstallation().GetID())
			},
			app.handlers,
		); err != nil {
			app.handleError(w, err, http.StatusBadRequest)
			return
		}
	case "user":
		if err := genericHandleFunc(
			ctx, app.logger, event, rawPayload,
			func(payload *github.UserEvent) (*github.Client, error) {
				handlerKey = getHandlerKey(event, payload)
				return app.getClient(payload.GetInstallation().GetID())
			},
			app.handlers,
		); err != nil {
			app.handleError(w, err, http.StatusBadRequest)
			return
		}
	case "watch":
		if err := genericHandleFunc(
			ctx, app.logger, event, rawPayload,
			func(payload *github.WatchEvent) (*github.Client, error) {
				handlerKey = getHandlerKey(event, payload)
				return app.getClient(payload.GetInstallation().GetID())
			},
			app.handlers,
		); err != nil {
			app.handleError(w, err, http.StatusBadRequest)
			return
		}
	case "workflow_dispatch":
		if err := genericHandleFunc(
			ctx, app.logger, event, rawPayload,
			func(payload *github.WorkflowDispatchEvent) (*github.Client, error) {
				handlerKey = getHandlerKey(event, payload)
				return app.getClient(payload.GetInstallation().GetID())
			},
			app.handlers,
		); err != nil {
			app.handleError(w, err, http.StatusBadRequest)
			return
		}
	case "workflow_job":
		if err := genericHandleFunc(
			ctx, app.logger, event, rawPayload,
			func(payload *github.WorkflowJobEvent) (*github.Client, error) {
				handlerKey = getHandlerKey(event, payload)
				return app.getClient(payload.GetInstallation().GetID())
			},
			app.handlers,
		); err != nil {
			app.handleError(w, err, http.StatusBadRequest)
			return
		}
	case "workflow_run":
		if err := genericHandleFunc(
			ctx, app.logger, event, rawPayload,
			func(payload *github.WorkflowRunEvent) (*github.Client, error) {
				handlerKey = getHandlerKey(event, payload)
				return app.getClient(payload.GetInstallation().GetID())
			},
			app.handlers,
		); err != nil {
			app.handleError(w, err, http.StatusBadRequest)
			return
		}
	default:
		app.handleError(w, fmt.Errorf("event %s not found", event), http.StatusNotFound)
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
