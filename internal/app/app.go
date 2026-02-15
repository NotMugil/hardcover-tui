package app

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/kevm/bubbleo/navstack"
	"github.com/kevm/bubbleo/window"
	zone "github.com/lrstanley/bubblezone"
	overlay "github.com/rmhubbert/bubbletea-overlay"
	"go.dalton.dog/bubbleup"

	"github.com/NotMugil/hardcover-tui/internal/api"
	"github.com/NotMugil/hardcover-tui/internal/api/queries"
	"github.com/NotMugil/hardcover-tui/internal/common"
	"github.com/NotMugil/hardcover-tui/internal/keystore"
	"github.com/NotMugil/hardcover-tui/internal/ui/bookdetail"
	"github.com/NotMugil/hardcover-tui/internal/ui/home"
	"github.com/NotMugil/hardcover-tui/internal/ui/journal"
	"github.com/NotMugil/hardcover-tui/internal/ui/lists"
	"github.com/NotMugil/hardcover-tui/internal/ui/progress"
	"github.com/NotMugil/hardcover-tui/internal/ui/review"
	"github.com/NotMugil/hardcover-tui/internal/ui/search"
	"github.com/NotMugil/hardcover-tui/internal/ui/setup"
	"github.com/NotMugil/hardcover-tui/internal/ui/stats"
)

// Screen is an interface that all screens implement.
type Screen interface {
	Init() tea.Cmd
	Update(tea.Msg) (tea.Model, tea.Cmd)
	View() string
}

// inputFocusable is implemented by screens that have text inputs.
type inputFocusable interface {
	InputFocused() bool
}

// sizable is implemented by screens that can adapt to terminal dimensions.
type sizable interface {
	SetSize(w, h int)
}

// loadable is implemented by screens that report when initial data is ready.
type loadable interface {
	Loaded() bool
}

// navTab defines a navigation tab.
type navTab struct {
	name   string
	zoneID string
}

var navTabs = []navTab{
	{"Home", "nav-home"},
	{"Search", "nav-search"},
	{"Lists", "nav-lists"},
	{"Stats", "nav-stats"},
}

// Model is the root application model.
type Model struct {
	nav        *navstack.Model
	win        *window.Model
	client     *api.Client
	user       *api.User
	spinner    spinner.Model
	keys       common.KeyMap
	help       help.Model
	loading    bool
	setupMode  bool
	setupScr   Screen
	err        error
	width      int
	height     int
	activeTab  int
	confirm    common.ConfirmState
	alert      bubbleup.AlertModel
	loader     common.Loader
	tabLoading bool
}

// keyringCheckMsg is returned after checking the keyring for an API key.
type keyringCheckMsg struct {
	apiKey string
	err    error
}

// userLoadedMsg is returned after fetching the user profile.
type userLoadedMsg struct {
	user *api.User
	err  error
}

// New creates the root application model.
func New() Model {
	s := spinner.New(
		spinner.WithSpinner(spinner.Dot),
		spinner.WithStyle(common.SpinnerStyle),
	)
	w := window.New(120, 30, 0, 0)
	n := navstack.New(&w)

	alertModel := bubbleup.NewAlertModel(50, false, 3*time.Second).
		WithMinWidth(20).
		WithPosition(bubbleup.TopRightPosition).
		WithUnicodePrefix()

	alertModel.RegisterNewAlertType(bubbleup.AlertDefinition{
		Key:       string(common.NotifySuccess),
		ForeColor: "#10B981", // ColorSuccess
		Prefix:    "\u2714",  // checkmark
	})

	return Model{
		nav:       &n,
		win:       &w,
		spinner:   s,
		keys:      common.Keys,
		help:      common.NewHelp(),
		loading:   true,
		setupMode: true,
		alert:     alertModel,
		loader:    common.NewLoader(),
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, checkKeyringCmd())
}

func checkKeyringCmd() tea.Cmd {
	return func() tea.Msg {
		apiKey, err := keystore.Load()
		return keyringCheckMsg{apiKey: apiKey, err: err}
	}
}

func (m Model) loadUser() tea.Cmd {
	client := m.client
	return func() tea.Msg {
		ctx, cancel := makeContext()
		defer cancel()
		user, err := queries.GetMe(ctx, client)
		return userLoadedMsg{user: user, err: err}
	}
}

// contentHeight returns the available height for screen content.
func (m Model) contentHeight() int {
	overhead := 5
	if m.help.ShowAll {
		overhead++
	}
	return m.height - overhead
}

// pushScreen pushes a new screen onto the navstack.
func (m Model) pushScreen(title string, screen Screen) (Model, tea.Cmd) {
	if s, ok := screen.(sizable); ok && m.width > 0 {
		s.SetSize(m.width, m.contentHeight())
	}
	item := navstack.NavigationItem{Title: title, Model: screen}
	cmd := m.nav.Push(item)
	return m, cmd
}

// switchTab clears the navstack and pushes the given tab screen.
func (m Model) switchTab(idx int) (Model, tea.Cmd) {
	if idx == m.activeTab && len(m.nav.StackSummary()) == 1 {
		cmd := m.nav.Update(navstack.ReloadCurrent{})
		if top := m.nav.Top(); top != nil {
			if s, ok := top.Model.(sizable); ok && m.width > 0 {
				s.SetSize(m.width, m.contentHeight())
			}
		}
		m.tabLoading = true
		loaderCmd := m.loader.Start()
		return m, tea.Batch(cmd, loaderCmd)
	}
	m.activeTab = idx
	_ = m.nav.Clear()
	screen := m.createTabScreen(idx)

	if l, ok := screen.(loadable); ok && l.Loaded() {
		return m.pushScreen(navTabs[idx].name, screen)
	}

	m.tabLoading = true
	loaderCmd := m.loader.Start()
	nm, pushCmd := m.pushScreen(navTabs[idx].name, screen)
	return nm, tea.Batch(pushCmd, loaderCmd)
}

// createTabScreen instantiates a screen for the given tab index.
func (m Model) createTabScreen(idx int) Screen {
	switch idx {
	case 0:
		return home.New(m.client, m.user)
	case 1:
		return search.New(m.client, m.user)
	case 2:
		return lists.New(m.client, m.user)
	case 3:
		return stats.New(m.client, m.user)
	default:
		return home.New(m.client, m.user)
	}
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var alertCmd tea.Cmd
	outAlert, alertTickCmd := m.alert.Update(msg)
	m.alert = outAlert.(bubbleup.AlertModel)
	alertCmd = alertTickCmd

	switch msg := msg.(type) {
	case common.NotifyMsg:
		alertKey := string(msg.Level)
		newAlertCmd := m.alert.NewAlertCmd(alertKey, msg.Message)
		return m, tea.Batch(alertCmd, newAlertCmd)

	case common.LoaderFrameMsg:
		if m.tabLoading {
			if top := m.nav.Top(); top != nil {
				if l, ok := top.Model.(loadable); ok && l.Loaded() {
					m.tabLoading = false
					m.loader.Stop()
					return m, alertCmd
				}
			}
			frameCmd := m.loader.Update()
			return m, tea.Batch(alertCmd, frameCmd)
		}
		return m, alertCmd

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.win.Width = msg.Width
		m.win.Height = msg.Height
		m.help.Width = msg.Width
		if m.setupMode && m.setupScr != nil {
			if s, ok := m.setupScr.(sizable); ok {
				s.SetSize(msg.Width, msg.Height)
			}
			return m, nil
		}
		if top := m.nav.Top(); top != nil {
			if s, ok := top.Model.(sizable); ok {
				s.SetSize(msg.Width, m.contentHeight())
			}
		}
		return m, nil

	case keyringCheckMsg:
		if msg.err != nil || msg.apiKey == "" {
			m.loading = false
			m.setupMode = true
			s := setup.New()
			m.setupScr = s
			return m, s.Init()
		}
		m.client = api.NewClient(msg.apiKey)
		return m, m.loadUser()

	case userLoadedMsg:
		m.loading = false
		if msg.err != nil {
			m.err = msg.err
			m.setupMode = true
			s := setup.New()
			m.setupScr = s
			return m, s.Init()
		}
		m.user = msg.user
		m.setupMode = false
		m.setupScr = nil
		m.activeTab = 0
		screen := home.New(m.client, m.user)
		m.tabLoading = true
		loaderCmd := m.loader.Start()
		nm, pushCmd := m.pushScreen("Home", screen)
		return nm, tea.Batch(pushCmd, loaderCmd)

	case setup.SetupCompleteMsg:
		m.client = api.NewClient(msg.Token)
		m.loading = true
		return m, tea.Batch(m.spinner.Tick, m.loadUser())

	case home.NavigateToBookMsg:
		title := "Book"
		if msg.UserBook != nil {
			title = msg.UserBook.Book.Title
		}
		screen := bookdetail.NewFromUserBook(m.client, m.user, msg.UserBook)
		nm, pushCmd := m.pushScreen(title, screen)
		if !screen.Loaded() {
			nm.tabLoading = true
			loaderCmd := nm.loader.Start()
			return nm, tea.Batch(pushCmd, loaderCmd)
		}
		return nm, pushCmd

	case search.NavigateToBookMsg:
		screen := bookdetail.NewFromBookID(m.client, m.user, msg.BookID)
		if len(msg.Genres) > 0 {
			screen.SetGenres(msg.Genres)
		}
		nm, pushCmd := m.pushScreen("Book", screen)
		if !screen.Loaded() {
			nm.tabLoading = true
			loaderCmd := nm.loader.Start()
			return nm, tea.Batch(pushCmd, loaderCmd)
		}
		return nm, pushCmd

	case lists.NavigateToBookFromListMsg:
		entries := make([]bookdetail.ListBookEntry, len(msg.ListBooks))
		for i, lb := range msg.ListBooks {
			entries[i] = bookdetail.ListBookEntry{BookID: lb.BookID, Title: lb.Title}
		}
		screen := bookdetail.NewFromListBook(m.client, m.user, msg.BookID, entries, msg.ListIndex, msg.ListID, msg.ListName)
		nm, pushCmd := m.pushScreen("Book", screen)
		if !screen.Loaded() {
			nm.tabLoading = true
			loaderCmd := nm.loader.Start()
			return nm, tea.Batch(pushCmd, loaderCmd)
		}
		return nm, pushCmd

	case bookdetail.NavigateToReviewMsg:
		screen := review.New(m.client, m.user, msg.UserBook)
		return m.pushScreen("Review", screen)

	case bookdetail.NavigateToProgressMsg:
		screen := progress.New(m.client, m.user, msg.UserBook)
		return m.pushScreen("Progress", screen)

	case progress.NavigateBackMsg:
		if len(m.nav.StackSummary()) > 1 {
			cmd := m.nav.Pop()
			if top := m.nav.Top(); top != nil {
				if s, ok := top.Model.(sizable); ok {
					s.SetSize(m.width, m.contentHeight())
				}
			}
			return m, cmd
		}

	case bookdetail.NavigateToJournalMsg:
		screen := journal.New(m.client, m.user, msg.UserBook)
		return m.pushScreen("Journal", screen)

	case spinner.TickMsg:
		if m.loading {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}
		if !m.setupMode {
			cmd := m.nav.Update(msg)
			return m, cmd
		}
		return m, nil

	case tea.MouseMsg:
		if !m.setupMode && !m.loading &&
			msg.Action == tea.MouseActionPress && msg.Button == tea.MouseButtonLeft {
			for i, t := range navTabs {
				if zone.Get(t.zoneID).InBounds(msg) {
					return m.switchTab(i)
				}
			}
		}
		if !m.setupMode && !m.loading {
			cmd := m.nav.Update(msg)
			return m, cmd
		}
		return m, nil

	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}

		if m.setupMode && m.setupScr != nil {
			updated, cmd := m.setupScr.Update(msg)
			m.setupScr = updated.(Screen)
			return m, cmd
		}

		if m.loading {
			return m, nil
		}

		if m.tabLoading {
			return m, nil
		}

		if top := m.nav.Top(); top != nil {
			if f, ok := top.Model.(inputFocusable); ok && f.InputFocused() {
				cmd := m.nav.Update(msg)
				return m, cmd
			}
		}

		if m.confirm.Active {
			confirmed, _ := m.confirm.HandleKey(msg.String())
			if !m.confirm.Active && confirmed {
				switch m.confirm.Action {
				case "logout":
					_ = keystore.Delete()
					m.client = nil
					m.user = nil
					m.setupMode = true
					s := setup.New()
					s.SetSize(m.width, m.height)
					m.setupScr = s
					_ = m.nav.Clear()
					return m, s.Init()
				}
			}
			return m, nil
		}

		switch {
		case key.Matches(msg, common.Keys.Quit):
			return m, tea.Quit
		case key.Matches(msg, common.Keys.Back):
			if len(m.nav.StackSummary()) > 1 {
				cmd := m.nav.Pop()
				if top := m.nav.Top(); top != nil {
					if s, ok := top.Model.(sizable); ok {
						s.SetSize(m.width, m.contentHeight())
					}
				}
				return m, cmd
			}
			cmd := m.nav.Update(msg)
			return m, cmd
		case key.Matches(msg, common.Keys.Help):
			m.help.ShowAll = !m.help.ShowAll
			return m, nil
		case key.Matches(msg, common.Keys.Logout):
			m.confirm = common.NewConfirm("Are you sure you want to log out?", "logout")
			return m, nil
		case key.Matches(msg, common.Keys.Library):
			return m.switchTab(0)
		case key.Matches(msg, common.Keys.Search):
			return m.switchTab(1)
		case key.Matches(msg, common.Keys.Lists):
			return m.switchTab(2)
		case key.Matches(msg, common.Keys.Stats):
			return m.switchTab(3)
		case key.Matches(msg, common.Keys.NextTab):
			next := (m.activeTab + 1) % len(navTabs)
			return m.switchTab(next)
		case key.Matches(msg, common.Keys.PrevTab):
			prev := (m.activeTab - 1 + len(navTabs)) % len(navTabs)
			return m.switchTab(prev)
		}

		cmd := m.nav.Update(msg)
		return m, cmd
	}

	if m.setupMode && m.setupScr != nil {
		updated, cmd := m.setupScr.Update(msg)
		m.setupScr = updated.(Screen)
		return m, cmd
	}

	if !m.loading {
		cmd := m.nav.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m Model) View() string {
	if m.loading {
		return common.AppStyle.Render(
			fmt.Sprintf("\n  %s Loading...\n", m.spinner.View()),
		)
	}

	if m.setupMode {
		if m.setupScr != nil {
			return m.setupScr.View()
		}
		return ""
	}

	if m.tabLoading {
		return zone.Scan(m.loader.View(m.width, m.height))
	}

	nav := m.renderNav()
	bc := m.renderBreadcrumb()

	var content string
	if top := m.nav.Top(); top != nil {
		content = top.Model.View()
	}

	status := m.renderStatusBar()
	help := m.renderHelp()

	output := lipgloss.JoinVertical(lipgloss.Left,
		nav,
		bc,
		content,
		status,
		help,
	)

	if m.confirm.Active {
		fg := common.RenderConfirmOverlay(m.confirm.Message, m.confirm.Cursor, 50)
		output = overlay.Composite(fg, output, overlay.Center, overlay.Center, 0, 0)
	}

	output = m.alert.Render(output)

	return zone.Scan(output)
}

func (m Model) renderNav() string {
	var items []string
	for i, t := range navTabs {
		label := fmt.Sprintf(" %d %s ", i+1, t.name)
		var rendered string
		if i == m.activeTab {
			rendered = common.ActiveTabStyle.Render(label)
		} else {
			rendered = common.InactiveTabStyle.Render(label)
		}
		items = append(items, zone.Mark(t.zoneID, rendered))
	}
	tabs := lipgloss.JoinHorizontal(lipgloss.Top, items...)

	hs := common.HelpStyles()
	shortcuts := common.HelpStyle.Render(
		hs.ShortKey.Render("?") + " " + hs.ShortDesc.Render("help") + "  " +
			hs.ShortKey.Render("esc") + " " + hs.ShortDesc.Render("back") + "  " +
			hs.ShortKey.Render("ctrl+q") + " " + hs.ShortDesc.Render("logout") + "  " +
			hs.ShortKey.Render("q") + " " + hs.ShortDesc.Render("quit"),
	)

	navW := m.width - 2 // AppStyle padding
	if navW < 40 {
		navW = 80
	}
	gap := navW - lipgloss.Width(tabs) - lipgloss.Width(shortcuts)
	if gap < 1 {
		gap = 1
	}

	return lipgloss.NewStyle().PaddingLeft(1).Render(
		tabs + strings.Repeat(" ", gap) + shortcuts,
	)
}

func (m Model) renderBreadcrumb() string {
	summary := m.nav.StackSummary()
	if len(summary) <= 1 {
		return ""
	}
	parts := make([]string, len(summary))
	for i, s := range summary {
		if i == len(summary)-1 {
			parts[i] = common.LabelStyle.Render(s)
		} else {
			parts[i] = common.ValueStyle.Render(s)
		}
	}
	return lipgloss.NewStyle().Padding(0, 1).Render(
		common.HelpStyle.Render(strings.Join(parts, " > ")),
	)
}

func (m Model) renderStatusBar() string {
	return ""
}

func (m Model) renderHelp() string {
	var pageBindings []key.Binding
	if top := m.nav.Top(); top != nil {
		if hb, ok := top.Model.(common.HelpBindable); ok {
			pageBindings = hb.HelpBindings()
		}
	}

	var fullOnlyBindings []key.Binding
	if top := m.nav.Top(); top != nil {
		if fhb, ok := top.Model.(common.FullHelpBindable); ok {
			fullOnlyBindings = fhb.FullHelpBindings()
		}
	}

	var helpView string
	if m.help.ShowAll {
		var all []key.Binding
		all = append(all, pageBindings...)
		all = append(all, fullOnlyBindings...)
		for _, group := range m.keys.FullHelp() {
			all = append(all, group...)
		}
		const colSize = 4
		var groups [][]key.Binding
		for i := 0; i < len(all); i += colSize {
			end := i + colSize
			if end > len(all) {
				end = len(all)
			}
			groups = append(groups, all[i:end])
		}
		helpView = m.help.FullHelpView(groups)
	} else {
		bindings := make([]key.Binding, 0, len(pageBindings)+3)
		bindings = append(bindings, pageBindings...)
		bindings = append(bindings, m.keys.ShortHelp()...)
		helpView = m.help.ShortHelpView(bindings)
	}

	var userPart string
	if m.user != nil {
		userPart = common.StatusBarStyle.Render(fmt.Sprintf("@%s", m.user.Username))
		if m.user.Pro {
			userPart += " " + common.SuccessStyle.Render("PRO")
		}
	}

	helpWidth := m.width
	if helpWidth <= 0 {
		helpWidth = 80
	}

	lines := strings.SplitN(helpView, "\n", 2)
	firstLine := lines[0]

	combined := lipgloss.NewStyle().Width(helpWidth).PaddingLeft(1).Render(
		lipgloss.JoinHorizontal(lipgloss.Top,
			firstLine,
			lipgloss.PlaceHorizontal(helpWidth-lipgloss.Width(firstLine)-1, lipgloss.Right, userPart),
		),
	)

	if len(lines) > 1 {
		return combined + "\n" + lines[1]
	}
	return combined
}
