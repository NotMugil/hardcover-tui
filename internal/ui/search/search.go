package search

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"hardcover-tui/internal/api"
	"hardcover-tui/internal/api/queries"
	"hardcover-tui/internal/common"
)

// NavigateToBookMsg signals the app to navigate to a book's detail view.
type NavigateToBookMsg struct {
	BookID int
	Genres []api.TagItem
}

type searchResultsMsg struct {
	books []api.Book
	err   error
}

// Model is the search screen model.
type Model struct {
	client       *api.Client
	user         *api.User
	textInput    textinput.Model
	results      []api.Book
	table        table.Model
	spinner      spinner.Model
	searching    bool
	inputFocused bool
	err          error
	width        int
	height       int
}

// New creates a new search screen.
func New(client *api.Client, user *api.User) *Model {
	ti := textinput.New()
	ti.Placeholder = "Search books..."
	ti.Width = 50
	ti.Cursor.Style = common.CursorStyle

	s := spinner.New(
		spinner.WithSpinner(spinner.Dot),
		spinner.WithStyle(common.SpinnerStyle),
	)

	t := newSearchTable(80, 15)

	return &Model{
		client:    client,
		user:      user,
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
	return nil
}

// Loaded returns true immediately — search loads on demand.
func (m *Model) Loaded() bool {
	return true
}

// InputFocused returns true when the screen is handling its own key input
// (search input focused, or navigating results table where esc refocuses input).
func (m *Model) InputFocused() bool {
	return m.inputFocused || len(m.results) > 0
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case searchResultsMsg:
		m.searching = false
		if msg.err != nil {
			m.err = msg.err
			return m, nil
		}
		m.results = msg.books
		m.table.SetRows(booksToRows(m.results))
		m.inputFocused = false
		m.textInput.Blur()
		m.table.Focus()
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
				if query == "" {
					return m, nil
				}
				m.searching = true
				m.err = nil
				return m, tea.Batch(m.spinner.Tick, m.doSearch(query))
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

		switch msg.String() {
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
		case "esc", "/":
			m.inputFocused = true
			m.textInput.Focus()
			m.table.Blur()
			return m, textinput.Blink
		}

		var cmd tea.Cmd
		m.table, cmd = m.table.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m *Model) doSearch(query string) tea.Cmd {
	client := m.client
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		books, err := queries.Search(ctx, client, query)
		return searchResultsMsg{books: books, err: err}
	}
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

	if m.searching {
		b.WriteString(fmt.Sprintf("  %s Searching...\n", m.spinner.View()))
		return common.AppStyle.Render(b.String())
	}

	if m.err != nil {
		b.WriteString(common.ErrorStyle.Render("Error: " + m.err.Error()))
		b.WriteString("\n\n")
	}

	if len(m.results) > 0 {
		panelW := m.width - 2
		if panelW < 40 {
			panelW = 80
		}
		tableW := panelW - 4
		if tableW < 40 {
			tableW = 40
		}
		m.table.SetColumns(tableColumns(tableW))
		m.table.SetWidth(tableW)
		tableView := lipgloss.NewStyle().Width(tableW).Render(m.table.View())
		b.WriteString(common.RenderPanel(
			fmt.Sprintf("Results (%d)", len(m.results)),
			tableView, panelW))
	} else {
		panelW := m.width - 2
		if panelW < 40 {
			panelW = 80
		}
		tableW := panelW - 4
		if tableW < 40 {
			tableW = 40
		}
		m.table.SetColumns(tableColumns(tableW))
		m.table.SetWidth(tableW)
		m.table.SetRows([]table.Row{})
		tableView := lipgloss.NewStyle().Width(tableW).Render(m.table.View())
		b.WriteString(common.RenderPanel("Results", tableView, panelW))
	}

	return common.AppStyle.Render(b.String())
}

// HelpBindings returns page-specific keybindings for the global help bar.
func (m *Model) HelpBindings() []key.Binding {
	if m.inputFocused {
		return []key.Binding{
			key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "search")),
		}
	}
	bindings := []key.Binding{
		key.NewBinding(key.WithKeys("/"), key.WithHelp("/", "focus input")),
	}
	if len(m.results) > 0 {
		bindings = append(bindings,
			key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "open book")),
		)
	}
	return bindings
}
