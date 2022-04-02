package main

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/shurcooL/githubv4"
)

var _ tea.Model = App{}

func newApp(_ *githubv4.Client, query pullRequestQuery, pullRequests []pullRequest) App {
	listView := newListView(query, pullRequests)
	detailsView := newDetailsView()

	return App{
		listView:    listView,
		detailsView: detailsView,
	}
}

type App struct {
	listView    ListView
	detailsView DetailsView

	isShowingDetails bool
}

func (a App) Init() tea.Cmd {
	return tea.Batch(
		a.listView.Init(),
	)
}

func (a App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg.(type) {
	case viewPullRequestDetails:
		a.isShowingDetails = true
	case hidePullRequestDetails:
		a.isShowingDetails = false
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" || msg.String() == "q" {
			return a, tea.Quit
		}

		// forward key messages to active view
		var cmd tea.Cmd
		if a.isShowingDetails {
			cmd = a.updateDetailsView(msg)
		} else {
			cmd = a.updateListView(msg)
		}
		return a, cmd
	default:
		// forward other messages to all views
		listViewCmd := a.updateListView(msg)
		detailsViewCmd := a.updateDetailsView(msg)
		return a, tea.Batch(listViewCmd, detailsViewCmd)
	}
}

func (a *App) updateListView(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	a.listView, cmd = a.listView.Update(msg)
	return cmd
}

func (a *App) updateDetailsView(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	a.detailsView, cmd = a.detailsView.Update(msg)
	return cmd
}

func (a App) View() string {
	if a.isShowingDetails {
		return a.detailsView.View()
	}
	return a.listView.View()
}
