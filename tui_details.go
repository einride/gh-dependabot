package main

import (
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

//nolint:gochecknoglobals
var detailsViewStyle = lipgloss.NewStyle().Padding(1, 4)

func newDetailsView() DetailsView {
	viewportModel := viewport.New(0, 0)
	helpModel := help.New()
	return DetailsView{
		viewportModel: viewportModel,
		helpModel:     helpModel,
	}
}

var _ help.KeyMap = DetailsView{}

type DetailsView struct {
	viewportModel viewport.Model
	helpModel     help.Model
}

func (d DetailsView) ShortHelp() []key.Binding {
	return []key.Binding{
		key.NewBinding(
			key.WithHelp("↑/k", "up"),
		),
		key.NewBinding(
			key.WithHelp("↓/j", "down"),
		),
		key.NewBinding(
			key.WithHelp("v", "hide details"),
		),
		key.NewBinding(
			key.WithHelp("q", "quit"),
		),
	}
}

func (d DetailsView) FullHelp() [][]key.Binding {
	return [][]key.Binding{d.ShortHelp()}
}

func (d DetailsView) Init() tea.Cmd {
	return nil
}

func (d DetailsView) Update(msg tea.Msg) (DetailsView, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		occupiedHorizontal, occupiedVertical := d.getOccupiedSize()
		d.viewportModel.Width = msg.Width - occupiedHorizontal
		d.viewportModel.Height = msg.Height - occupiedVertical
		return d, nil
	case viewPullRequestDetails:
		d.viewportModel.SetContent(d.viewPullRequest(msg.pr))
	case hidePullRequestDetails:
		d.viewportModel.SetContent("")
	case tea.KeyMsg:
		if msg.String() == "v" {
			cmd = hidePullRequestDetailsCmd()
		}
	}

	nextViewport, viewportCmd := d.viewportModel.Update(msg)
	nextHelp, helpCmd := d.helpModel.Update(msg)

	d.viewportModel = nextViewport
	d.helpModel = nextHelp

	return d, tea.Batch(cmd, viewportCmd, helpCmd)
}

func (d DetailsView) View() string {
	content := d.viewportModel.View()
	h := d.helpModel.View(d)

	return detailsViewStyle.Render(lipgloss.JoinVertical(
		lipgloss.Top,
		content,
		h,
	))
}

func (d DetailsView) viewPullRequest(pr pullRequest) string {
	content := lipgloss.NewStyle().
		Background(lipgloss.Color("62")).
		Foreground(lipgloss.Color("230")).
		Padding(0, 1).Render(pr.Title())
	content += "\n\n"
	content += pr.Description()
	content += "\n\n---\n\n"

	content += pr.bodyText

	return content
}

func (d DetailsView) getOccupiedSize() (int, int) {
	horizontal := detailsViewStyle.GetHorizontalFrameSize()
	vertical := detailsViewStyle.GetVerticalFrameSize() + lipgloss.Height(d.helpModel.View(d))
	return horizontal, vertical
}
