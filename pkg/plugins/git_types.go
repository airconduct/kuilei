package plugins

type GitRepo struct {
	Name  string
	Owner GitUser
}

type GitUser struct {
	Name string
}

// GitCommentEventAction coerces multiple actions into its generic equivalent.
type GitCommentEventAction string

// Comments indicate values that are coerced to the specified value.
const (
	// GitCommentActionCreated means something was created/opened/submitted
	GitCommentActionCreated GitCommentEventAction = "created" // "opened", "submitted"
	// GitCommentActionEdited means something was edited.
	GitCommentActionEdited GitCommentEventAction = "edited"
	// GitCommentActionDeleted means something was deleted/dismissed.
	GitCommentActionDeleted GitCommentEventAction = "deleted" // "dismissed"
)

// GitCommentEvent is a fake event type that is instantiated for any git event that contains
// comment like content.
// The specific events that are also handled as GenericCommentEvents are:
// - issue_comment events
// - pull_request_review events
// - pull_request_review_comment events
// - pull_request events with action in ["opened", "edited"]
// - issue events with action in ["opened", "edited"]
//
// Issue and PR "closed" events are not coerced to the "deleted" Action and do not trigger
// a GenericCommentEvent because these events don't actually remove the comment content from GH.
type GitCommentEvent struct {
	GitComment

	Action GitCommentEventAction

	Repo GitRepo
}

type GitComment struct {
	ID           int    `json:"id"`
	NodeID       string `json:"node_id"`
	CommentID    int
	IsPR         bool
	Body         string
	HTMLURL      string
	Number       int
	User         GitUser
	IssueAuthor  GitUser
	Assignees    []GitUser
	IssueState   string
	IssueTitle   string
	IssueBody    string
	IssueHTMLURL string
}

type GitIssueCommentEvent struct {
	GitIssueComment

	Action GitCommentEventAction

	Issue GitIssue
	Repo  GitRepo
}

type GitIssueComment struct {
	ID   int
	Body string
	User GitUser
	URL  string
}

type GitIssue struct {
	ID        int
	Number    int
	State     string
	Locked    bool
	Title     string
	Body      string
	Labels    []Label
	Assignees []GitUser
	User      GitUser
}

type Label struct {
	ID    int64
	Name  string
	Color string
}
