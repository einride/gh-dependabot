package main

import (
	"github.com/atotto/clipboard"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/einride/gh-dependabot/internal/gh"
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

func mergePullRequest(pr pullRequest, method mergeMethod) tea.Cmd {
	return func() tea.Msg {
		if _, err := gh.Run("pr", "review", "--approve", pr.url); err != nil {
			return errorMessage{err: err}
		}
		var args []string
		if method == MethodDependabot {
			args = []string{"pr", "comment", "--body", string(method), pr.url}
		} else {
			args = []string{"pr", "merge", "--auto", string(method), pr.url}
		}
		if _, err := gh.Run(args...); err != nil {
			return errorMessage{err: err}
		}
		return pullRequestMerged{pr: pr}
	}
}

type pullRequestRebased struct {
	pr pullRequest
}

func rebasePullRequest(pr pullRequest) tea.Cmd {
	return func() tea.Msg {
		if _, err := gh.Run("pr", "comment", "--body", "@dependabot rebase", pr.url); err != nil {
			return errorMessage{err: err}
		}
		return pullRequestRebased{pr: pr}
	}
}

type pullRequestRecreated struct {
	pr pullRequest
}

func recreatePullRequest(pr pullRequest) tea.Cmd {
	return func() tea.Msg {
		if _, err := gh.Run("pr", "comment", "--body", "@dependabot recreate", pr.url); err != nil {
			return errorMessage{err: err}
		}
		return pullRequestRecreated{pr: pr}
	}
}

type pullRequestOpenedInBrowser struct {
	pr pullRequest
}

func openInBrowser(pr pullRequest) tea.Cmd {
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
