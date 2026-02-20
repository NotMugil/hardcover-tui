package home

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/lipgloss"
	overlay "github.com/rmhubbert/bubbletea-overlay"

	"github.com/NotMugil/hardcover-tui/internal/api"
	"github.com/NotMugil/hardcover-tui/internal/common"
)

// renderReadingItems renders currently-reading books with padding and pagination.
func (m *Model) renderReadingItems(innerW, maxH int) string {
	if len(m.reading) == 0 {
		return common.ValueStyle.Render("Nothing currently reading")
	}

	barW := innerW - 10
	if barW < 8 {
		barW = 8
	}
	m.progress.Width = barW

	type entry struct {
		text string
	}
	var entries []entry
	for i, ub := range m.reading {
		var b strings.Builder
		title := common.Truncate(ub.Book.Title, innerW)
		if m.readingFocused && i == m.readingCursor {
			b.WriteString(common.LabelStyle.Render("> " + title))
		} else {
			b.WriteString(common.LabelStyle.Render("  " + title))
		}
		b.WriteString("\n")
		author := common.Truncate("  by "+ub.Book.Authors(), innerW)
		b.WriteString(common.ValueStyle.Render(author))
		if len(ub.UserBookReads) > 0 {
			read := ub.UserBookReads[0]
			if read.ProgressPages != nil && ub.Book.Pages != nil && *ub.Book.Pages > 0 {
				pct := float64(*read.ProgressPages) / float64(*ub.Book.Pages)
				b.WriteString("\n")
				b.WriteString(fmt.Sprintf("  %s %d%%", m.progress.ViewAs(pct), int(pct*100)))
			}
		}
		entries = append(entries, entry{text: b.String()})
	}

	linesPerEntry := 4 // title + author + progress + padding
	visible := maxH / linesPerEntry
	if visible < 1 {
		visible = 1
	}

	if m.readingCursor < m.readingScroll {
		m.readingScroll = m.readingCursor
	}
	if m.readingCursor >= m.readingScroll+visible {
		m.readingScroll = m.readingCursor - visible + 1
	}
	if m.readingScroll < 0 {
		m.readingScroll = 0
	}

	end := m.readingScroll + visible
	if end > len(entries) {
		end = len(entries)
	}

	var out strings.Builder
	for i := m.readingScroll; i < end; i++ {
		if i > m.readingScroll {
			out.WriteString("\n\n") // padding between items
		}
		out.WriteString(entries[i].text)
	}

	if len(entries) > visible {
		out.WriteString("\n")
		out.WriteString(common.HelpStyle.Render(
			fmt.Sprintf("  %d-%d of %d", m.readingScroll+1, end, len(entries))))
	}

	return out.String()
}

func (m *Model) renderFilterBar() string {
	var parts []string
	for i, name := range filterNames {
		if i == m.filter {
			style := lipgloss.NewStyle().
				Bold(true).
				Foreground(filterColors[i]).
				Background(common.ColorHighlight).
				Padding(0, 1)
			parts = append(parts, style.Render(name))
		} else {
			style := lipgloss.NewStyle().
				Foreground(common.ColorMuted).
				Padding(0, 1)
			parts = append(parts, style.Render(name))
		}
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, parts...)
}

func (m *Model) View() string {
	if m.loading {
		return common.AppStyle.Render(
			fmt.Sprintf("\n  %s Loading library...\n", m.spinner.View()),
		)
	}

	if m.err != nil {
		return common.AppStyle.Render(
			common.ErrorStyle.Render("Error: " + m.err.Error()),
		)
	}

	fbW := m.width - 2
	if fbW < 40 {
		fbW = 80
	}
	fbH := m.height
	if fbH < 10 {
		fbH = 30
	}
	m.flexBox.SetWidth(fbW)
	m.flexBox.SetHeight(fbH)
	m.flexBox.ForceRecalculate()

	leftCell := m.flexBox.GetRow(0).GetCell(0)
	centerCell := m.flexBox.GetRow(0).GetCell(1)
	rightCell := m.flexBox.GetRow(0).GetCell(2)
	leftW := leftCell.GetWidth()
	centerW := centerCell.GetWidth()
	rightW := rightCell.GetWidth()
	cellH := leftCell.GetHeight()

	panelInnerW := func(w int) int {
		return w - 4
	}

	name := m.user.Username
	if m.user.Name != nil && *m.user.Name != "" {
		name = *m.user.Name
	}

	var statsContent strings.Builder
	statsContent.WriteString(common.TitleStyle.Render(fmt.Sprintf("@%s", m.user.Username)))
	if m.user.Name != nil && *m.user.Name != "" {
		statsContent.WriteString("\n" + common.ValueStyle.Render(name))
	}
	if m.user.Pro {
		statsContent.WriteString("  " + common.SuccessStyle.Render("PRO"))
	}
	if m.user.Flair != nil && *m.user.Flair != "" {
		statsContent.WriteString("\n" + lipgloss.NewStyle().Foreground(common.ColorAccent).Bold(true).Render(*m.user.Flair))
	}
	statsContent.WriteString("\n")
	statsContent.WriteString(fmt.Sprintf("Reads: %s\n",
		common.LabelStyle.Render(fmt.Sprintf("%d", m.user.BooksCount)),
	))
	statsContent.WriteString(fmt.Sprintf("Followers: %s\n",
		common.LabelStyle.Render(fmt.Sprintf("%d", m.user.FollowersCount)),
	))
	statsContent.WriteString(fmt.Sprintf("Following: %s",
		common.LabelStyle.Render(fmt.Sprintf("%d", m.user.FollowedUsersCount)),
	))

	var userPanelContent string
	if m.avatarArt != "" {
		userPanelContent = lipgloss.JoinHorizontal(lipgloss.Top,
			m.avatarArt, " ", statsContent.String())
	} else {
		userPanelContent = statsContent.String()
	}

	anyFocused := m.readingFocused || m.activityFocused
	renderPanel := common.RenderPanel
	if anyFocused {
		renderPanel = common.RenderDimPanel
	}

	userPanel := renderPanel("Profile", userPanelContent, leftW)

	var leftPanels []string
	leftPanels = append(leftPanels, userPanel)

	if len(m.reading) > 0 {
		profileH := lipgloss.Height(userPanel)
		clockH := 4 // 2 content + 2 border
		readingH := cellH - profileH - clockH
		if readingH < 8 {
			readingH = 8
		}
		innerW := panelInnerW(leftW)

		readingTitle := "Progress Tracker"

		innerH := readingH - 2 // border
		readingContent := m.renderReadingItems(innerW, innerH)

		var readingPanel string
		if m.readingFocused {
			readingPanel = common.RenderActivePanel(readingTitle, readingContent, leftW, readingH)
		} else {
			readingPanel = renderPanel(readingTitle, readingContent, leftW, readingH)
		}
		leftPanels = append(leftPanels, readingPanel)
	}

	zone, _ := m.currentTime.Zone()
	clockStr := m.currentTime.Format("03:04:05 PM")
	dateStr := m.currentTime.Format("Mon, Jan 02 2006")
	clockContent := common.LabelStyle.Render(clockStr) + " " + common.ValueStyle.Render(zone) + "\n" + common.ValueStyle.Render(dateStr)
	clockPanel := renderPanel("Clock", clockContent, leftW)
	leftPanels = append(leftPanels, clockPanel)

	leftCol := lipgloss.JoinVertical(lipgloss.Left, leftPanels...)
	leftCell.SetContent(leftCol)

	libPanelH := cellH
	if libPanelH < 8 {
		libPanelH = 8
	}

	listW := panelInnerW(centerW)
	listH := libPanelH - 6 // panel border + filter bar + legend
	if listW > 0 && listH > 0 {
		m.list.SetSize(listW, listH)
	}

	var lib strings.Builder
	lib.WriteString(m.renderFilterBar())
	lib.WriteString("\n")
	if m.booksLoading {
		lib.WriteString(fmt.Sprintf("  %s Loading...\n", m.spinner.View()))
	} else {
		lib.WriteString(m.list.View())
	}

	legendStr := ""
	if m.filter == 0 {
		legendStr = renderStatusLegend()
	}
	contentH := lipgloss.Height(lib.String())
	innerH := libPanelH - 2 // subtract panel borders
	pad := innerH - contentH
	if legendStr != "" {
		pad -= 1 // legend line
	}
	if pad < 0 {
		pad = 0
	}
	lib.WriteString(strings.Repeat("\n", pad))
	if legendStr != "" {
		lib.WriteString(legendStr)
	}

	centerPanel := renderPanel("Library", lib.String(), centerW, libPanelH)
	centerCell.SetContent(centerPanel)

	activityPanel := m.renderActivityPanel(rightW, cellH)
	rightCell.SetContent(activityPanel)

	body := m.flexBox.Render()

	if m.confirm.Active {
		fg := common.RenderConfirmOverlay(m.confirm.Message, m.confirm.Cursor, 50)
		body = overlay.Composite(fg, body, overlay.Center, overlay.Center, 0, 0)
	}

	return common.AppStyle.Render(body)
}

// renderStatusLegend returns a compact one-line status color legend.
func renderStatusLegend() string {
	statuses := []struct {
		id    int
		label string
	}{
		{1, "Want"},
		{2, "Reading"},
		{3, "Read"},
		{4, "Paused"},
		{5, "DNF"},
		{6, "Ignored"},
	}
	var parts []string
	for _, s := range statuses {
		color := common.StatusColor(s.id)
		sq := lipgloss.NewStyle().Foreground(color).Bold(true).Render("■")
		parts = append(parts, sq+" "+common.HelpStyle.Render(s.label))
	}
	return strings.Join(parts, "  ")
}

// renderActivityPanel renders the activity feed for the right column.
func (m *Model) renderActivityPanel(width, height int) string {
	innerW := width - 4

	var filterBar strings.Builder
	meLabel := "Mine"
	forYouLabel := "For You"
	if m.activityFilter == activityFilterMe {
		meLabel = lipgloss.NewStyle().Bold(true).Foreground(common.ColorPrimary).
			Background(common.ColorHighlight).Padding(0, 1).Render(meLabel)
		forYouLabel = lipgloss.NewStyle().Foreground(common.ColorMuted).Padding(0, 1).Render(forYouLabel)
	} else {
		meLabel = lipgloss.NewStyle().Foreground(common.ColorMuted).Padding(0, 1).Render(meLabel)
		forYouLabel = lipgloss.NewStyle().Bold(true).Foreground(common.ColorPrimary).
			Background(common.ColorHighlight).Padding(0, 1).Render(forYouLabel)
	}
	filterBar.WriteString(meLabel)
	filterBar.WriteString(forYouLabel)

	var content strings.Builder
	content.WriteString(filterBar.String())
	content.WriteString("\n")

	if m.activityLoading {
		content.WriteString(fmt.Sprintf("\n%s Loading...", m.spinner.View()))
	} else if m.activityErr != nil {
		content.WriteString("\n" + common.ErrorStyle.Render(m.activityErr.Error()))
	} else if len(m.activities) == 0 {
		content.WriteString("\n" + common.ValueStyle.Render("No activity yet"))
	} else {
		linesPerItem := 3
		availH := height - 4 // panel borders(2) + filter bar(1) + pagination(1)
		if availH < 3 {
			availH = 3
		}
		visible := availH / linesPerItem
		if visible < 1 {
			visible = 1
		}

		if m.activityCursor < m.activityScroll {
			m.activityScroll = m.activityCursor
		}
		if m.activityCursor >= m.activityScroll+visible {
			m.activityScroll = m.activityCursor - visible + 1
		}
		if m.activityScroll < 0 {
			m.activityScroll = 0
		}

		end := m.activityScroll + visible
		if end > len(m.activities) {
			end = len(m.activities)
		}

		var items strings.Builder
		for i := m.activityScroll; i < end; i++ {
			if i > m.activityScroll {
				items.WriteString("\n")
			}
			act := m.activities[i]
			cursor := "  "
			if m.activityFocused && i == m.activityCursor {
				cursor = "> "
			}
			line := renderActivityItem(act, innerW-2, m.activityFilter == activityFilterForYou)
			lines := strings.Split(line, "\n")
			for j, l := range lines {
				if j == 0 {
					items.WriteString(cursor + l)
				} else {
					items.WriteString("\n  " + l)
				}
			}
		}

		if len(m.activities) > visible {
			items.WriteString("\n")
			items.WriteString(common.HelpStyle.Render(
				fmt.Sprintf("  %d-%d of %d", m.activityScroll+1, end, len(m.activities))))
		}

		content.WriteString(items.String())
	}

	title := "Activity"
	if m.activityFocused {
		return common.RenderActivePanel(title, content.String(), width, height)
	}
	if m.readingFocused {
		return common.RenderDimPanel(title, content.String(), width, height)
	}
	return common.RenderPanel(title, content.String(), width, height)
}

// renderActivityItem renders a single activity entry as a descriptive sentence.
func renderActivityItem(act api.Activity, maxW int, showUser bool) string {
	var b strings.Builder

	prefix := ""
	if showUser && act.User != nil {
		prefix = common.TitleStyle.Render("@"+act.User.Username) + " "
	}

	bookTitle := ""
	if act.Book != nil {
		bookTitle = common.ValueStyle.Render(common.Truncate(act.Book.Title, maxW-4))
	}

	data := act.ParseData()
	var label string

	switch act.Event {
	case "UserBookActivity":
		label = describeUserBookActivity(data, bookTitle)
	case "GoalActivity":
		label = describeGoalActivity(data)
	case "ListActivity":
		label = describeListActivity(data, bookTitle)
	case "PromptActivity":
		label = describePromptActivity(data)
	default:
		label = describeFallbackActivity(act.Event, bookTitle)
	}

	b.WriteString(prefix + label)
	b.WriteString("\n  " + common.HelpStyle.Render(relativeTime(act.CreatedAt)))

	return b.String()
}

func describeUserBookActivity(data api.ActivityParsedData, bookTitle string) string {
	if data.UserBook == nil {
		return withBook(common.LabelStyle.Render("Updated"), bookTitle)
	}
	ub := data.UserBook

	if ub.Review != nil && *ub.Review != "" {
		return withBook(common.LabelStyle.Render("Reviewed"), bookTitle)
	}

	if ub.Rating != nil && *ub.Rating != "" {
		return withBook(common.LabelStyle.Render("Rated "+*ub.Rating), bookTitle)
	}

	if ub.StatusID != nil {
		var status string
		switch *ub.StatusID {
		case 1:
			status = "Wants to read"
		case 2:
			status = "Started reading"
		case 3:
			status = "Finished"
		case 4:
			status = "Paused"
		case 5:
			status = "Did not finish"
		case 6:
			status = "Removed"
		default:
			status = "Updated"
		}
		return withBook(common.LabelStyle.Render(status), bookTitle)
	}

	return withBook(common.LabelStyle.Render("Updated"), bookTitle)
}

func describeGoalActivity(data api.ActivityParsedData) string {
	if data.Goal == nil {
		return common.LabelStyle.Render("Updated reading goal")
	}
	g := data.Goal
	if g.PercentComplete >= 1.0 {
		return common.LabelStyle.Render("Completed reading goal")
	}
	if g.Description != "" {
		return common.LabelStyle.Render("Set goal: ") +
			common.ValueStyle.Render(g.Description)
	}
	return common.LabelStyle.Render("Set a reading goal")
}

func describeListActivity(data api.ActivityParsedData, bookTitle string) string {
	if data.List == nil {
		return withBook(common.LabelStyle.Render("Updated a list"), bookTitle)
	}
	return withBook(
		common.LabelStyle.Render("Updated list: ")+common.ValueStyle.Render(data.List.Name),
		bookTitle,
	)
}

func describePromptActivity(data api.ActivityParsedData) string {
	if data.Prompt != nil && data.Prompt.Question != "" {
		return common.LabelStyle.Render("Answered: ") +
			common.ValueStyle.Render(data.Prompt.Question)
	}
	return common.LabelStyle.Render("Answered a prompt")
}

func describeFallbackActivity(event, bookTitle string) string {
	s := strings.ReplaceAll(event, "_", " ")
	if len(s) > 0 {
		s = strings.ToUpper(s[:1]) + s[1:]
	}
	return withBook(common.LabelStyle.Render(s), bookTitle)
}

func withBook(label, bookTitle string) string {
	if bookTitle != "" {
		return label + "\n  " + bookTitle
	}
	return label
}

// relativeTime parses a timestamp and returns a human-readable relative time.
func relativeTime(ts string) string {
	for _, format := range []string{
		time.RFC3339Nano,
		time.RFC3339,
		"2006-01-02T15:04:05",
		"2006-01-02",
	} {
		if t, err := time.Parse(format, ts); err == nil {
			d := time.Since(t)
			switch {
			case d < time.Minute:
				return "just now"
			case d < time.Hour:
				return fmt.Sprintf("%dm ago", int(d.Minutes()))
			case d < 24*time.Hour:
				return fmt.Sprintf("%dh ago", int(d.Hours()))
			case d < 7*24*time.Hour:
				return fmt.Sprintf("%dd ago", int(d.Hours()/24))
			default:
				return t.Format("Jan 02")
			}
		}
	}
	return ts
}

// HelpBindings returns page-specific keybindings for the global help bar.
func (m *Model) HelpBindings() []key.Binding {
	return []key.Binding{
		key.NewBinding(key.WithKeys("r"), key.WithHelp("r", "reading")),
		key.NewBinding(key.WithKeys("f"), key.WithHelp("f", "filter")),
		key.NewBinding(key.WithKeys("a"), key.WithHelp("a", "activity")),
		key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "details")),
		key.NewBinding(key.WithKeys("/"), key.WithHelp("/", "search")),
	}
}

// FullHelpBindings returns extra keybindings only shown in the expanded help view.
func (m *Model) FullHelpBindings() []key.Binding {
	return []key.Binding{
		key.NewBinding(key.WithKeys("shift+f"), key.WithHelp("shift+f", "filter prev")),
		key.NewBinding(key.WithKeys("]"), key.WithHelp("]", "next page")),
		key.NewBinding(key.WithKeys("["), key.WithHelp("[", "prev page")),
	}
}
