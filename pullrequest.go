package main

import (
	"context"
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/lipgloss"
	"github.com/shurcooL/githubv4"
)

type pullRequest struct {
	title          string
	url            string
	createdAt      time.Time
	updatedAt      time.Time
	owner          string
	repository     string
	state          githubv4.PullRequestState
	checkStatus    githubv4.StatusState
	mergeable      githubv4.MergeableState
	reviewDecision githubv4.PullRequestReviewDecision
}

var _ list.DefaultItem = pullRequest{}

// Title implements list.DefaultItem.
func (p pullRequest) Title() string {
	return p.repository
}

// Description implements list.DefaultItem.
func (p pullRequest) Description() string {
	result := fmt.Sprintf("%s %s", checkStatusEmoji(p.checkStatus), p.title)
	switch p.mergeable {
	case "", githubv4.MergeableStateMergeable: // do nothing
	default:
		result += " [" + lipgloss.NewStyle().Foreground(lipgloss.Color("11")).Render(string(p.mergeable)) + "]"
	}
	switch p.reviewDecision {
	case "", githubv4.PullRequestReviewDecisionReviewRequired: // do nothing
	case githubv4.PullRequestReviewDecisionApproved:
		result += " [" + lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Render(string(p.reviewDecision)) + "]"
	default:
		result += " [" + lipgloss.NewStyle().Foreground(lipgloss.Color("11")).Render(string(p.reviewDecision)) + "]"
	}
	return result
}

// FilterValue implements list.DefaultItem.
func (p pullRequest) FilterValue() string {
	return p.repository + " " + p.title
}

func checkStatusEmoji(c githubv4.StatusState) string {
	switch c {
	case githubv4.StatusStateSuccess:
		return lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Render("✓")
	case githubv4.StatusStateFailure:
		return lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Render("✘")
	case githubv4.StatusStatePending:
		return lipgloss.NewStyle().Foreground(lipgloss.Color("11")).Render("…")
	default:
		return lipgloss.NewStyle().Foreground(lipgloss.Color("11")).Render("?")
	}
}

type pullRequestPage struct {
	PullRequests []pullRequest
	TotalCount   int
	EndCursor    string
	HasNextPage  bool
}

type pullRequestQuery struct {
	username string
	org      string
	team     string
	cursor   string
}

func (q pullRequestQuery) Filter() string {
	switch {
	case q.org != "":
		return "org:" + q.org
	case q.team != "":
		return "team-review-requested:" + q.team
	default:
		return "review-requested:" + q.username
	}
}

func (q pullRequestQuery) SearchQuery() string {
	return "type:pr state:open archived:false author:app/dependabot " + q.Filter()
}

func loadPullRequestPage(client *githubv4.Client, prQuery pullRequestQuery) (*pullRequestPage, error) {
	type searchQuery struct {
		IssueCount int
		PageInfo   struct {
			EndCursor   string
			HasNextPage bool
		}
		Nodes []struct {
			PullRequest struct {
				State          githubv4.PullRequestState
				Mergeable      githubv4.MergeableState
				ReviewDecision githubv4.PullRequestReviewDecision
				URL            string
				Title          string
				CreatedAt      githubv4.DateTime
				UpdatedAt      githubv4.DateTime
				Repository     struct {
					Name  string
					Owner struct {
						Login string
					}
				}
				Commits struct {
					Nodes []struct {
						Commit struct {
							StatusCheckRollup struct {
								State githubv4.StatusState
							}
						}
					}
				} `graphql:"commits(last: 1)"`
			} `graphql:"... on PullRequest"`
		}
	}
	variables := map[string]interface{}{
		"searchQuery": githubv4.String(prQuery.SearchQuery()),
		"first":       githubv4.Int(25),
	}
	var search searchQuery
	if prQuery.cursor == "" {
		var query struct {
			Search searchQuery `graphql:"search(query: $searchQuery, type: ISSUE, first: $first)"`
		}
		if err := client.Query(context.Background(), &query, variables); err != nil {
			return nil, fmt.Errorf("load pull request page: %w", err)
		}
		search = query.Search
	} else {
		var query struct {
			Search searchQuery `graphql:"search(query: $searchQuery, type: ISSUE, first: $first, after: $after)"`
		}
		variables["after"] = githubv4.String(prQuery.cursor)
		if err := client.Query(context.Background(), &query, variables); err != nil {
			return nil, fmt.Errorf("load pull request page: %w", err)
		}
		search = query.Search
	}
	page := pullRequestPage{
		TotalCount:  search.IssueCount,
		EndCursor:   search.PageInfo.EndCursor,
		HasNextPage: search.PageInfo.HasNextPage,
	}
	for _, node := range search.Nodes {
		page.PullRequests = append(page.PullRequests, pullRequest{
			title:          node.PullRequest.Title,
			url:            node.PullRequest.URL,
			createdAt:      node.PullRequest.CreatedAt.Time,
			updatedAt:      node.PullRequest.UpdatedAt.Time,
			owner:          node.PullRequest.Repository.Owner.Login,
			repository:     node.PullRequest.Repository.Name,
			state:          node.PullRequest.State,
			mergeable:      node.PullRequest.Mergeable,
			reviewDecision: node.PullRequest.ReviewDecision,
		})
		if len(node.PullRequest.Commits.Nodes) > 0 {
			page.PullRequests[len(page.PullRequests)-1].checkStatus =
				node.PullRequest.Commits.Nodes[0].Commit.StatusCheckRollup.State
		}
	}
	return &page, nil
}
