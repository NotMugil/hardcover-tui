package journal

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"

	"hardcover-tui/internal/api"
	"hardcover-tui/internal/api/mutations"
	"hardcover-tui/internal/api/queries"
	"hardcover-tui/internal/common"
)

type journalsLoadedMsg struct {
	journals []api.ReadingJournal
	err      error
}

type journalSavedMsg struct {
	err error
}

type viewMode int

const (
	modeList viewMode = iota
	modeWrite
)

// journalItem implements list.DefaultItem for the bubbles list.
type journalItem struct {
	data api.ReadingJournal
}

func (i journalItem) Title() string {
	date := i.data.ActionAt
	if len(date) > 10 {
		date = date[:10]
	}
	bookTitle := ""
	if i.data.Book != nil {
		bookTitle = " - " + i.data.Book.Title
	}
	return fmt.Sprintf("%s [%s]%s", date, i.data.Event, bookTitle)
}

func (i journalItem) Description() string {
	if i.data.Entry != nil && *i.data.Entry != "" {
		entry := *i.data.Entry
		if len(entry) > 80 {
			entry = entry[:77] + "..."
		}
		return entry
	}
	return ""
}

func (i journalItem) FilterValue() string {
	title := i.data.Event
	if i.data.Book != nil {
		title += " " + i.data.Book.Title
	}
	return title
}

// Model is the journal screen model.
type Model struct {
	client   *api.Client
	user     *api.User
	userBook *api.UserBook
	journals []api.ReadingJournal
	list     list.Model
	textarea textarea.Model
	spinner  spinner.Model
	loading  bool
	err      error
	success  bool
	mode     viewMode
	width    int
	height   int
}

// New creates a new journal screen.
func New(client *api.Client, user *api.User, ub *api.UserBook) *Model {
	ta := textarea.New()
	ta.Placeholder = "Write a journal entry..."
	ta.SetWidth(60)
	ta.SetHeight(6)
	ta.Cursor.Style = common.CursorStyle

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

	l := list.New([]list.Item{}, delegate, 80, 15)
	l.SetShowTitle(false)
	l.SetShowStatusBar(true)
	l.SetShowHelp(false)
	l.SetFilteringEnabled(true)
	l.DisableQuitKeybindings()
	l.Styles.NoItems = common.ValueStyle

	return &Model{
		client:   client,
		user:     user,
		userBook: ub,
		list:     l,
		textarea: ta,
		spinner:  s,
		loading:  true,
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
	return tea.Batch(m.spinner.Tick, m.loadJournals())
}

// InputFocused returns true when writing a new entry or filtering.
func (m *Model) InputFocused() bool {
	return m.mode == modeWrite || m.list.FilterState() == list.Filtering
}

func (m *Model) loadJournals() tea.Cmd {
	client := m.client
	user := m.user
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		journals, err := queries.GetReadingJournals(ctx, client, user.ID, 20)
		return journalsLoadedMsg{journals: journals, err: err}
	}
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case journalsLoadedMsg:
		m.loading = false
		if msg.err != nil {
			m.err = msg.err
			return m, nil
		}
		m.journals = msg.journals
		items := make([]list.Item, len(m.journals))
		for i, j := range m.journals {
			items[i] = journalItem{data: j}
		}
		m.list.SetItems(items)
		return m, nil

	case journalSavedMsg:
		m.loading = false
		m.mode = modeList
		if msg.err != nil {
			m.err = msg.err
			return m, nil
		}
		m.success = true
		m.textarea.SetValue("")
		return m, m.loadJournals()

	case spinner.TickMsg:
		if m.loading {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}

	case tea.KeyMsg:
		if m.loading {
			return m, nil
		}

		if m.mode == modeWrite {
			switch msg.String() {
			case "ctrl+s":
				entry := strings.TrimSpace(m.textarea.Value())
				if entry == "" {
					return m, nil
				}
				m.loading = true
				m.err = nil
				m.success = false
				return m, tea.Batch(m.spinner.Tick, m.saveEntry(entry))
			case "esc":
				m.mode = modeList
				m.textarea.Blur()
				return m, nil
			}
			var cmd tea.Cmd
			m.textarea, cmd = m.textarea.Update(msg)
			return m, cmd
		}

		if m.list.FilterState() == list.Filtering {
			var cmd tea.Cmd
			m.list, cmd = m.list.Update(msg)
			return m, cmd
		}

		switch msg.String() {
		case "n":
			m.mode = modeWrite
			m.success = false
			m.textarea.Focus()
			return m, textarea.Blink
		case "d":
			if item, ok := m.list.SelectedItem().(journalItem); ok {
				m.loading = true
				return m, tea.Batch(m.spinner.Tick, m.deleteEntry(item.data.ID))
			}
		}
	}

	if !m.loading && m.mode == modeList {
		var cmd tea.Cmd
		m.list, cmd = m.list.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m *Model) saveEntry(entry string) tea.Cmd {
	client := m.client
	ub := m.userBook
	return func() tea.Msg {
		if ub == nil {
			return journalSavedMsg{err: fmt.Errorf("no book selected")}
		}
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		now := time.Now().Format("2006-01-02")
		err := mutations.InsertReadingJournal(ctx, client, ub.BookID, "note", entry, now)
		return journalSavedMsg{err: err}
	}
}

func (m *Model) deleteEntry(id int) tea.Cmd {
	client := m.client
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		err := mutations.DeleteReadingJournal(ctx, client, id)
		if err != nil {
			return journalsLoadedMsg{err: err}
		}
		journals, err := queries.GetReadingJournals(ctx, client, m.user.ID, 20)
		return journalsLoadedMsg{journals: journals, err: err}
	}
}

func (m *Model) View() string {
	if m.loading {
		return common.AppStyle.Render(
			fmt.Sprintf("\n  %s Loading journal...\n", m.spinner.View()),
		)
	}

	var b strings.Builder

	b.WriteString(common.TitleStyle.Render("Reading Journal"))
	b.WriteString("\n\n")

	if m.err != nil {
		b.WriteString(common.ErrorStyle.Render("Error: "+m.err.Error()) + "\n\n")
	}

	if m.success {
		b.WriteString(common.SuccessStyle.Render("Entry saved!") + "\n\n")
	}

	if m.mode == modeWrite {
		var write strings.Builder
		write.WriteString(common.LabelStyle.Render("New Journal Entry"))
		if m.userBook != nil {
			write.WriteString(fmt.Sprintf(" - %s", m.userBook.Book.Title))
		}
		write.WriteString("\n\n")
		write.WriteString(m.textarea.View())
		write.WriteString("\n\n")
		write.WriteString(common.HelpStyle.Render("ctrl+s: save | esc: cancel"))
		b.WriteString(common.PanelActiveStyle.Render(write.String()))
		return common.AppStyle.Render(b.String())
	}

	b.WriteString(common.PanelStyle.Render(m.list.View()))

	b.WriteString("\n")
	b.WriteString(common.HelpStyle.Render("n: new entry | d: delete | j/k: navigate | esc: back"))

	return common.AppStyle.Render(b.String())
}
