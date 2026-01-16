package tui

import (
	"fmt"

	"gogws/internal/config"
	"gogws/internal/git"
	"gogws/internal/gws"
	"gogws/internal/theme"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type sessionState int

const (
	stateLoading sessionState = iota
	stateReady
	stateError
)

type Model struct {
	cfg           *config.Config
	projects      []gws.Project
	statuses      []git.RepositoryStatus
	list          list.Model
	selectedIndex int
	state         sessionState
	err           error
	theme         theme.Theme
	width         int
	height        int
	keys          keyMap
}

type keyMap struct {
	Up        key.Binding
	Down      key.Binding
	Enter     key.Binding
	Refresh   key.Binding
	ToggleAll key.Binding
	Quit      key.Binding
	Help      key.Binding
}

func defaultKeyMap() keyMap {
	return keyMap{
		Up: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑/k", "up"),
		),
		Down: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓/j", "down"),
		),
		Enter: key.NewBinding(
			key.WithKeys("enter", " "),
			key.WithHelp("enter/space", "select"),
		),
		Refresh: key.NewBinding(
			key.WithKeys("r"),
			key.WithHelp("r", "refresh"),
		),
		ToggleAll: key.NewBinding(
			key.WithKeys("a"),
			key.WithHelp("a", "toggle all"),
		),
		Quit: key.NewBinding(
			key.WithKeys("q", "ctrl+c"),
			key.WithHelp("q", "quit"),
		),
		Help: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "help"),
		),
	}
}

type statusLoadedMsg struct {
	statuses []git.RepositoryStatus
}

type errMsg struct {
	err error
}

func NewModel(cfg *config.Config, projects []gws.Project) Model {
	keys := defaultKeyMap()

	items := make([]list.Item, len(projects))
	for i, project := range projects {
		items[i] = repositoryItem{
			project: project,
			status:  git.RepositoryStatus{Path: project.Path},
		}
	}

	delegate := newItemDelegate()
	l := list.New(items, delegate, 0, 0)
	l.Title = "Git Workspace Manager"
	l.SetShowStatusBar(true)
	l.SetFilteringEnabled(true)
	l.Styles.Title = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("63")).
		Padding(0, 1)

	return Model{
		cfg:      cfg,
		projects: projects,
		list:     l,
		state:    stateLoading,
		theme:    theme.GetTheme(),
		keys:     keys,
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		loadStatuses(m.cfg, m.projects),
	)
}

func loadStatuses(cfg *config.Config, projects []gws.Project) tea.Cmd {
	return func() tea.Msg {
		statuses := make([]git.RepositoryStatus, len(projects))
		for i, project := range projects {
			repoPath := cfg.WorkspaceRoot + "/" + project.Path
			statuses[i] = git.GetStatus(repoPath)
		}
		return statusLoadedMsg{statuses: statuses}
	}
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		h, v := lipgloss.NewStyle().Margin(1, 2).GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v-4)
		return m, nil

	case statusLoadedMsg:
		m.statuses = msg.statuses
		m.state = stateReady

		items := make([]list.Item, len(m.projects))
		for i, project := range m.projects {
			items[i] = repositoryItem{
				project: project,
				status:  m.statuses[i],
			}
		}
		m.list.SetItems(items)
		return m, nil

	case errMsg:
		m.err = msg.err
		m.state = stateError
		return m, nil

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Quit):
			return m, tea.Quit

		case key.Matches(msg, m.keys.Refresh):
			m.state = stateLoading
			return m, loadStatuses(m.cfg, m.projects)

		case key.Matches(msg, m.keys.Enter):
			m.selectedIndex = m.list.Index()
			return m, nil
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m Model) View() string {
	if m.state == stateError {
		return m.theme.Error.Render(fmt.Sprintf("Error: %v\n\nPress q to quit.", m.err))
	}

	if m.state == stateLoading {
		return m.theme.Info.Render("Loading repository status...\n\nPress q to quit.")
	}

	mainView := m.list.View()
	detailView := m.renderDetails()

	if m.width > 120 {
		leftStyle := lipgloss.NewStyle().
			Width(m.width / 2).
			Height(m.height - 4)

		rightStyle := lipgloss.NewStyle().
			Width(m.width / 2).
			Height(m.height - 4).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("63")).
			Padding(1)

		return lipgloss.JoinHorizontal(
			lipgloss.Top,
			leftStyle.Render(mainView),
			rightStyle.Render(detailView),
		) + "\n" + m.renderStatusBar()
	}

	return mainView + "\n" + m.renderStatusBar()
}

func (m Model) renderDetails() string {
	if m.selectedIndex < 0 || m.selectedIndex >= len(m.statuses) {
		return m.theme.Subtle.Render("Select a repository to see details")
	}

	status := m.statuses[m.selectedIndex]
	project := m.projects[m.selectedIndex]

	var details string
	details += m.theme.Title.Render("Repository Details") + "\n\n"
	details += m.theme.Path.Render("Path: ") + status.Path + "\n"

	if status.Exists {
		details += m.theme.Branch.Render("Branch: ") + status.Branch + "\n\n"

		if status.Clean {
			details += m.theme.Success.Render("✓ Working tree clean") + "\n"
		} else {
			details += m.theme.Warning.Render("● Working tree dirty") + "\n"
			if status.Uncommitted > 0 {
				details += fmt.Sprintf("  %s\n", m.theme.Warning.Render(fmt.Sprintf("%d uncommitted changes", status.Uncommitted)))
			}
			if status.Untracked > 0 {
				details += fmt.Sprintf("  %s\n", m.theme.Info.Render(fmt.Sprintf("%d untracked files", status.Untracked)))
			}
		}

		details += "\n"
		if status.Ahead > 0 {
			details += m.theme.Success.Render(fmt.Sprintf("↑ %d commits ahead of origin", status.Ahead)) + "\n"
		}
		if status.Behind > 0 {
			details += m.theme.Error.Render(fmt.Sprintf("↓ %d commits behind origin", status.Behind)) + "\n"
		}
		if status.Ahead == 0 && status.Behind == 0 && status.HasRemote {
			details += m.theme.Success.Render("✓ In sync with origin") + "\n"
		}
	} else {
		details += m.theme.Error.Render("✗ Repository not cloned") + "\n"
	}

	details += "\n" + m.theme.Subtle.Render("Remotes:") + "\n"
	for _, remote := range project.Remotes {
		details += fmt.Sprintf("  %s: %s\n", m.theme.Remote.Render(remote.Name), remote.URL)
	}

	return details
}

func (m Model) renderStatusBar() string {
	total := len(m.projects)
	clean := 0
	changed := 0
	missing := 0

	for _, status := range m.statuses {
		if !status.Exists {
			missing++
		} else if status.Clean && status.Ahead == 0 && status.Behind == 0 {
			clean++
		} else {
			changed++
		}
	}

	statusText := fmt.Sprintf(
		"Total: %d | Clean: %d | Changed: %d | Missing: %d | %s",
		total,
		clean,
		changed,
		missing,
		m.theme.Subtle.Render("r: refresh | ?: help | q: quit"),
	)

	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Render(statusText)
}

func Run(cfg *config.Config, projects []gws.Project) error {
	m := NewModel(cfg, projects)
	p := tea.NewProgram(m, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		return err
	}

	return nil
}
