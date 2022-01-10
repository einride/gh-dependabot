package main

import (
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/einride/gh-dependabot/internal/gh"
	"github.com/shurcooL/githubv4"
)

//nolint: gochecknoglobals
var appStyle = lipgloss.NewStyle().Padding(1, 2)

type keyMap struct {
	merge  key.Binding
	rebase key.Binding
	browse key.Binding // open PR in default browser.
}

func newKeyMap() *keyMap {
	return &keyMap{
		merge: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "merge"),
		),
		rebase: key.NewBinding(
			key.WithKeys("r"),
			key.WithHelp("r", "rebase"),
		),
		browse: key.NewBinding(
			key.WithKeys("b"),
			key.WithHelp("b", "open in browser"),
		),
	}
}

func (d keyMap) Bindings() []key.Binding {
	return []key.Binding{
		d.merge,
		d.rebase,
		d.browse,
	}
}

type model struct {
	listModel list.Model
	client    *githubv4.Client
	keyMap    *keyMap
}

var _ tea.Model = model{}

func newModel(client *githubv4.Client, query pullRequestQuery, pullRequests []pullRequest) model {
	keyMap := newKeyMap()
	items := make([]list.Item, len(pullRequests))
	for i, pr := range pullRequests {
		items[i] = pr
	}
	listModel := list.NewModel(items, list.NewDefaultDelegate(), 0, 0)
	listModel.Title = fmt.Sprintf("Pull Requests | %s", query.Filter())
	listModel.SetSpinner(spinner.Points)
	listModel.AdditionalFullHelpKeys = keyMap.Bindings
	listModel.AdditionalShortHelpKeys = listModel.AdditionalFullHelpKeys
	return model{
		listModel: listModel,
		keyMap:    keyMap,
		client:    client,
	}
}

// Init implements tea.Model.
func (m model) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model.
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		topGap, rightGap, bottomGap, leftGap := appStyle.GetPadding()
		m.listModel.SetSize(msg.Width-leftGap-rightGap, msg.Height-topGap-bottomGap)
	case mergePullRequestMessage:
		m.listModel.StopSpinner()
		if msg.err != nil {
			cmds = append(cmds, m.listModel.NewStatusMessage(msg.err.Error()))
		} else {
			cmds = append(cmds, m.listModel.NewStatusMessage("Approved "+msg.pr.url))
		}
	case rebasePullRequestMessage:
		m.listModel.StopSpinner()
		if msg.err != nil {
			cmds = append(cmds, m.listModel.NewStatusMessage(msg.err.Error()))
		} else {
			cmds = append(cmds, m.listModel.NewStatusMessage("Rebasing "+msg.pr.url))
		}
	case openInBrowserMessage:
		if msg.err != nil {
			cmds = append(cmds, m.listModel.NewStatusMessage(msg.err.Error()))
		} else {
			cmds = append(cmds, m.listModel.NewStatusMessage("opened "+msg.pr.url))
		}
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keyMap.merge):
			if selectedItem, ok := m.listModel.SelectedItem().(pullRequest); ok {
				m.listModel.RemoveItem(m.listModel.Index())
				cmds = append(cmds, m.listModel.StartSpinner())
				cmds = append(cmds, m.mergePullRequest(selectedItem))
			}
		case key.Matches(msg, m.keyMap.rebase):
			if selectedItem, ok := m.listModel.SelectedItem().(pullRequest); ok {
				m.listModel.RemoveItem(m.listModel.Index())
				cmds = append(cmds, m.listModel.StartSpinner())
				cmds = append(cmds, m.rebasePullRequest(selectedItem))
			}
		case key.Matches(msg, m.keyMap.browse):
			if selectedItem, ok := m.listModel.SelectedItem().(pullRequest); ok {
				cmds = append(cmds, m.openInBrowser(selectedItem))
			}
		}
	}
	newListModel, cmd := m.listModel.Update(msg)
	m.listModel = newListModel
	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

// View implements tea.Model.
func (m model) View() string {
	return appStyle.Render(m.listModel.View())
}

func (m model) mergePullRequest(pr pullRequest) tea.Cmd {
	return func() tea.Msg {
		result, err := gh.Run("pr", "review", "--approve", "--body", "@dependabot merge", pr.url)
		return mergePullRequestMessage{
			pr:     pr,
			result: result,
			err:    err,
		}
	}
}

func (m model) rebasePullRequest(pr pullRequest) tea.Cmd {
	return func() tea.Msg {
		result, err := gh.Run("pr", "comment", "--body", "@dependabot rebase", pr.url)
		return rebasePullRequestMessage{
			pr:     pr,
			result: result,
			err:    err,
		}
	}
}

func (m model) openInBrowser(pr pullRequest) tea.Cmd {
	return func() tea.Msg {
		result, err := gh.Run("pr", "view", "--web", pr.url)
		return openInBrowserMessage{
			pr:     pr,
			result: result,
			err:    err,
		}
	}
}

type mergePullRequestMessage struct {
	pr     pullRequest
	result string
	err    error
}

type rebasePullRequestMessage struct {
	pr     pullRequest
	result string
	err    error
}

type openInBrowserMessage struct {
	pr     pullRequest
	result string
	err    error
}
