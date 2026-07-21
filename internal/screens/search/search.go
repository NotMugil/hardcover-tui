package search

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/NotMugil/hardcover-tui/internal/api"
	"github.com/NotMugil/hardcover-tui/internal/commands"
	"github.com/NotMugil/hardcover-tui/internal/common"
)

// NavigateToBookMsg signals the app to navigate to a book's detail view.
type NavigateToBookMsg struct {
	BookID int
	Genres []api.TagItem
}

// Model is the search screen model.
type Model struct {
	deps         commands.Deps
	textInput    textinput.Model
	results      []api.Book
	table        table.Model
	spinner      spinner.Model
	searching    bool
	searched     bool
	activeQuery  string
	inputFocused bool
	tableFocused bool
	err          error
	width        int
	height       int
}

// New creates a new search screen.
func New(deps commands.Deps) *Model {
	ti := textinput.New()
	ti.Placeholder = "Search books by title, author, or genre..."
	ti.Width = 50
	ti.Cursor.Style = common.CursorStyle

	s := spinner.New(
		spinner.WithSpinner(spinner.Dot),
		spinner.WithStyle(common.SpinnerStyle),
	)

	t := newSearchTable(80, 15)

	return &Model{
		deps:      deps,
		textInput: ti,
		table:     t,
		spinner:   s,
	}
}

// newSearchTable creates a styled table for search results.
func newSearchTable(width, height int) table.Model {
	columns := tableColumns(width)

	t := table.New(
		table.WithColumns(columns),
		table.WithRows([]table.Row{}),
		table.WithFocused(true),
		table.WithHeight(height),
	)

	st := table.DefaultStyles()
	st.Header = st.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(common.ColorBorder).
		BorderBottom(true).
		Bold(true).
		Foreground(common.ColorPrimary)
	st.Selected = st.Selected.
		Foreground(common.ColorText).
		Background(common.ColorHighlight).
		Bold(false)
	st.Cell = st.Cell.
		Foreground(common.ColorSubtext)
	t.SetStyles(st)

	return t
}

// tableColumns returns table column definitions scaled to the given width.
func tableColumns(width int) []table.Column {
	usable := width - 12
	if usable < 40 {
		usable = 40
	}
	titleW := usable * 30 / 100
	authorW := usable * 25 / 100
	formatW := usable * 10 / 100
	ratingW := usable * 10 / 100
	pagesW := usable * 10 / 100
	usersW := usable - titleW - authorW - formatW - ratingW - pagesW

	return []table.Column{
		{Title: "Title", Width: titleW},
		{Title: "Author", Width: authorW},
		{Title: "Format", Width: formatW},
		{Title: "Rating", Width: ratingW},
		{Title: "Pages", Width: pagesW},
		{Title: "Readers", Width: usersW},
	}
}

// SetSize updates the available terminal dimensions.
func (m *Model) SetSize(w, h int) {
	m.width = w
	m.height = h
	contentW := w - 4
	contentH := h - 10
	if contentW > 0 && contentH > 0 {
		m.table.SetColumns(tableColumns(contentW))
		m.table.SetHeight(contentH)
	}
}

func (m *Model) Init() tea.Cmd {
	m.searching = true
	return tea.Batch(m.spinner.Tick, commands.SearchBooks(m.deps, ""))
}

// Loaded returns true immediately — search loads on demand.
func (m *Model) Loaded() bool {
	return true
}

// InputFocused returns true when the screen is handling its own key input
// (search input focused, or actively navigating the results table).
func (m *Model) InputFocused() bool {
	return m.inputFocused || m.tableFocused
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case commands.SearchResultsMsg:
		m.searching = false
		m.searched = true
		if msg.Err != nil {
			m.err = msg.Err
			return m, nil
		}
		m.results = msg.Books
		m.table.SetRows(booksToRows(m.results))
		if len(m.results) > 0 {
			m.inputFocused = false
			m.tableFocused = true
			m.textInput.Blur()
			m.table.Focus()
		}
		return m, nil

	case spinner.TickMsg:
		if m.searching {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}

	case tea.KeyMsg:
		if m.inputFocused {
			switch msg.String() {
			case "enter":
				query := strings.TrimSpace(m.textInput.Value())
				m.activeQuery = query
				m.searching = true
				m.err = nil
				return m, tea.Batch(m.spinner.Tick, commands.SearchBooks(m.deps, query))
			case "ctrl+l":
				m.textInput.SetValue("")
				m.activeQuery = ""
				m.searching = true
				m.err = nil
				return m, tea.Batch(m.spinner.Tick, commands.SearchBooks(m.deps, ""))
			case "esc":
				m.inputFocused = false
				m.textInput.Blur()
				if len(m.results) > 0 {
					m.table.Focus()
				}
				return m, nil
			}
			var cmd tea.Cmd
			m.textInput, cmd = m.textInput.Update(msg)
			return m, cmd
		}

		switch strings.ToLower(msg.String()) {
		case "enter":
			row := m.table.SelectedRow()
			if row != nil {
				idx := m.table.Cursor()
				if idx >= 0 && idx < len(m.results) {
					book := m.results[idx]
					return m, func() tea.Msg {
						return NavigateToBookMsg{BookID: book.ID, Genres: book.Genres}
					}
				}
			}
		case "/":
			m.inputFocused = true
			m.tableFocused = false
			m.textInput.Focus()
			m.table.Blur()
			return m, textinput.Blink
		case "ctrl+l":
			m.textInput.SetValue("")
			m.activeQuery = ""
			m.searching = true
			m.err = nil
			return m, tea.Batch(m.spinner.Tick, commands.SearchBooks(m.deps, ""))
		case "esc":
			m.tableFocused = false
			m.table.Blur()
			return m, nil
		}

		var cmd tea.Cmd
		m.table, cmd = m.table.Update(msg)
		return m, cmd
	}

	return m, nil
}

// booksToRows converts books to table rows.
func booksToRows(books []api.Book) []table.Row {
	rows := make([]table.Row, len(books))
	for i, b := range books {
		author := b.Authors()
		format := b.FormatIndicator()
		rating := "-"
		if b.Rating != nil {
			rating = fmt.Sprintf("%.1f", *b.Rating)
		}
		pages := "-"
		if b.Pages != nil {
			pages = fmt.Sprintf("%d", *b.Pages)
		}
		readers := fmt.Sprintf("%d", b.UsersCount)

		rows[i] = table.Row{b.Title, author, format, rating, pages, readers}
	}
	return rows
}

func (m *Model) View() string {
	var b strings.Builder

	if m.inputFocused {
		b.WriteString(common.FocusedBorderStyle.Render(m.textInput.View()))
	} else {
		b.WriteString(common.BlurredBorderStyle.Render(m.textInput.View()))
	}
	b.WriteString("\n\n")

	panelW := m.width - 2
	if panelW < 10 {
		panelW = 10
	}

	if m.searching {
		b.WriteString(common.RenderPanel("Search", fmt.Sprintf("  %s Searching...\n", m.spinner.View()), panelW))
		return common.AppStyle.Render(b.String())
	}

	if m.err != nil {
		b.WriteString(common.ErrorStyle.Render("Error: " + m.err.Error()))
		b.WriteString("\n\n")
	}

	panelTitle := "Popular Books"
	if m.activeQuery != "" {
		panelTitle = fmt.Sprintf("Results for \"%s\"", m.activeQuery)
	}

	if len(m.results) > 0 {
		panelTitle = fmt.Sprintf("%s (%d)", panelTitle, len(m.results))
		tableW := panelW - 4
		if tableW < 40 {
			tableW = 40
		}
		m.table.SetColumns(tableColumns(tableW))
		m.table.SetWidth(tableW)
		tableView := lipgloss.NewStyle().Width(tableW).Render(m.table.View())
		b.WriteString(common.RenderPanel(panelTitle, tableView, panelW))
	} else if m.searched {
		emptyMsg := "No books found."
		if m.activeQuery != "" {
			emptyMsg = fmt.Sprintf("No books found matching \"%s\". Try another search term.", m.activeQuery)
		} else {
			emptyMsg = "No books found. Press '/' to type a search query."
		}
		emptyView := lipgloss.NewStyle().Padding(1, 2).Render(common.ValueStyle.Render(emptyMsg))
		b.WriteString(common.RenderPanel(panelTitle, emptyView, panelW))
	} else {
		tableW := panelW - 4
		if tableW < 40 {
			tableW = 40
		}
		m.table.SetColumns(tableColumns(tableW))
		m.table.SetWidth(tableW)
		m.table.SetRows([]table.Row{})
		tableView := lipgloss.NewStyle().Width(tableW).Render(m.table.View())
		b.WriteString(common.RenderPanel(panelTitle, tableView, panelW))
	}

	return common.AppStyle.Render(b.String())
}

// HelpBindings returns page-specific keybindings for the global help bar.
func (m *Model) HelpBindings() []key.Binding {
	if m.inputFocused {
		return []key.Binding{
			key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "search")),
			key.NewBinding(key.WithKeys("ctrl+l"), key.WithHelp("ctrl+l", "clear search")),
		}
	}
	bindings := []key.Binding{
		key.NewBinding(key.WithKeys("/"), key.WithHelp("/", "focus input")),
		key.NewBinding(key.WithKeys("ctrl+l"), key.WithHelp("ctrl+l", "clear search")),
	}
	if len(m.results) > 0 {
		bindings = append(bindings,
			key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "open book")),
		)
	}
	return bindings
}
