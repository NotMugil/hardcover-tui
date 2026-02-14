package home

import (
	"time"

	"github.com/76creates/stickers/flexbox"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/NotMugil/hardcover-tui/internal/api"
	"github.com/NotMugil/hardcover-tui/internal/common"
)

// NavigateToBookMsg signals the app to navigate to a book's detail view.
type NavigateToBookMsg struct {
	UserBook *api.UserBook
}

type booksLoadedMsg struct {
	books     []api.UserBook
	reading   []api.UserBook
	avatarArt string
	err       error
}

type clockTickMsg time.Time

// filterSettledMsg fires after a short delay to trigger the actual data load.
type filterSettledMsg struct {
	filter int
}

type activitiesLoadedMsg struct {
	activities []api.Activity
	err        error
}

// activityFilter selects whose activities to display.
type activityFilter int

const (
	activityFilterMe        activityFilter = iota // user's own activity
	activityFilterFollowing                       // following users' activity
)

// bookItem implements list.DefaultItem for the bubbles list.
type bookItem struct {
	userBook api.UserBook
	filterID int // 0 = all, 1-6 = specific status filter
}

func (i bookItem) Title() string {
	title := i.userBook.Book.Title
	if i.filterID == 0 {
		statusColor := common.StatusColor(i.userBook.StatusID)
		sq := lipgloss.NewStyle().Foreground(statusColor).Bold(true).Render("■")
		title = sq + " " + title
	}
	return title
}

func (i bookItem) Description() string {
	desc := "by " + i.userBook.Book.Authors()
	if i.userBook.Rating != nil {
		desc += " " + common.RenderRatingBar(*i.userBook.Rating, 10)
	}
	return desc
}

func (i bookItem) FilterValue() string {
	return i.userBook.Book.Title
}

type booksOnlyLoadedMsg struct {
	books []api.UserBook
	err   error
}

// Model is the library screen model.
type Model struct {
	client         *api.Client
	user           *api.User
	books          []api.UserBook
	reading        []api.UserBook
	avatarArt      string
	list           list.Model
	readingFocused bool
	readingCursor  int // cursor for currently reading items
	readingScroll  int // scroll offset for currently reading pagination
	progress       progress.Model
	filter         int  // 0 = all, 1-6 = status filter
	filterPending  bool // true while waiting for filter debounce
	spinner        spinner.Model
	loading        bool
	booksLoading   bool // only books are loading (filter change)
	err            error
	page           int
	pageSize       int
	width          int
	height         int
	initialized    bool // profile/reading loaded once
	currentTime    time.Time
	flexBox        *flexbox.FlexBox
	activities      []api.Activity
	activityFilter  activityFilter
	activityLoading bool
	activityFocused bool
	activityCursor  int
	activityScroll  int
	activityErr     error
	confirm    common.ConfirmState
	confirmURL string
}

// New creates a new library screen.
func New(client *api.Client, user *api.User) *Model {
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
	l.SetFilteringEnabled(true)
	l.DisableQuitKeybindings()
	l.Styles.NoItems = common.ValueStyle

	p := progress.New(
		progress.WithSolidFill(string(common.ColorCurrentlyReading)),
		progress.WithoutPercentage(),
	)

	fb := flexbox.New(0, 0)
	row := fb.NewRow().AddCells(
		flexbox.NewCell(6, 1),
		flexbox.NewCell(9, 1),
		flexbox.NewCell(5, 1),
	)
	fb.AddRows([]*flexbox.Row{row})

	return &Model{
		client:   client,
		user:     user,
		list:     l,
		progress: p,
		spinner:  s,
		loading:  true,
		pageSize: 50,
		flexBox:  fb,
	}
}

// SetSize updates the available terminal dimensions.
func (m *Model) SetSize(w, h int) {
	m.width = w
	m.height = h
}

// Loaded returns true once the initial data has been fetched.
func (m *Model) Loaded() bool {
	return !m.loading
}

// InputFocused returns true when the list is in filter mode or activity is focused.
func (m *Model) InputFocused() bool {
	return m.list.FilterState() == list.Filtering || m.activityFocused || m.confirm.Active
}

func (m *Model) Init() tea.Cmd {
	m.currentTime = time.Now()
	return tea.Batch(m.spinner.Tick, m.loadInitial(), m.tickClock(), m.loadActivities())
}

var filterNames = []string{
	"All",
	"Want to Read",
	"Reading",
	"Read",
	"Paused",
	"DNF",
	"Ignored",
}

var filterColors = []lipgloss.Color{
	common.ColorText,
	common.ColorWantToRead,
	common.ColorCurrentlyReading,
	common.ColorRead,
	common.ColorPaused,
	common.ColorDNF,
	common.ColorIgnored,
}
