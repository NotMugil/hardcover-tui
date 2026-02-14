package stats

import (
	"fmt"
	"sort"
	"time"

	"github.com/NimbleMarkets/ntcharts/barchart"
	"github.com/NimbleMarkets/ntcharts/canvas/runes"
	tslc "github.com/NimbleMarkets/ntcharts/linechart/timeserieslinechart"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/NotMugil/hardcover-tui/internal/api"
	"github.com/NotMugil/hardcover-tui/internal/common"
)

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case statsLoadedMsg:
		m.loading = false
		if msg.err != nil {
			m.err = msg.err
			return m, nil
		}
		m.goals = msg.goals
		m.counts = msg.counts
		m.userBooks = msg.userBooks
		m.readingHistory = msg.readingHistory
		m.computeStats()
		m.buildCharts()
		return m, nil

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
		if m.vpReady {
			var cmd tea.Cmd
			m.vp, cmd = m.vp.Update(msg)
			return m, cmd
		}
	}
	return m, nil
}

func (m *Model) computeStats() {
	m.genreCounts = make(map[string]int)
	m.fictionCount = 0
	m.nonfictionCount = 0
	m.unknownLitCount = 0
	m.physicalCount = 0
	m.ebookCount = 0
	m.audiobookCount = 0
	m.unknownFmtCount = 0

	for _, ub := range m.userBooks {
		if ub.LiteraryTypeID != nil {
			switch *ub.LiteraryTypeID {
			case api.LiteraryTypeFiction:
				m.fictionCount++
			case api.LiteraryTypeNonfiction:
				m.nonfictionCount++
			default:
				m.unknownLitCount++
			}
		} else {
			m.unknownLitCount++
		}

		if ub.EditionFormat != nil {
			switch *ub.EditionFormat {
			case "audiobook", "audio_cd", "audio_cassette":
				m.audiobookCount++
			case "ebook", "kindle_edition":
				m.ebookCount++
			case "hardcover", "paperback", "mass_market_paperback", "library_binding", "board_book":
				m.physicalCount++
			default:
				m.physicalCount++ // default to physical for unrecognized formats
			}
		} else {
			m.unknownFmtCount++
		}

		for _, g := range ub.Genres {
			m.genreCounts[g]++
		}

	}
}

func (m *Model) buildCharts() {
	m.buildChartsForWidth(m.width)
}

func (m *Model) buildChartsForWidth(panelW int) {
	if panelW < 30 {
		panelW = 30
	}
	m.lastChartW = panelW

	chartW := (panelW * 55 / 100) - 6
	if chartW < 20 {
		chartW = 20
	}

	axisStyle := lipgloss.NewStyle().Foreground(common.ColorBorder)
	labelStyle := lipgloss.NewStyle().Foreground(common.ColorSubtext)

	topGenres := sortedTopN(m.genreCounts, 10)
	genreH := 12
	m.genreChart = barchart.New(chartW, genreH,
		barchart.WithStyles(axisStyle, labelStyle),
	)
	var genreBars []barchart.BarData
	m.genreLabels = nil
	for i, g := range topGenres {
		color := chartColors[i%len(chartColors)]
		style := lipgloss.NewStyle().Foreground(color)
		label := fmt.Sprintf("%d", i+1)
		genreBars = append(genreBars, barchart.BarData{
			Label: label,
			Values: []barchart.BarValue{
				{Name: g.Name, Value: float64(g.Count), Style: style},
			},
		})
		m.genreLabels = append(m.genreLabels, chartLabel{
			Index: i + 1,
			Name:  g.Name,
			Count: g.Count,
			Color: color,
		})
	}
	m.genreChart.PushAll(genreBars)
	m.genreChart.Draw()

	m.timeChartReady = false
	if len(m.readingHistory) > 0 {
		m.buildTimeChart(panelW)
	}
}

// buildTimeChart constructs the time series line chart for pages read over time.
func (m *Model) buildTimeChart(panelW int) {
	chartW := panelW - 8
	if chartW < 30 {
		chartW = 30
	}
	chartH := 14

	type monthKey struct {
		Year  int
		Month time.Month
	}
	physicalByMonth := make(map[monthKey]float64)
	ebookByMonth := make(map[monthKey]float64)
	audiobookByMonth := make(map[monthKey]float64)

	for _, entry := range m.readingHistory {
		t, err := time.Parse(time.RFC3339Nano, entry.FinishedAt)
		if err != nil {
			t, err = time.Parse(time.RFC3339, entry.FinishedAt)
		}
		if err != nil {
			t, err = time.Parse("2006-01-02T15:04:05", entry.FinishedAt)
		}
		if err != nil {
			t, err = time.Parse("2006-01-02", entry.FinishedAt)
		}
		if err != nil {
			continue
		}

		mk := monthKey{Year: t.Year(), Month: t.Month()}
		switch entry.EditionFormat {
		case "audiobook":
			audiobookByMonth[mk] += float64(entry.Pages)
		case "ebook":
			ebookByMonth[mk] += float64(entry.Pages)
		default:
			physicalByMonth[mk] += float64(entry.Pages)
		}
	}

	allMonths := make(map[monthKey]bool)
	for k := range physicalByMonth {
		allMonths[k] = true
	}
	for k := range ebookByMonth {
		allMonths[k] = true
	}
	for k := range audiobookByMonth {
		allMonths[k] = true
	}

	if len(allMonths) == 0 {
		return
	}

	sorted := make([]monthKey, 0, len(allMonths))
	for k := range allMonths {
		sorted = append(sorted, k)
	}
	sort.Slice(sorted, func(i, j int) bool {
		if sorted[i].Year != sorted[j].Year {
			return sorted[i].Year < sorted[j].Year
		}
		return sorted[i].Month < sorted[j].Month
	})

	physColor := lipgloss.NewStyle().Foreground(common.ColorPrimary)
	ebookColor := lipgloss.NewStyle().Foreground(common.ColorSuccess)
	audioColor := lipgloss.NewStyle().Foreground(common.ColorWarning)

	axisStyle := lipgloss.NewStyle().Foreground(common.ColorBorder)
	labelStyle := lipgloss.NewStyle().Foreground(common.ColorSubtext)

	m.timeChart = tslc.New(chartW, chartH,
		tslc.WithDataSetStyle("Physical (Read)", physColor),
		tslc.WithDataSetLineStyle("Physical (Read)", runes.ThinLineStyle),
		tslc.WithDataSetStyle("Ebooks (Read)", ebookColor),
		tslc.WithDataSetLineStyle("Ebooks (Read)", runes.ThinLineStyle),
		tslc.WithDataSetStyle("Audiobooks (Listened)", audioColor),
		tslc.WithDataSetLineStyle("Audiobooks (Listened)", runes.ThinLineStyle),
		tslc.WithAxesStyles(axisStyle, labelStyle),
	)

	for _, mk := range sorted {
		t := time.Date(mk.Year, mk.Month, 15, 0, 0, 0, 0, time.UTC)

		if v := physicalByMonth[mk]; v > 0 {
			m.timeChart.PushDataSet("Physical (Read)", tslc.TimePoint{Time: t, Value: v})
		}
		if v := ebookByMonth[mk]; v > 0 {
			m.timeChart.PushDataSet("Ebooks (Read)", tslc.TimePoint{Time: t, Value: v})
		}
		if v := audiobookByMonth[mk]; v > 0 {
			m.timeChart.PushDataSet("Audiobooks (Listened)", tslc.TimePoint{Time: t, Value: v})
		}
	}

	m.timeChart.DrawAll()
	m.timeChartReady = true
}
