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

// mergeState represents the merge state status of a PR.
// See https://docs.github.com/en/graphql/reference/enums#mergestatestatus
type mergeState int

const (
	mergeStateUnknown  mergeState = iota // The state has not been identified
	mergeStateBehind                     // The head ref is out of date
	mergeStateBlocked                    // The merge is blocked
	mergeStateClean                      // Mergeable and passing commit status
	mergeStateDirty                      // The merge commit cannot be cleanly created (conflicts)
	mergeStateDraft                      // The merge is blocked due to the pull request being a draft
	mergeStateHasHooks                   // Mergeable with passing commit status and pre-receive hooks
	mergeStateUnstable                   // Mergeable with non-passing commit status
)

// getMergeState returns the merge state status of a PR.
func getMergeState(url string) mergeState {
	status, err := gh.Run("pr", "view", url, "--json", "mergeStateStatus", "--jq", ".mergeStateStatus")
	if err != nil {
		return mergeStateUnknown
	}
	switch status {
	case "BEHIND":
		return mergeStateBehind
	case "BLOCKED":
		return mergeStateBlocked
	case "CLEAN":
		return mergeStateClean
	case "DIRTY":
		return mergeStateDirty
	case "DRAFT":
		return mergeStateDraft
	case "HAS_HOOKS":
		return mergeStateHasHooks
	case "UNSTABLE":
		return mergeStateUnstable
	default:
		return mergeStateUnknown
	}
}

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
			// Approve the PR first
			if _, err := gh.Run("pr", "review", "--approve", pr.url); err != nil {
				return errorMessage{err: err}
			}
			// Check merge state
			state := getMergeState(pr.url)
			// Request rebase if branch is behind
			var requestedRebase bool
			if state == mergeStateBehind {
				_, _ = gh.Run("pr", "comment", "--body", "@dependabot rebase", pr.url)
				requestedRebase = true
			}
			// Scenario 1: Try auto-merge (preferred - waits for checks)
			if tryAutoMerge(pr.url) {
				return pullRequestMerged{pr: pr}
			}
			// Scenario 2: Auto-merge not available, try direct merge (only if mergeable)
			if state == mergeStateClean || state == mergeStateHasHooks {
				if err := directMerge(pr.url); err == nil {
					return pullRequestMerged{pr: pr}
				}
			}
			// Scenario 3: Behind and no auto-merge - rebase and post helpful comment
			if !requestedRebase {
				_, _ = gh.Run("pr", "comment", "--body", "@dependabot rebase", pr.url)
			}
			_, _ = gh.Run("pr", "comment", "--body",
				"⚠️ This PR needs to be rebased before it can be merged. "+
					"Auto-merge is not enabled for this repository.\n\n"+
					"Consider enabling auto-merge in repository settings to simplify the merge process.",
				pr.url)
			return errorMessage{err: fmt.Errorf("PR rebasing - auto-merge unavailable, merge manually after rebase")}
		default:
			return errorMessage{err: fmt.Errorf("unknown merge method: %q", method)}
		}
		return pullRequestMerged{pr: pr}
	}
}

// tryAutoMerge attempts to enable auto-merge on a PR.
// If the branch is behind, it requests a rebase first.
// Returns true if auto-merge was successfully enabled.
// Tries methods in order: rebase → squash → merge.
func tryAutoMerge(url string) bool {
	methods := []string{"--rebase", "--squash", "--merge"}
	for _, method := range methods {
		if _, err := gh.Run("pr", "merge", "--auto", method, url); err == nil {
			return true
		}
	}
	return false
}

// directMerge attempts to merge a PR directly without auto-merge.
// Tries methods in order: rebase → squash → merge.
func directMerge(url string) error {
	methods := []string{"--rebase", "--squash", "--merge"}
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
