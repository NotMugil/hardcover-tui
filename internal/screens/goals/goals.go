package goals

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/NotMugil/hardcover-tui/internal/api"
	"github.com/NotMugil/hardcover-tui/internal/commands"
	"github.com/NotMugil/hardcover-tui/internal/common"
)

// goalItem implements list.DefaultItem for the bubbles list.
type goalItem struct {
	data api.Goal
}

func (i goalItem) Title() string {
	state := i.data.State
	return fmt.Sprintf("Goal #%d: %d/%d %s [%s]",
		i.data.ID, int(i.data.Progress), i.data.Goal, i.data.Metric, state)
}

func (i goalItem) Description() string {
	pct := i.data.PercentComplete()
	bar := common.RenderBar(pct, 20, fmt.Sprintf("%d%%", int(pct*100)))
	desc := fmt.Sprintf("%s  |  %s -> %s", bar, i.data.StartDate, i.data.EndDate)
	if i.data.Description != nil && *i.data.Description != "" {
		desc += "  |  " + *i.data.Description
	}
	return desc
}

func (i goalItem) FilterValue() string {
	return fmt.Sprintf("Goal %d %s", i.data.ID, i.data.Metric)
}

// Model is the goals screen model.
type Model struct {
	deps    commands.Deps
	goals   []api.Goal
	list    list.Model
	spinner spinner.Model
	loading bool
	err     error
	width   int
	height  int
}

// New creates a new goals screen.
func New(deps commands.Deps) *Model {
	s := spinner.New(
		spinner.WithSpinner(spinner.Dot),
		spinner.WithStyle(common.SpinnerStyle),
	)

	delegate := list.NewDefaultDelegate()
	delegate.Styles.SelectedTitle = delegate.Styles.SelectedTitle.
		Foreground(common.ColorPrimary).
		BorderLeftForeground(common.ColorPrimary)
	delegate.Styles.SelectedDesc = delegate.Styles.SelectedDesc.
		Foreground(common.ColorSubtext).
		BorderLeftForeground(common.ColorPrimary)

	l := list.New([]list.Item{}, delegate, 80, 20)
	l.SetShowTitle(false)
	l.SetShowStatusBar(true)
	l.SetShowHelp(false)
	l.SetFilteringEnabled(false)
	l.DisableQuitKeybindings()
	l.Styles.NoItems = common.ValueStyle

	return &Model{
		deps:    deps,
		list:    l,
		spinner: s,
		loading: true,
	}
}

// SetSize updates the available terminal dimensions.
func (m *Model) SetSize(w, h int) {
	m.width = w
	m.height = h
	contentW := w - 4
	contentH := h - 8
	if contentW > 0 && contentH > 0 {
		m.list.SetSize(contentW, contentH)
	}
}

func (m *Model) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, commands.LoadGoals(m.deps, m.deps.User.ID))
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case commands.GoalsLoadedMsg:
		m.loading = false
		if msg.Err != nil {
			m.err = msg.Err
			return m, nil
		}
		m.goals = msg.Goals
		items := make([]list.Item, len(m.goals))
		for i, g := range m.goals {
			items[i] = goalItem{data: g}
		}
		m.list.SetItems(items)
		return m, nil

	case spinner.TickMsg:
		if m.loading {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}
	}

	if !m.loading {
		var cmd tea.Cmd
		m.list, cmd = m.list.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m *Model) View() string {
	if m.loading {
		return common.AppStyle.Render(
			fmt.Sprintf("\n  %s Loading goals...\n", m.spinner.View()),
		)
	}

	var b strings.Builder

	b.WriteString(common.TitleStyle.Render("Reading Goals"))
	b.WriteString("\n\n")

	if m.err != nil {
		b.WriteString(common.ErrorStyle.Render("Error: "+m.err.Error()) + "\n\n")
	}

	if len(m.goals) == 0 {
		b.WriteString(common.PanelStyle.Render(
			common.ValueStyle.Render("No active goals") + "\n" +
				common.HelpStyle.Render("Create goals on hardcover.app to track them here"),
		))
	} else {
		b.WriteString(common.PanelStyle.Render(m.list.View()))
	}

	b.WriteString("\n")
	b.WriteString(common.HelpStyle.Render("j/k: navigate"))

	return common.AppStyle.Render(b.String())
}
