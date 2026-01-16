package tui

import (
	"fmt"
	"io"

	"gogws/internal/git"
	"gogws/internal/gws"
	"gogws/internal/theme"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type repositoryItem struct {
	project gws.Project
	status  git.RepositoryStatus
}

func (i repositoryItem) FilterValue() string {
	return i.project.Path
}

func (i repositoryItem) Title() string {
	return i.project.Path
}

func (i repositoryItem) Description() string {
	if !i.status.Exists {
		return "✗ Not cloned"
	}

	if i.status.Clean && i.status.Ahead == 0 && i.status.Behind == 0 {
		return fmt.Sprintf("✓ %s - clean", i.status.Branch)
	}

	desc := fmt.Sprintf("● %s", i.status.Branch)
	changes := []string{}

	if i.status.Uncommitted > 0 {
		changes = append(changes, fmt.Sprintf("%d uncommitted", i.status.Uncommitted))
	}
	if i.status.Untracked > 0 {
		changes = append(changes, fmt.Sprintf("%d untracked", i.status.Untracked))
	}
	if i.status.Ahead > 0 {
		changes = append(changes, fmt.Sprintf("↑%d", i.status.Ahead))
	}
	if i.status.Behind > 0 {
		changes = append(changes, fmt.Sprintf("↓%d", i.status.Behind))
	}

	if len(changes) > 0 {
		desc += " - "
		for i, change := range changes {
			if i > 0 {
				desc += ", "
			}
			desc += change
		}
	}

	return desc
}

type itemDelegate struct {
	theme theme.Theme
}

func newItemDelegate() list.DefaultDelegate {
	d := list.NewDefaultDelegate()

	d.Styles.SelectedTitle = d.Styles.SelectedTitle.
		Foreground(lipgloss.Color("63")).
		BorderForeground(lipgloss.Color("63"))

	d.Styles.SelectedDesc = d.Styles.SelectedDesc.
		Foreground(lipgloss.Color("246")).
		BorderForeground(lipgloss.Color("63"))

	return d
}

func (d itemDelegate) Height() int {
	return 2
}

func (d itemDelegate) Spacing() int {
	return 1
}

func (d itemDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd {
	return nil
}

func (d itemDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	i, ok := item.(repositoryItem)
	if !ok {
		return
	}

	str := fmt.Sprintf("%d. %s", index+1, i.Title())

	var fn func(...string) string
	if index == m.Index() {
		fn = d.theme.Path.Render
	} else {
		fn = lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render
	}

	fmt.Fprint(w, fn(str))
}
