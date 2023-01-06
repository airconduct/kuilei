package pluginhelpers

import (
	"context"

	"github.com/shurcooL/githubv4"

	"github.com/airconduct/go-probot"
	"github.com/airconduct/kuilei/pkg/plugins"
)

func GitSearchClientFromGithub(cli probot.GitGraphQLClient) plugins.GitSearchClient {
	return &githubGraphqlClient{
		Client: cli,
	}
}

type githubGraphqlClient struct {
	Client probot.GitGraphQLClient
}

func (c *githubGraphqlClient) SearchPR(ctx context.Context, repo plugins.GitRepo, state string) ([]plugins.GitPullRequest, error) {
	q := &prSearchQuery{}
	if err := c.Client.Query(ctx, q, map[string]interface{}{
		"owner":  githubv4.String(repo.Owner.Name),
		"repo":   githubv4.String(repo.Name),
		"states": []githubv4.PullRequestState{githubv4.PullRequestState(state)},
	}); err != nil {
		return nil, err
	}
	prs := []plugins.GitPullRequest{}
	for _, pr := range q.Repository.PullRequests.Nodes {
		// Get labels
		labels := []plugins.Label{}
		for _, l := range pr.Labels.Nodes {
			labels = append(labels, plugins.Label{
				Name:  l.Name,
				Color: l.Color,
			})
		}
		// Get assignees
		assignees := []plugins.GitUser{}
		for _, u := range pr.Assignees.Nodes {
			assignees = append(assignees, plugins.GitUser{Name: u.Login})
		}
		// Get commits
		commits := []plugins.GitCommit{}
		for _, node := range pr.Commits.Nodes {
			// Get commit status
			statuses := []plugins.GitCommitStatus{}
			for _, s := range node.Commit.Status.Contexts {
				statuses = append(statuses, plugins.GitCommitStatus{
					Context:     s.Context,
					State:       s.State,
					Description: s.Description,
				})
			}
			// Get commit checks
			checks := []plugins.GitCommitCheck{}
			for _, c := range node.Commit.StatusCheckRollup.Contexts.Nodes {
				checks = append(checks, plugins.GitCommitCheck{
					Name:       c.CheckRun.Name,
					Status:     c.CheckRun.Status,
					Conclusion: c.CheckRun.Conclusion,
				})
			}
			commits = append(commits, plugins.GitCommit{
				Sha:      node.Commit.OID,
				Statuses: statuses,
				Checks:   checks,
			})
		}

		prs = append(prs, plugins.GitPullRequest{
			Number:    int(pr.Number),
			State:     state,
			Locked:    pr.Locked,
			Title:     pr.Title,
			Body:      pr.Body,
			Mergeable: pr.Mergeable,
			Head: plugins.GitBranch{
				Ref: pr.HeadRefName,
				Sha: pr.HeadRefOid,
			},
			Labels:    labels,
			Commits:   commits,
			Assignees: assignees,
			User:      plugins.GitUser{Name: pr.Author.Login},
		})
	}
	return prs, nil
}

/**** QraphQL query example:

query($owner: String!, $repo: String!, $states:[PullRequestState!]) {
  repository(owner:$owner,name:$repo) {
    pullRequests(first:10,states:$states,orderBy:{field:CREATED_AT,direction:ASC}){
      nodes{
        number
        state
        locked
        title
        headRefOid
        headRefName
        body
		mergeable
        commits(last:1){
          nodes{
            commit{
              oid
              statusCheckRollup {
                contexts(last:100){
                  nodes{
                    ... on CheckRun{
                      name
                      conclusion
                      status
                    }
                  }
                }
              }
              status {
                contexts {
                  context
                  state
                  description
                }
              }
            }
          }
        }
        labels(first:100) {
          nodes{
            id
            name
            color
          }
        }
        assignees(first:100){
          nodes{
            login
          }
        }
        author{
          login
        }
      }
    }
  }
}
*****/

type prSearchQuery struct {
	Repository struct {
		PullRequests struct {
			Nodes []struct {
				Number      int
				State       string
				Locked      bool
				Title       string
				Body        string
				Mergeable   string
				HeadRefOid  string
				HeadRefName string
				Labels      struct {
					Nodes []struct {
						Name  string
						Color string
					}
				} `graphql:"labels(first:100)"`
				Commits struct {
					Nodes []struct {
						Commit struct {
							OID               string `graphql:"oid"`
							StatusCheckRollup struct {
								Contexts struct {
									Nodes []struct {
										CheckRun struct {
											Name       string
											Conclusion string
											Status     string
										} `graphql:"... on CheckRun"`
									}
								} `graphql:"contexts(last:100)"`
							}
							Status struct {
								Contexts []struct {
									Context     string
									State       string
									Description string
								}
							}
						}
					}
				} `graphql:"commits(last:1)"`
				Assignees struct {
					Nodes []struct {
						Login string
					}
				} `graphql:"assignees(first:100)"`
				Author struct {
					Login string
				}
			}
		} `graphql:"pullRequests(first:100,states:$states,orderBy:{field:CREATED_AT,direction:ASC})"`
	} `graphql:"repository(owner: $owner, name: $repo)"`
}
