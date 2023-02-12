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

// GitPREventAction coerces multiple actions into its generic equivalent.
type GitPREventAction string

const (
	// GitPRActionCreated means something was created/opened/submitted
	GitPRActionCreated GitPREventAction = "created" // "opened", "submitted"
	// GitPRActionEdited means something was edited.
	GitPRActionEdited GitPREventAction = "edited"
	// v means something was deleted/dismissed.
	GitPRActionDeleted GitPREventAction = "deleted" // "dismissed"
)

type GitPREvent struct {
	GitPullRequest

	Action GitPREventAction
	Repo   GitRepo
}

type GitPullRequestState = string

const (
	PullRequestStateOpen   GitPullRequestState = "OPEN"
	PullRequestStateClosed GitPullRequestState = "CLOSED"
	PullRequestStateMerged GitPullRequestState = "MERGED"
)

type GitMergeableState = string

// Whether or not a PullRequest can be merged.
const (
	GitMergeableStateMergeable   GitMergeableState = "MERGEABLE"   // The pull request can be merged.
	GitMergeableStateConflicting GitMergeableState = "CONFLICTING" // The pull request cannot be merged due to merge conflicts.
	GitMergeableStateUnknown     GitMergeableState = "UNKNOWN"     // The mergeability of the pull request is still being calculated.
)

type GitPullRequest struct {
	ID        int
	Number    int
	State     GitPullRequestState
	Head      GitBranch
	Locked    bool
	Title     string
	Body      string
	Mergeable GitMergeableState
	Labels    []Label
	Assignees []GitUser
	User      GitUser
}

type GitPullRequestSearchResult struct {
	GitPullRequest
	Commits []GitCommit
}

type GitBranch struct {
	Ref string
	SHA string
}

type GitCommit struct {
	Sha      string
	Statuses []GitCommitStatus
	Checks   []GitCommitCheck
}

type GitCommitFile struct {
	Path string
}

type Label struct {
	ID    int64
	Name  string
	Color string
}

type GitStatusState = string

const (
	GitStatusStateExpected GitStatusState = "EXPECTED" // Status is expected.
	GitStatusStateError    GitStatusState = "ERROR"    // Status is errored.
	GitStatusStateFailure  GitStatusState = "FAILURE"  // Status is failing.
	GitStatusStatePending  GitStatusState = "PENDING"  // Status is pending.
	GitStatusStateSuccess  GitStatusState = "SUCCESS"  // Status is successful.
)

type GitCommitStatus struct {
	Context     string
	State       GitStatusState
	TargetURL   string
	Description string
}

type GitCheckStatus = string

// The possible states of a check run in a status rollup.
const (
	GitCheckStatusCompleted  GitCheckStatus = "COMPLETED"   // The check run has been completed.
	GitCheckStatusInProgress GitCheckStatus = "IN_PROGRESS" // The check run is in progress.
	GitCheckStatusQueued     GitCheckStatus = "QUEUED"      // The check run has been queued.
)

type GitCheckConclusion = string

// The possible states for a check suite or run conclusion.
const (
	GitCheckConclusionStateActionRequired GitCheckConclusion = "ACTION_REQUIRED" // The check suite or run requires action.
	GitCheckConclusionStateCancelled      GitCheckConclusion = "CANCELLED"       // The check suite or run has been cancelled.
	GitCheckConclusionStateFailure        GitCheckConclusion = "FAILURE"         // The check suite or run has failed.
	GitCheckConclusionStateNeutral        GitCheckConclusion = "NEUTRAL"         // The check suite or run was neutral.
	GitCheckConclusionStateSuccess        GitCheckConclusion = "SUCCESS"         // The check suite or run has succeeded.
	GitCheckConclusionStateSkipped        GitCheckConclusion = "SKIPPED"         // The check suite or run was skipped.
	GitCheckConclusionStateStale          GitCheckConclusion = "STALE"           // The check suite or run was marked stale by GitHub. Only GitHub can use this conclusion.
	GitCheckConclusionStateTimedOut       GitCheckConclusion = "TIMED_OUT"       // The check suite or run has timed out.
)

type GitCommitCheck struct {
	Name       string
	Status     GitCheckStatus
	Conclusion GitCheckConclusion
}
