package main

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/einride/gh-dependabot/internal/gh"
)

type errorMessage struct {
	err error
}

type pullRequestMerged struct {
	pr pullRequest
}

func mergePullRequest(pr pullRequest, mergeMethod string) tea.Cmd {
	return func() tea.Msg {
		if _, err := gh.Run("pr", "review", "--approve", pr.url); err != nil {
			return errorMessage{err: err}
		}
		if _, err := gh.Run("pr", "merge", "--auto", mergeMethod, pr.url); err != nil {
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
