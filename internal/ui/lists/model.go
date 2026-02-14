package lists

import (
	"fmt"

	"github.com/76creates/stickers/flexbox"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"hardcover-tui/internal/api"
	"hardcover-tui/internal/common"
)

// NavigateToBookFromListMsg signals the app to navigate to a book's detail from a list.
type NavigateToBookFromListMsg struct {
	BookID    int
	ListBooks []ListBookEntry
	ListIndex int
	ListID    int
	ListName  string
}

// ListBookEntry holds minimal info for list book navigation context.
type ListBookEntry struct {
	BookID int
	Title  string
}

// listSettledMsg fires after a debounce delay to trigger loading list books.
type listSettledMsg struct {
	listID int
}

type listsLoadedMsg struct {
	lists []api.List
	err   error
}

type listBooksLoadedMsg struct {
	books []api.ListBook
	err   error
}

type listCreatedMsg struct {
	list *api.List
	err  error
}

type listDeletedMsg struct {
	err error
}

type searchBooksMsg struct {
	books []api.Book
	err   error
}

type bookAddedToListMsg struct {
	err error
}

type bookRemovedFromListMsg struct {
	err error
}

type privacyUpdatedMsg struct {
	err error
}

type inputMode int

const (
	modeNormal inputMode = iota
	modeCreate
	modeAddBook
	modePrivacy
	modeConfirm
)

// listItem implements list.DefaultItem for the bubbles list.
type listItem struct {
	data api.List
}

func (i listItem) Title() string {
	var badgeText string
	var badgeBg lipgloss.Color
	switch api.PrivacySettingID(i.data.PrivacySettingID) {
	case api.PrivacyPublic:
		badgeText = "public"
		badgeBg = common.ColorSuccess
	case api.PrivacyFollowers:
		badgeText = "followers"
		badgeBg = common.ColorSecondary
	case api.PrivacyPrivate:
		badgeText = "private"
		badgeBg = common.ColorMuted
	default:
		badgeText = "private"
		badgeBg = common.ColorMuted
	}
	vis := lipgloss.NewStyle().
		Foreground(common.ColorBackground).
		Background(badgeBg).
		Bold(true).
		Padding(0, 1).
		Render(badgeText)
	return fmt.Sprintf("%s %s", vis, i.data.Name)
}

func (i listItem) Description() string {
	return fmt.Sprintf("%d books", i.data.BooksCount)
}

func (i listItem) FilterValue() string {
	return i.data.Name
}

// bookListItem implements list.DefaultItem for books in a list.
type bookListItem struct {
	data api.ListBook
}

func (i bookListItem) Title() string {
	return i.data.Book.Title
}

func (i bookListItem) Description() string {
	desc := "by " + i.data.Book.Authors()
	if i.data.Book.Rating != nil {
		desc += "  " + common.RenderRatingBar(*i.data.Book.Rating, 10)
	}
	return desc
}

func (i bookListItem) FilterValue() string {
	return i.data.Book.Title
}

// Model is the lists screen model.
type Model struct {
	client        *api.Client
	user          *api.User
	lists         []api.List
	listBooks     []api.ListBook
	selectedIdx   int
	list          list.Model
	bookList      list.Model
	focusRight    bool
	spinner       spinner.Model
	loading       bool
	booksLoading  bool
	err           error
	mode          inputMode
	nameInput     textinput.Model
	width         int
	height        int
	flexBox       *flexbox.FlexBox
	pendingListID int // debounce: ID of list waiting to load
	searchInput   textinput.Model
	searchResults []api.Book
	searchTable   table.Model
	searching     bool
	addSuccess    bool
	addErr        error
	privacyCursor int
	confirm       common.ConfirmState
	confirmItemID int // ID of item being confirmed for delete/remove
}

// New creates a new lists screen.
func New(client *api.Client, user *api.User) *Model {
	s := spinner.New(
		spinner.WithSpinner(spinner.Dot),
		spinner.WithStyle(common.SpinnerStyle),
	)

	ti := textinput.New()
	ti.Placeholder = "List name..."
	ti.Width = 40
	ti.Cursor.Style = common.CursorStyle

	delegate := list.NewDefaultDelegate()
	delegate.Styles.SelectedTitle = delegate.Styles.SelectedTitle.
		Foreground(common.ColorPrimary).
		BorderLeftForeground(common.ColorPrimary)
	delegate.Styles.SelectedDesc = delegate.Styles.SelectedDesc.
		Foreground(common.ColorSubtext).
		BorderLeftForeground(common.ColorPrimary)

	l := list.New([]list.Item{}, delegate, 30, 20)
	l.SetShowTitle(false)
	l.SetShowStatusBar(false)
	l.SetShowHelp(false)
	l.SetFilteringEnabled(true)
	l.DisableQuitKeybindings()
	l.Styles.NoItems = common.ValueStyle

	bookDelegate := list.NewDefaultDelegate()
	bookDelegate.Styles.SelectedTitle = bookDelegate.Styles.SelectedTitle.
		Foreground(common.ColorPrimary).
		BorderLeftForeground(common.ColorPrimary)
	bookDelegate.Styles.SelectedDesc = bookDelegate.Styles.SelectedDesc.
		Foreground(common.ColorSubtext).
		BorderLeftForeground(common.ColorPrimary)
	bl := list.New([]list.Item{}, bookDelegate, 50, 20)
	bl.SetShowTitle(false)
	bl.SetShowStatusBar(false)
	bl.SetShowHelp(false)
	bl.SetFilteringEnabled(false)
	bl.DisableQuitKeybindings()
	bl.Styles.NoItems = common.ValueStyle

	fb := flexbox.New(0, 0)
	row := fb.NewRow().AddCells(
		flexbox.NewCell(3, 1),
		flexbox.NewCell(8, 1),
	)
	fb.AddRows([]*flexbox.Row{row})

	si := textinput.New()
	si.Placeholder = "Search books to add..."
	si.Width = 40
	si.Cursor.Style = common.CursorStyle

	st := newSearchResultTable(50, 10)

	return &Model{
		client:      client,
		user:        user,
		list:        l,
		bookList:    bl,
		spinner:     s,
		loading:     true,
		nameInput:   ti,
		flexBox:     fb,
		searchInput: si,
		searchTable: st,
	}
}

// SetSize updates the available terminal dimensions.
func (m *Model) SetSize(w, h int) {
	m.width = w
	m.height = h
}

// Loaded returns true once the initial list data has been fetched.
func (m *Model) Loaded() bool {
	return !m.loading
}

func (m *Model) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, m.loadLists())
}

// InputFocused returns true when text input is active.
func (m *Model) InputFocused() bool {
	return m.mode == modeCreate || m.mode == modeAddBook || m.mode == modePrivacy || m.mode == modeConfirm || m.list.FilterState() == list.Filtering
}

// newSearchResultTable creates a styled table for search results in the add-book flow.
func newSearchResultTable(width, height int) table.Model {
	cols := searchTableColumns(width)
	t := table.New(
		table.WithColumns(cols),
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

func searchTableColumns(width int) []table.Column {
	usable := width - 4
	if usable < 30 {
		usable = 30
	}
	titleW := usable * 50 / 100
	authorW := usable * 30 / 100
	ratingW := usable - titleW - authorW
	return []table.Column{
		{Title: "Title", Width: titleW},
		{Title: "Author", Width: authorW},
		{Title: "Rating", Width: ratingW},
	}
}
