package bookdetail

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/76creates/stickers/flexbox"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/NotMugil/hardcover-tui/internal/api"
	"github.com/NotMugil/hardcover-tui/internal/common"
)

// Navigation messages for app.go to catch.
type NavigateToReviewMsg struct {
	UserBook *api.UserBook
}

type NavigateToProgressMsg struct {
	UserBook *api.UserBook
}

type NavigateToJournalMsg struct {
	UserBook *api.UserBook
}

type bookLoadedMsg struct {
	userBook *api.UserBook
	err      error
}

// bookFromBookIDMsg is returned when loading via book ID (e.g. from list detail).
type bookFromBookIDMsg struct {
	book     *api.Book
	userBook *api.UserBook // may be nil if user hasn't added this book
	err      error
}

type statusUpdatedMsg struct {
	err error
}

type ratingUpdatedMsg struct {
	err error
}

type bookAddedMsg struct {
	userBook *api.UserBook
	err      error
}

type coverLoadedMsg struct {
	art string
}

type tagsLoadedMsg struct {
	genres []api.TagItem
	err    error
}

type reviewsLoadedMsg struct {
	reviews []api.BookReview
	err     error
}

type userListsLoadedMsg struct {
	lists []api.List
	err   error
}

type bookAddedToListMsg struct {
	listName string
	err      error
}

type bookRemovedFromListMsg struct {
	err error
}

type viewMode int

const (
	modeDetail viewMode = iota
	modeStatusSelect
	modeRatingSelect
	modeJournal
	modeJournalWrite
	modeReviewRead
	modeListSelect
	modeConfirm
)

// Journal messages
type journalsLoadedMsg struct {
	journals []api.ReadingJournal
	err      error
}

type journalSavedMsg struct {
	err error
}

type journalDeletedMsg struct {
	err error
}

// journalItem implements list.DefaultItem for journal entries in the inline view.
type journalItem struct {
	data api.ReadingJournal
}

func (i journalItem) Title() string {
	date := i.data.ActionAt
	if len(date) > 10 {
		date = date[:10]
	}
	return fmt.Sprintf("%s [%s]", date, i.data.Event)
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
	return i.data.Event
}

// stripHTML removes HTML tags and decodes common HTML entities.
var htmlTagRe = regexp.MustCompile(`<[^>]*>`)

func stripHTML(s string) string {
	s = htmlTagRe.ReplaceAllString(s, "")
	r := strings.NewReplacer(
		"&amp;", "&",
		"&lt;", "<",
		"&gt;", ">",
		"&quot;", `"`,
		"&#39;", "'",
		"&nbsp;", " ",
	)
	s = r.Replace(s)
	s = strings.Join(strings.Fields(s), " ")
	return s
}

// reviewItem implements list.DefaultItem for popular reviews.
type reviewItem struct {
	data api.BookReview
}

func (i reviewItem) Title() string {
	header := "@" + i.data.User.Username
	if i.data.Rating != nil {
		header += fmt.Sprintf("  %.1f/5", *i.data.Rating)
	}
	if i.data.LikesCount > 0 {
		header += fmt.Sprintf("  %d likes", i.data.LikesCount)
	}
	return header
}

func (i reviewItem) Description() string {
	if i.data.Review != nil && *i.data.Review != "" {
		text := stripHTML(*i.data.Review)
		if i.data.ReviewHasSpoilers {
			text = "[spoiler] " + text
		}
		if len(text) > 120 {
			text = text[:117] + "..."
		}
		return text
	}
	return "(no text)"
}

func (i reviewItem) FilterValue() string {
	return i.data.User.Username
}

// ListBookEntry holds minimal info for list next/prev navigation.
// Exported so listdetail and app.go can build the slice.
type ListBookEntry struct {
	BookID int
	Title  string
}

// Model is the book detail screen model.
type Model struct {
	client         *api.Client
	user           *api.User
	userBook       *api.UserBook
	book           *api.Book // standalone book (when loaded from book_id without user_book)
	bookID         int
	loadByBook     bool // true if loading by book_id rather than user_book_id
	coverArt       string
	spinner        spinner.Model
	loading        bool
	err            error
	mode           viewMode
	cursor         int // for status/rating selection
	width          int
	height         int
	descExpanded   bool // whether the description panel is fully expanded
	genres         []api.TagItem
	reviews        []api.BookReview
	reviewList     list.Model
	reviewMode     bool // true when browsing reviews list
	reviewViewport viewport.Model
	selectedReview *api.BookReview
	listBooks      []ListBookEntry
	listIndex      int
	listID         int
	listName       string
	journals       []api.ReadingJournal
	journalList    list.Model
	journalTA      textarea.Model
	journalLoading bool
	journalErr     error
	journalSuccess bool
	flexBox        *flexbox.FlexBox
	userLists      []api.List
	listCursor     int
	listLoading    bool
	listSuccess    bool
	listErr        error
	confirm        common.ConfirmState
	confirmItemID  int
	confirmReturn  viewMode // mode to return to if cancelled
}

// NewFromUserBook creates a book detail screen from an existing UserBook.
func NewFromUserBook(client *api.Client, user *api.User, ub *api.UserBook) *Model {
	s := spinner.New(
		spinner.WithSpinner(spinner.Dot),
		spinner.WithStyle(common.SpinnerStyle),
	)
	m := &Model{
		client:     client,
		user:       user,
		userBook:   ub,
		spinner:    s,
		flexBox:    newDetailFlexBox(),
		reviewList: newReviewList(),
	}
	m.initJournal()
	return m
}

// NewFromID creates a book detail screen that loads from a user_book ID.
func NewFromID(client *api.Client, user *api.User, id int) *Model {
	s := spinner.New(
		spinner.WithSpinner(spinner.Dot),
		spinner.WithStyle(common.SpinnerStyle),
	)
	m := &Model{
		client:     client,
		user:       user,
		bookID:     id,
		spinner:    s,
		loading:    true,
		flexBox:    newDetailFlexBox(),
		reviewList: newReviewList(),
	}
	m.initJournal()
	return m
}

// NewFromBookID creates a book detail screen that loads from a book ID.
// This first fetches the book, then tries to find the user's relationship.
func NewFromBookID(client *api.Client, user *api.User, bookID int) *Model {
	s := spinner.New(
		spinner.WithSpinner(spinner.Dot),
		spinner.WithStyle(common.SpinnerStyle),
	)
	m := &Model{
		client:     client,
		user:       user,
		bookID:     bookID,
		loadByBook: true,
		spinner:    s,
		loading:    true,
		flexBox:    newDetailFlexBox(),
		reviewList: newReviewList(),
	}
	m.initJournal()
	return m
}

// NewFromListBook creates a book detail screen from a list context.
// It loads the book by ID and stores the list of books for next/prev navigation.
func NewFromListBook(client *api.Client, user *api.User, bookID int, listBooks []ListBookEntry, listIndex int, listID int, listName string) *Model {
	m := NewFromBookID(client, user, bookID)
	m.listBooks = listBooks
	m.listIndex = listIndex
	m.listID = listID
	m.listName = listName
	return m
}

// Loaded reports whether the book detail screen has finished its initial load.
func (m *Model) Loaded() bool {
	return !m.loading
}

// SetGenres pre-populates the genres from search results so they display
// immediately while the full book details are still loading.
func (m *Model) SetGenres(genres []api.TagItem) {
	m.genres = genres
}

// newDetailFlexBox creates the FlexBox layout: 1 row, 2 cells (35% left, 65% right).
func newDetailFlexBox() *flexbox.FlexBox {
	fb := flexbox.New(0, 0)
	row := fb.NewRow().AddCells(
		flexbox.NewCell(7, 1),  // left ~35%
		flexbox.NewCell(13, 1), // right ~65%
	)
	fb.AddRows([]*flexbox.Row{row})
	return fb
}

func newReviewList() list.Model {
	delegate := list.NewDefaultDelegate()
	delegate.Styles.SelectedTitle = delegate.Styles.SelectedTitle.
		Foreground(common.ColorPrimary).
		BorderLeftForeground(common.ColorPrimary)
	delegate.Styles.SelectedDesc = delegate.Styles.SelectedDesc.
		Foreground(common.ColorSubtext).
		BorderLeftForeground(common.ColorPrimary)

	rl := list.New([]list.Item{}, delegate, 40, 10)
	rl.SetShowTitle(false)
	rl.SetShowStatusBar(true)
	rl.SetShowHelp(false)
	rl.SetFilteringEnabled(false)
	rl.DisableQuitKeybindings()
	rl.Styles.NoItems = common.ValueStyle
	return rl
}

func (m *Model) initJournal() {
	delegate := list.NewDefaultDelegate()
	delegate.Styles.SelectedTitle = delegate.Styles.SelectedTitle.
		Foreground(common.ColorPrimary).
		BorderLeftForeground(common.ColorPrimary)
	delegate.Styles.SelectedDesc = delegate.Styles.SelectedDesc.
		Foreground(common.ColorSubtext).
		BorderLeftForeground(common.ColorPrimary)

	l := list.New([]list.Item{}, delegate, 50, 15)
	l.SetShowTitle(false)
	l.SetShowStatusBar(false)
	l.SetShowHelp(false)
	l.SetFilteringEnabled(false)
	l.DisableQuitKeybindings()
	l.Styles.NoItems = common.ValueStyle

	ta := textarea.New()
	ta.Placeholder = "Write a journal entry..."
	ta.SetWidth(50)
	ta.SetHeight(6)
	ta.Cursor.Style = common.CursorStyle

	m.journalList = l
	m.journalTA = ta
}

func (m *Model) Init() tea.Cmd {
	if m.loading {
		if m.loadByBook {
			return tea.Batch(m.spinner.Tick, m.loadBookByBookID())
		}
		return tea.Batch(m.spinner.Tick, m.loadBook())
	}
	var cmds []tea.Cmd
	if m.userBook != nil {
		if m.userBook.Book.CoverURL() != "" {
			cmds = append(cmds, m.loadCover(m.userBook.Book.CoverURL()))
		}
		cmds = append(cmds, m.loadTags(m.userBook.Book.ID))
		cmds = append(cmds, m.loadReviews(m.userBook.Book.ID))
	}
	if len(cmds) > 0 {
		return tea.Batch(cmds...)
	}
	return nil
}

// SetSize updates the available terminal dimensions.
func (m *Model) SetSize(w, h int) {
	m.width = w
	m.height = h
}

func (m *Model) getWidth() int {
	if m.width > 0 {
		return m.width
	}
	return 80
}

// InputFocused returns true when an overlay or sub-mode is active,
// preventing the global ESC/Back handler from popping the navstack.
func (m *Model) InputFocused() bool {
	return m.mode != modeDetail || m.reviewMode
}

// switchToListBook resets the model state for loading a new book from the list.
func (m *Model) switchToListBook(idx int) {
	entry := m.listBooks[idx]
	m.bookID = entry.BookID
	m.loadByBook = true
	m.loading = true
	m.book = nil
	m.userBook = nil
	m.coverArt = ""
	m.genres = nil

	m.reviews = nil
	m.reviewList.SetItems(nil)
	m.reviewMode = false
	m.descExpanded = false
	m.mode = modeDetail
	m.err = nil
}
