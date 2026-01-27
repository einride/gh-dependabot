package main

import (
	"context"
	"fmt"
	"time"

	"github.com/atotto/clipboard"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/einride/gh-dependabot/internal/gh"
	"golang.org/x/time/rate"
)

type errorMessage struct {
	err error
}

type pullRequestMerged struct {
	pr pullRequest
}

type mergeMethod string

const (
	MethodRebase     mergeMethod = "--rebase"
	MethodMerge      mergeMethod = "--merge"
	MethodSquash     mergeMethod = "--squash"
	MethodDependabot mergeMethod = "dependabot auto-merge"
)

type commander struct {
	limiter *rate.Limiter
}

func newCommander() commander {
	limiter := rate.NewLimiter(rate.Every(time.Second*2), 1)
	return commander{
		limiter: limiter,
	}
}

func (c commander) mergePullRequest(pr pullRequest, method mergeMethod) tea.Cmd {
	return func() tea.Msg {
		switch method {
		case MethodRebase:
			fallthrough
		case MethodMerge:
			fallthrough
		case MethodSquash:
			//nolint:errcheck // hard coded limiter burst, can't fail
			c.limiter.Wait(context.Background())
			if _, err := gh.Run("pr", "review", "--approve", pr.url); err != nil {
				return errorMessage{err: err}
			}
			if _, err := gh.Run("pr", "merge", "--auto", string(method), pr.url); err != nil {
				return errorMessage{err: err}
			}
		case MethodDependabot:
			//nolint:errcheck // hard coded limiter burst, can't fail
			c.limiter.Wait(context.Background())
			// Check if rebase is needed before requesting it
			if needsRebase(pr.url) {
				if _, err := gh.Run("pr", "comment", "--body", "@dependabot rebase", pr.url); err != nil {
					return errorMessage{err: err}
				}
			}
			// Approve the PR
			if _, err := gh.Run("pr", "review", "--approve", pr.url); err != nil {
				return errorMessage{err: err}
			}
			// Try auto-merge first, then direct merge
			// Try methods in order: rebase → squash → merge
			if err := mergeWithFallback(pr.url); err != nil {
				return errorMessage{err: err}
			}
		default:
			return errorMessage{err: fmt.Errorf("unknown merge method: %q", method)}
		}
		return pullRequestMerged{pr: pr}
	}
}

// needsRebase checks if a PR branch is behind the base branch.
// See https://docs.github.com/en/graphql/reference/enums#mergestatestatus
func needsRebase(url string) bool {
	status, err := gh.Run("pr", "view", url, "--json", "mergeStateStatus", "--jq", ".mergeStateStatus")
	if err != nil {
		return false // If we can't check, skip the rebase request
	}
	return status == "BEHIND"
}

// mergeWithFallback tries to merge a PR using different methods in order of preference.
// First tries auto-merge (rebase → squash → merge), then direct merge with the same order.
func mergeWithFallback(url string) error {
	methods := []string{"--rebase", "--squash", "--merge"}
	// Try auto-merge first with each method
	for _, method := range methods {
		if _, err := gh.Run("pr", "merge", "--auto", method, url); err == nil {
			return nil
		}
	}
	// Fall back to direct merge with each method
	for _, method := range methods {
		if _, err := gh.Run("pr", "merge", method, url); err == nil {
			return nil
		}
	}
	return fmt.Errorf("failed to merge PR: all merge methods failed")
}

type pullRequestRebased struct {
	pr pullRequest
}

func (c commander) rebasePullRequest(pr pullRequest) tea.Cmd {
	return func() tea.Msg {
		//nolint:errcheck // hard coded limiter burst, can't fail
		c.limiter.Wait(context.Background())
		if _, err := gh.Run("pr", "comment", "--body", "@dependabot rebase", pr.url); err != nil {
			return errorMessage{err: err}
		}
		return pullRequestRebased{pr: pr}
	}
}

type pullRequestRecreated struct {
	pr pullRequest
}

func (c commander) recreatePullRequest(pr pullRequest) tea.Cmd {
	return func() tea.Msg {
		//nolint:errcheck // hard coded limiter burst, can't fail
		c.limiter.Wait(context.Background())
		if _, err := gh.Run("pr", "comment", "--body", "@dependabot recreate", pr.url); err != nil {
			return errorMessage{err: err}
		}
		return pullRequestRecreated{pr: pr}
	}
}

type pullRequestOpenedInBrowser struct {
	pr pullRequest
}

func (c commander) openInBrowser(pr pullRequest) tea.Cmd {
	return func() tea.Msg {
		if _, err := gh.Run("pr", "view", "--web", pr.url); err != nil {
			return errorMessage{err: err}
		}
		return pullRequestOpenedInBrowser{pr: pr}
	}
}

type hidePullRequestDetails struct{}

func hidePullRequestDetailsCmd() tea.Cmd {
	return func() tea.Msg {
		return hidePullRequestDetails{}
	}
}

type viewPullRequestDetails struct {
	pr pullRequest
}

func viewPullRequestDetailsCmd(pr pullRequest) tea.Cmd {
	return func() tea.Msg {
		return viewPullRequestDetails{pr: pr}
	}
}

type copyCheckout struct {
	pr pullRequest
}

func copyCheckoutCmd(pr pullRequest) tea.Cmd {
	return func() tea.Msg {
		if err := clipboard.WriteAll("gh pr checkout " + pr.number); err != nil {
			return errorMessage{err: err}
		}
		return copyCheckout{pr: pr}
	}
}
