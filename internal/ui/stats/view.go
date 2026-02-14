package stats

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/lipgloss"

	"hardcover-tui/internal/common"
)

func (m *Model) View() string {
	if m.loading {
		return common.AppStyle.Render(
			fmt.Sprintf("\n  %s Loading stats...\n", m.spinner.View()),
		)
	}

	panelW := m.width - 4
	if panelW < 60 {
		panelW = 80
	}

	if panelW != m.lastChartW && m.counts != nil {
		m.buildChartsForWidth(panelW)
	}

	if m.err != nil {
		return common.AppStyle.Render(
			common.ErrorStyle.Render("Error: " + m.err.Error()),
		)
	}

	var sections []string

	{
		var goalsContent strings.Builder
		if len(m.goals) == 0 {
			goalsContent.WriteString(common.ValueStyle.Render("No active goals"))
			goalsContent.WriteString("\n")
			goalsContent.WriteString(common.HelpStyle.Render("Create goals on hardcover.app to track them here"))
		} else {
			barW := panelW - 10
			if barW < 10 {
				barW = 10
			}
			m.goalProgress.Width = barW
			for i, g := range m.goals {
				if i > 0 {
					goalsContent.WriteString("\n\n")
				}
				pct := g.PercentComplete()
				pctInt := int(pct * 100)
				remaining := g.Remaining()
				days := g.DaysRemaining()
				name := g.DisplayName()

				goalsContent.WriteString(common.LabelStyle.Render(name))
				goalsContent.WriteString("\n")
				goalsContent.WriteString(m.goalProgress.ViewAs(pct))
				goalsContent.WriteString("\n")

				statsLine := fmt.Sprintf("%d/%d %s", int(g.Progress), g.Goal, g.Metric)
				statsLine += fmt.Sprintf("  |  %d %s remaining", remaining, g.Metric)
				statsLine += fmt.Sprintf("  |  %d days left", days)
				statsLine += fmt.Sprintf("  |  %d%%", pctInt)
				goalsContent.WriteString(common.ValueStyle.Render(statsLine))

				if g.Description != nil && *g.Description != "" && *g.Description != name {
					goalsContent.WriteString("\n")
					goalsContent.WriteString(common.HelpStyle.Render(*g.Description))
				}
			}
		}
		sections = append(sections, common.RenderPanel("Reading Goals", goalsContent.String(), panelW))
	}

	{
		var litContent strings.Builder
		litTotal := m.fictionCount + m.nonfictionCount + m.unknownLitCount
		litContent.WriteString(common.LabelStyle.Render(fmt.Sprintf("Total Books: %d", litTotal)))
		litContent.WriteString("\n\n")

		litContent.WriteString(renderPieBar([]pieSlice{
			{Label: "Fiction", Count: m.fictionCount, Color: common.ColorPrimary},
			{Label: "Nonfiction", Count: m.nonfictionCount, Color: common.ColorWarning},
			{Label: "Unknown", Count: m.unknownLitCount, Color: common.ColorMuted},
		}, panelW-6))
		litContent.WriteString("\n\n")

		ficStyle := lipgloss.NewStyle().Foreground(common.ColorPrimary)
		nfStyle := lipgloss.NewStyle().Foreground(common.ColorWarning)
		unkStyle := lipgloss.NewStyle().Foreground(common.ColorMuted)
		barW := panelW - 6

		if litTotal > 0 {
			ficText := ficStyle.Render(fmt.Sprintf("Fiction: %d (%d%%)",
				m.fictionCount, m.fictionCount*100/litTotal))
			nfText := nfStyle.Render(fmt.Sprintf("Nonfiction: %d (%d%%)",
				m.nonfictionCount, m.nonfictionCount*100/litTotal))
			row := lipgloss.JoinHorizontal(lipgloss.Top,
				lipgloss.NewStyle().Width(barW/2).Align(lipgloss.Left).Render(ficText),
				lipgloss.NewStyle().Width(barW-barW/2).Align(lipgloss.Right).Render(nfText),
			)
			litContent.WriteString(row)
			if m.unknownLitCount > 0 {
				litContent.WriteString("\n")
				litContent.WriteString(unkStyle.Render(fmt.Sprintf("Unknown: %d (%d%%)",
					m.unknownLitCount, m.unknownLitCount*100/litTotal)))
			}
		} else {
			litContent.WriteString(common.ValueStyle.Render("No data available"))
		}
		sections = append(sections, common.RenderPanel("Fiction vs. Non-Fiction", litContent.String(), panelW))
	}

	{
		var fmtContent strings.Builder
		fmtTotal := m.physicalCount + m.ebookCount + m.audiobookCount + m.unknownFmtCount
		fmtContent.WriteString(common.LabelStyle.Render(fmt.Sprintf("Total Books: %d", fmtTotal)))
		fmtContent.WriteString("\n\n")

		fmtContent.WriteString(renderPieBar([]pieSlice{
			{Label: "Physical", Count: m.physicalCount, Color: common.ColorPrimary},
			{Label: "Ebook", Count: m.ebookCount, Color: common.ColorSuccess},
			{Label: "Audiobook", Count: m.audiobookCount, Color: common.ColorWarning},
			{Label: "Unknown", Count: m.unknownFmtCount, Color: common.ColorMuted},
		}, panelW-6))
		fmtContent.WriteString("\n\n")

		physStyle := lipgloss.NewStyle().Foreground(common.ColorPrimary)
		eStyle := lipgloss.NewStyle().Foreground(common.ColorSuccess)
		aStyle := lipgloss.NewStyle().Foreground(common.ColorWarning)
		unkStyle := lipgloss.NewStyle().Foreground(common.ColorMuted)
		barW := panelW - 6

		if fmtTotal > 0 {
			physText := physStyle.Render(fmt.Sprintf("Physical: %d (%d%%)",
				m.physicalCount, m.physicalCount*100/fmtTotal))
			eText := eStyle.Render(fmt.Sprintf("Ebook: %d (%d%%)",
				m.ebookCount, m.ebookCount*100/fmtTotal))
			aText := aStyle.Render(fmt.Sprintf("Audiobook: %d (%d%%)",
				m.audiobookCount, m.audiobookCount*100/fmtTotal))
			colW := barW / 3
			row := lipgloss.JoinHorizontal(lipgloss.Top,
				lipgloss.NewStyle().Width(colW).Align(lipgloss.Left).Render(physText),
				lipgloss.NewStyle().Width(colW).Align(lipgloss.Center).Render(eText),
				lipgloss.NewStyle().Width(barW-2*colW).Align(lipgloss.Right).Render(aText),
			)
			fmtContent.WriteString(row)
			if m.unknownFmtCount > 0 {
				fmtContent.WriteString("\n")
				fmtContent.WriteString(unkStyle.Render(fmt.Sprintf("Unknown: %d (%d%%)",
					m.unknownFmtCount, m.unknownFmtCount*100/fmtTotal)))
			}
		} else {
			fmtContent.WriteString(common.ValueStyle.Render("No data available"))
		}
		sections = append(sections, common.RenderPanel("Physical vs. Ebook vs. Audiobook", fmtContent.String(), panelW))
	}

	if m.timeChartReady {
		var tsContent strings.Builder
		tsContent.WriteString(m.timeChart.View())
		tsContent.WriteString("\n\n")
		physBlock := lipgloss.NewStyle().Foreground(common.ColorPrimary).Render("━━")
		ebookBlock := lipgloss.NewStyle().Foreground(common.ColorSuccess).Render("━━")
		audioBlock := lipgloss.NewStyle().Foreground(common.ColorWarning).Render("━━")
		tsContent.WriteString(physBlock + " Physical (Read)  ")
		tsContent.WriteString(ebookBlock + " Ebooks (Read)  ")
		tsContent.WriteString(audioBlock + " Audiobooks (Listened)")
		sections = append(sections, common.RenderPanel("Pages Over Time", tsContent.String(), panelW))
	}

	{
		topGenres := sortedTopN(m.genreCounts, 10)
		if len(topGenres) > 0 {
			chartView := m.genreChart.View()
			legendView := renderChartLegend(m.genreLabels, panelW/3)
			combined := lipgloss.JoinHorizontal(lipgloss.Top, chartView, "  ", legendView)
			sections = append(sections, common.RenderPanel("Top Genres", combined, panelW))
		} else {
			sections = append(sections, common.RenderPanel("Top Genres", common.ValueStyle.Render("No genre data"), panelW))
		}
	}

	fullContent := lipgloss.JoinVertical(lipgloss.Left, sections...)

	contentH := lipgloss.Height(fullContent)
	availH := m.height
	if availH < 5 {
		availH = 30
	}

	if contentH > availH {
		if !m.vpReady || m.vp.Width != panelW || m.vp.Height != availH {
			m.vp = viewport.New(panelW, availH)
			m.vp.Style = lipgloss.NewStyle()
			m.vpReady = true
			m.lastVpContent = ""
		}
		if fullContent != m.lastVpContent {
			m.lastVpContent = fullContent
			m.vp.SetContent(fullContent)
		}
		return common.AppStyle.Render(m.vp.View())
	}

	m.vpReady = false
	return common.AppStyle.Render(fullContent)
}

// pieSlice represents a segment for renderPieBar.
type pieSlice struct {
	Label string
	Count int
	Color lipgloss.Color
}

// renderPieBar renders a horizontal "pie bar" showing colored proportional segments.
func renderPieBar(slices []pieSlice, width int) string {
	total := 0
	for _, s := range slices {
		total += s.Count
	}
	if total == 0 || width <= 0 {
		return common.ValueStyle.Render(strings.Repeat("\u2591", width))
	}

	var bar strings.Builder
	remaining := width
	for i, s := range slices {
		segW := s.Count * width / total
		if i == len(slices)-1 {
			segW = remaining // last segment gets the remainder
		}
		if segW > remaining {
			segW = remaining
		}
		if segW > 0 {
			style := lipgloss.NewStyle().Foreground(s.Color)
			bar.WriteString(style.Render(strings.Repeat("\u2588", segW)))
			remaining -= segW
		}
	}

	return bar.String()
}

// renderChartLegend renders a numbered legend for vertical bar charts.
// Each entry on its own line with a colored block for clarity.
func renderChartLegend(labels []chartLabel, maxW int) string {
	if len(labels) == 0 {
		return ""
	}
	var b strings.Builder
	for i, l := range labels {
		style := lipgloss.NewStyle().Foreground(l.Color)
		block := style.Render("\u2588")
		entry := fmt.Sprintf(" %d. %s (%d)", l.Index, l.Name, l.Count)
		b.WriteString(block + entry)
		if i < len(labels)-1 {
			b.WriteString("\n")
		}
	}
	return lipgloss.NewStyle().Width(maxW).Render(b.String())
}
