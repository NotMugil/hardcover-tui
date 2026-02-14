package stats

import (
	"sort"

	"github.com/76creates/stickers/flexbox"
	"github.com/NimbleMarkets/ntcharts/barchart"
	tslc "github.com/NimbleMarkets/ntcharts/linechart/timeserieslinechart"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"hardcover-tui/internal/api"
	"hardcover-tui/internal/common"
)

type statsLoadedMsg struct {
	goals          []api.Goal
	counts         map[api.StatusID]int
	userBooks      []api.StatsUserBook
	readingHistory []api.ReadingHistoryEntry
	err            error
}

// chartLabel stores a legend entry for a vertical bar chart.
type chartLabel struct {
	Index int
	Name  string
	Count int
	Color lipgloss.Color
}

// Model is the stats screen model.
type Model struct {
	client    *api.Client
	user      *api.User
	goals     []api.Goal
	counts    map[api.StatusID]int
	userBooks []api.StatsUserBook
	fictionCount    int
	nonfictionCount int
	unknownLitCount int
	physicalCount   int
	ebookCount      int
	audiobookCount  int
	unknownFmtCount int
	genreCounts     map[string]int
	genreChart     barchart.Model
	genreLabels    []chartLabel
	timeChart      tslc.Model
	timeChartReady bool
	readingHistory []api.ReadingHistoryEntry
	goalProgress progress.Model
	vp            viewport.Model
	vpReady       bool
	lastVpContent string // cache to avoid resetting scroll on identical content
	spinner       spinner.Model
	loading       bool
	err           error
	width         int
	height        int
	flexBox       *flexbox.FlexBox
	lastChartW    int // track last width charts were built for
}

// New creates a new stats screen.
func New(client *api.Client, user *api.User) *Model {
	s := spinner.New(
		spinner.WithSpinner(spinner.Dot),
		spinner.WithStyle(common.SpinnerStyle),
	)

	fb := flexbox.New(0, 0)
	row0 := fb.NewRow().AddCells(flexbox.NewCell(1, 3))
	row1 := fb.NewRow().AddCells(flexbox.NewCell(1, 2))
	row2 := fb.NewRow().AddCells(flexbox.NewCell(1, 3))
	fb.AddRows([]*flexbox.Row{row0, row1, row2})

	gp := progress.New(
		progress.WithSolidFill(string(common.ColorPrimary)),
	)

	return &Model{
		client:       client,
		user:         user,
		spinner:      s,
		loading:      true,
		flexBox:      fb,
		goalProgress: gp,
	}
}

// SetSize updates the available terminal dimensions.
func (m *Model) SetSize(w, h int) {
	m.width = w
	m.height = h
}

// Loaded returns true once the stats data has been fetched.
func (m *Model) Loaded() bool {
	return !m.loading
}

func (m *Model) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, m.loadStats())
}

// HelpBindings returns page-specific keybindings for the help bar.
func (m *Model) HelpBindings() []key.Binding {
	return []key.Binding{
		key.NewBinding(key.WithKeys("j", "down"), key.WithHelp("j/↓", "scroll down")),
		key.NewBinding(key.WithKeys("k", "up"), key.WithHelp("k/↑", "scroll up")),
	}
}

var chartColors = []lipgloss.Color{
	common.ColorPrimary,
	common.ColorSuccess,
	common.ColorWarning,
	common.ColorDanger,
	common.ColorSecondary,
	lipgloss.Color("#a78bfa"), // violet
	lipgloss.Color("#f472b6"), // pink
	lipgloss.Color("#34d399"), // emerald
	lipgloss.Color("#fbbf24"), // amber
	lipgloss.Color("#60a5fa"), // blue
}

// sortedTopN returns the top N entries from a map, sorted by count descending.
type countEntry struct {
	Name  string
	Count int
}

func sortedTopN(m map[string]int, n int) []countEntry {
	entries := make([]countEntry, 0, len(m))
	for k, v := range m {
		entries = append(entries, countEntry{Name: k, Count: v})
	}
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Count > entries[j].Count
	})
	if len(entries) > n {
		entries = entries[:n]
	}
	return entries
}
