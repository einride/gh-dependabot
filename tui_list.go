package main

import (
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

//nolint:gochecknoglobals
var listViewStyle = lipgloss.NewStyle().Padding(1, 2)

type keyMap struct {
	mergeRebase     key.Binding
	mergeDefault    key.Binding
	mergeSquash     key.Binding
	mergeDependabot key.Binding
	rebase          key.Binding
	recreate        key.Binding
	view            key.Binding
	browse          key.Binding // open PR in default browser.
	copyCheckout    key.Binding
}

func newKeyMap() *keyMap {
	return &keyMap{
		mergeRebase: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "merge (rebase)"),
		),
		mergeDefault: key.NewBinding(
			key.WithKeys("ctrl+m"),
			key.WithHelp("ctrl+m", "merge (merge-commit)"),
		),
		mergeSquash: key.NewBinding(
			key.WithKeys("M"),
			key.WithHelp("shift+m", "merge (squash)"),
		),
		mergeDependabot: key.NewBinding(
			key.WithKeys("alt+m"),
			key.WithHelp("alt+m", "merge (Dependabot)"),
		),
		rebase: key.NewBinding(
			key.WithKeys("r"),
			key.WithHelp("r", "rebase"),
		),
		recreate: key.NewBinding(
			key.WithKeys("R"),
			key.WithHelp("shift+r", "recreate"),
		),
		browse: key.NewBinding(
			key.WithKeys("b"),
			key.WithHelp("b", "open in browser"),
		),
		view: key.NewBinding(
			key.WithKeys("v"),
			key.WithHelp("v", "view details"),
		),
		copyCheckout: key.NewBinding(
			key.WithKeys("c"),
			key.WithHelp("c", "copy checkout command"),
		),
	}
}

func (d keyMap) Bindings() []key.Binding {
	return []key.Binding{
		d.mergeRebase,
		d.mergeDefault,
		d.mergeSquash,
		d.mergeDependabot,
		d.rebase,
		d.recreate,
		d.view,
		d.browse,
		d.copyCheckout,
	}
}

type ListView struct {
	listModel list.Model
	keyMap    *keyMap
}

func newListView(query pullRequestQuery, pullRequests []pullRequest) ListView {
	keyMap := newKeyMap()
	listModel := list.New(convertListItems(pullRequests), list.NewDefaultDelegate(), 0, 0)
	listModel.Title = fmt.Sprintf("Pull Requests | %s", query.Filter())
	listModel.SetSpinner(spinner.Points)
	listModel.AdditionalFullHelpKeys = keyMap.Bindings
	listModel.AdditionalShortHelpKeys = listModel.AdditionalFullHelpKeys
	return ListView{
		listModel: listModel,
		keyMap:    keyMap,
	}
}

func (m ListView) Init() tea.Cmd {
	return nil
}

func (m ListView) Update(msg tea.Msg) (ListView, tea.Cmd) {
	var cmds []tea.Cmd
	switch msg := msg.(type) {
	case errorMessage:
		m.listModel.StopSpinner()
		cmds = append(cmds, m.listModel.NewStatusMessage(msg.err.Error()))
	case pullRequestMerged:
		m.listModel.StopSpinner()
		cmds = append(cmds, m.listModel.NewStatusMessage("Approved "+msg.pr.url))
	case pullRequestRebased:
		m.listModel.StopSpinner()
		cmds = append(cmds, m.listModel.NewStatusMessage("Rebased "+msg.pr.url))
	case pullRequestRecreated:
		m.listModel.StopSpinner()
		cmds = append(cmds, m.listModel.NewStatusMessage("Recreated "+msg.pr.url))
	case pullRequestOpenedInBrowser:
		m.listModel.StopSpinner()
		cmds = append(cmds, m.listModel.NewStatusMessage("Opened "+msg.pr.url))
	case tea.WindowSizeMsg:
		topGap, rightGap, bottomGap, leftGap := listViewStyle.GetPadding()
		m.listModel.SetSize(msg.Width-leftGap-rightGap, msg.Height-topGap-bottomGap)
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keyMap.mergeRebase):
			if selectedItem, ok := m.listModel.SelectedItem().(pullRequest); ok {
				m.listModel.RemoveItem(m.listModel.Index())
				cmds = append(
					cmds,
					m.listModel.StartSpinner(),
					mergePullRequest(selectedItem, MethodRebase),
				)
			}
		case key.Matches(msg, m.keyMap.mergeDefault):
			if selectedItem, ok := m.listModel.SelectedItem().(pullRequest); ok {
				m.listModel.RemoveItem(m.listModel.Index())
				cmds = append(
					cmds,
					m.listModel.StartSpinner(),
					mergePullRequest(selectedItem, MethodMerge),
				)
			}
		case key.Matches(msg, m.keyMap.mergeSquash):
			if selectedItem, ok := m.listModel.SelectedItem().(pullRequest); ok {
				m.listModel.RemoveItem(m.listModel.Index())
				cmds = append(
					cmds,
					m.listModel.StartSpinner(),
					mergePullRequest(selectedItem, MethodSquash),
				)
			}
		case key.Matches(msg, m.keyMap.mergeDependabot):
			if selectedItem, ok := m.listModel.SelectedItem().(pullRequest); ok {
				m.listModel.RemoveItem(m.listModel.Index())
				cmds = append(
					cmds,
					m.listModel.StartSpinner(),
					mergePullRequest(selectedItem, MethodDependabot),
				)
			}
		case key.Matches(msg, m.keyMap.rebase):
			if selectedItem, ok := m.listModel.SelectedItem().(pullRequest); ok {
				cmds = append(
					cmds,
					m.listModel.StartSpinner(),
					rebasePullRequest(selectedItem),
				)
			}
		case key.Matches(msg, m.keyMap.recreate):
			if selectedItem, ok := m.listModel.SelectedItem().(pullRequest); ok {
				cmds = append(
					cmds,
					m.listModel.StartSpinner(),
					recreatePullRequest(selectedItem),
				)
			}
		case key.Matches(msg, m.keyMap.browse):
			if selectedItem, ok := m.listModel.SelectedItem().(pullRequest); ok {
				cmds = append(cmds, openInBrowser(selectedItem))
			}
		case key.Matches(msg, m.keyMap.view):
			if selectedItem, ok := m.listModel.SelectedItem().(pullRequest); ok {
				cmds = append(cmds, viewPullRequestDetailsCmd(selectedItem))
			}
		case key.Matches(msg, m.keyMap.copyCheckout):
			if selectedItem, ok := m.listModel.SelectedItem().(pullRequest); ok {
				cmds = append(cmds, copyCheckoutCmd(selectedItem))
			}
		}
	}
	newListModel, cmd := m.listModel.Update(msg)
	m.listModel = newListModel
	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

func (m ListView) View() string {
	return listViewStyle.Render(m.listModel.View())
}

func convertListItems(pullRequests []pullRequest) []list.Item {
	items := make([]list.Item, len(pullRequests))
	for i, pr := range pullRequests {
		items[i] = pr
	}
	return items
}
