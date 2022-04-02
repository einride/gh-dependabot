package main

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/shurcooL/githubv4"
)

var _ tea.Model = App{}

func newApp(_ *githubv4.Client, query pullRequestQuery, pullRequests []pullRequest) App {
	listView := newListView(query, pullRequests)

	return App{
		listView: listView,
	}
}

type App struct {
	listView ListView
}

func (a App) Init() tea.Cmd {
	return tea.Batch(
		a.listView.Init(),
	)
}

func (a App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// forward key messages to active view
		// todo: check which is active
		cmd := a.updateListView(msg)
		return a, cmd
	default:
		// forward all other messages to all views
		cmd := a.updateListView(msg)
		return a, cmd
	}
}

func (a *App) updateListView(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	a.listView, cmd = a.listView.Update(msg)
	return cmd
}

func (a App) View() string {
	return a.listView.View()
}
