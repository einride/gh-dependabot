package main

import (
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
	MethodDependabot mergeMethod = "@dependabot merge"
)

type commander struct {
	limiters map[string]*rate.Limiter
}

func newCommander(pullRequests []pullRequest) commander {
	limiters := make(map[string]*rate.Limiter)
	for _, pr := range pullRequests {
		if _, ok := limiters[pr.repository]; !ok {
			limiters[pr.repository] = rate.NewLimiter(rate.Every(time.Second*5), 1)
		}
	}
	return commander{
		limiters: limiters,
	}
}

func (c commander) mergePullRequest(pr pullRequest, method mergeMethod) tea.Cmd {
	return func() tea.Msg {
		limiter, ok := c.limiters[pr.repository]
		if !ok {
			return errorMessage{err: fmt.Errorf("mismanaged state, no limiter")}
		}
		switch method {
		case MethodRebase:
			fallthrough
		case MethodMerge:
			fallthrough
		case MethodSquash:
			r := limiter.Reserve()
			time.Sleep(r.Delay())
			if _, err := gh.Run("pr", "review", "--approve", pr.url); err != nil {
				return errorMessage{err: err}
			}
			if _, err := gh.Run("pr", "merge", "--auto", string(method), pr.url); err != nil {
				return errorMessage{err: err}
			}
		case MethodDependabot:
			r := limiter.Reserve()
			time.Sleep(r.Delay())
			if _, err := gh.Run("pr", "review", "--approve", "--body", string(method), pr.url); err != nil {
				return errorMessage{err: err}
			}
		default:
			return errorMessage{err: fmt.Errorf("unknown merge method: %q", method)}
		}
		return pullRequestMerged{pr: pr}
	}
}

type pullRequestRebased struct {
	pr pullRequest
}

func (c commander) rebasePullRequest(pr pullRequest) tea.Cmd {
	return func() tea.Msg {
		limiter, ok := c.limiters[pr.repository]
		if !ok {
			return errorMessage{err: fmt.Errorf("mismanaged state, no limiter")}
		}
		r := limiter.Reserve()
		time.Sleep(r.Delay())

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
		limiter, ok := c.limiters[pr.repository]
		if !ok {
			return errorMessage{err: fmt.Errorf("mismanaged state, no limiter")}
		}
		r := limiter.Reserve()
		time.Sleep(r.Delay())
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
		limiter, ok := c.limiters[pr.repository]
		if !ok {
			return errorMessage{err: fmt.Errorf("mismanaged state, no limiter")}
		}
		r := limiter.Reserve()
		time.Sleep(r.Delay())
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
