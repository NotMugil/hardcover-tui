package bookdetail

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/lipgloss"
	overlay "github.com/rmhubbert/bubbletea-overlay"

	"github.com/NotMugil/hardcover-tui/internal/api"
	"github.com/NotMugil/hardcover-tui/internal/common"
)

func (m *Model) renderListOverlay(maxW int) string {
	w := 50
	if w > maxW-4 {
		w = maxW - 4
	}

	var sel strings.Builder
	if m.listLoading {
		sel.WriteString(fmt.Sprintf("  %s Loading lists...\n", m.spinner.View()))
		return common.RenderActivePanel("Add to List", sel.String(), w)
	}
	if m.listErr != nil {
		sel.WriteString(common.ErrorStyle.Render("Error: " + m.listErr.Error()))
		sel.WriteString("\n\n")
		sel.WriteString(common.HelpStyle.Render("esc: close"))
		return common.RenderActivePanel("Add to List", sel.String(), w)
	}

	sel.WriteString(common.LabelStyle.Render("Select a list:"))
	sel.WriteString("\n\n")
	for i, l := range m.userLists {
		cursor := "  "
		if i == m.listCursor {
			cursor = "> "
		}
		name := l.Name
		count := fmt.Sprintf(" (%d books)", l.BooksCount)
		sel.WriteString(fmt.Sprintf("%s%s%s\n", cursor, common.ValueStyle.Render(name), common.HelpStyle.Render(count)))
	}
	if len(m.userLists) == 0 {
		sel.WriteString(common.ValueStyle.Render("  No lists found. Create one in the Lists tab."))
		sel.WriteString("\n")
	}
	sel.WriteString("\n")
	sel.WriteString(common.HelpStyle.Render("enter: add | esc: cancel"))

	return common.RenderActivePanel("Add to List", sel.String(), w)
}

func (m *Model) View() string {
	if m.loading {
		return common.AppStyle.Render(
			fmt.Sprintf("\n  %s Loading book...\n", m.spinner.View()),
		)
	}

	if m.err != nil {
		return common.AppStyle.Render(
			common.ErrorStyle.Render("Error: " + m.err.Error()),
		)
	}

	if m.userBook == nil && m.book == nil {
		return common.AppStyle.Render(
			common.ValueStyle.Render("No book data available"),
		)
	}

	var book api.Book
	if m.userBook != nil {
		book = m.userBook.Book
	} else if m.book != nil {
		book = *m.book
	}

	fbW := m.getWidth() - 2
	if fbW < 60 {
		fbW = 80
	}
	m.flexBox.SetWidth(fbW)
	m.flexBox.SetHeight(m.height)
	m.flexBox.ForceRecalculate()

	leftCell := m.flexBox.GetRow(0).GetCell(0)
	rightCell := m.flexBox.GetRow(0).GetCell(1)
	leftW := leftCell.GetWidth()
	rightW := rightCell.GetWidth()

	var leftPanels []string

	{
		panelInnerW := leftW - 4 // panel borders + padding
		coverW := 0
		if m.coverArt != "" {
			coverW = lipgloss.Width(m.coverArt) + 1 // +1 for spacer
		}
		textW := panelInnerW - coverW
		if textW < 10 {
			textW = 10
		}

		var details strings.Builder
		titleWrapped := lipgloss.NewStyle().Width(textW).Render(book.Title)
		details.WriteString(common.TitleStyle.Render(titleWrapped))
		details.WriteString("\n")
		if book.Subtitle != nil && *book.Subtitle != "" {
			subtitleWrapped := lipgloss.NewStyle().Width(textW).Render(*book.Subtitle)
			details.WriteString(common.SubtitleStyle.Render(subtitleWrapped))
			details.WriteString("\n")
		}
		authorWrapped := lipgloss.NewStyle().Width(textW).Render("by " + book.Authors())
		details.WriteString(common.ValueStyle.Render(authorWrapped))
		if book.Pages != nil {
			details.WriteString("\n" + common.ValueStyle.Render(fmt.Sprintf("%d pages", *book.Pages)))
		}
		if book.ReleaseYear != nil {
			details.WriteString("\n" + common.ValueStyle.Render(fmt.Sprintf("Published: %d", *book.ReleaseYear)))
		}

		var bookContent string
		if m.coverArt != "" {
			coverView := m.coverArt
			detailsView := lipgloss.NewStyle().Width(textW).Render(details.String())
			bookContent = lipgloss.JoinHorizontal(lipgloss.Top, coverView, " ", detailsView)
		} else {
			bookContent = details.String()
		}

		if len(m.listBooks) > 0 {
			listInfo := fmt.Sprintf("from %s %d of %d",
				common.LabelStyle.Render(m.listName),
				m.listIndex+1, len(m.listBooks))
			bookContent += "\n" + common.HelpStyle.Render(listInfo)
		}

		leftPanels = append(leftPanels, common.RenderPanel("Book", bookContent, leftW))
	}

	{
		var s strings.Builder
		if m.userBook != nil {
			status := api.StatusID(m.userBook.StatusID)
			statusStyle := lipglossWithFg(common.StatusColor(m.userBook.StatusID))
			s.WriteString(statusStyle.Render(status.String()))
		} else {
			s.WriteString(common.ValueStyle.Render("Not in library"))
		}
		leftPanels = append(leftPanels, common.RenderPanel("Status", s.String(), leftW))
	}

	if m.listSuccess {
		leftPanels = append(leftPanels, common.SuccessStyle.Render("Added to list!"))
	}

	{
		genres := m.genres
		if len(genres) == 0 {
			genres = deduplicateTags(book.Genres)
		}
		if len(genres) > 5 {
			genres = genres[:5]
		}
		if len(genres) > 0 {
			var names []string
			for _, g := range genres {
				names = append(names, g.Name)
			}
			leftPanels = append(leftPanels, common.RenderPanel("Genres", common.ValueStyle.Render(strings.Join(names, ", ")), leftW))
		}
	}

	leftCol := lipgloss.JoinVertical(lipgloss.Left, leftPanels...)
	leftCell.SetContent(leftCol)

	var rightPanels []string
	rightInner := rightW - 4

	if m.mode == modeJournal || m.mode == modeJournalWrite {
		if m.mode == modeJournalWrite {
			var write strings.Builder
			write.WriteString(common.LabelStyle.Render("New Journal Entry"))
			write.WriteString("\n\n")
			write.WriteString(m.journalTA.View())
			write.WriteString("\n\n")
			write.WriteString(common.HelpStyle.Render("ctrl+s: save | esc: cancel"))
			rightPanels = append(rightPanels, common.RenderActivePanel("Write Entry", write.String(), rightW))
		}
		if m.journalErr != nil {
			rightPanels = append(rightPanels, common.ErrorStyle.Render("Error: "+m.journalErr.Error()))
		}
		if m.journalSuccess {
			rightPanels = append(rightPanels, common.SuccessStyle.Render("Entry saved!"))
		}
		if m.journalLoading {
			rightPanels = append(rightPanels, common.RenderPanel("Journal",
				fmt.Sprintf("  %s Loading...\n", m.spinner.View()), rightW))
		} else if m.mode == modeJournal {
			m.journalList.SetSize(rightInner, m.height-12)
			rightPanels = append(rightPanels, common.RenderActivePanel("Journal", m.journalList.View(), rightW))
			rightPanels = append(rightPanels, common.RenderPanel("Help",
				common.HelpStyle.Render("n: new entry | d: delete | j/k: navigate | esc: back"), rightW))
		}
	} else {
		if book.Description != nil && *book.Description != "" {
			desc := *book.Description
			wrapped := lipgloss.NewStyle().Width(rightInner).Render(desc)
			lines := strings.Split(wrapped, "\n")
			maxLines := 8
			if m.descExpanded || len(lines) <= maxLines {
				hint := ""
				if len(lines) > maxLines {
					hint = "\n" + common.HelpStyle.Render("[m] read less")
				}
				rightPanels = append(rightPanels, common.RenderPanel("Description",
					common.ValueStyle.Render(wrapped)+hint, rightW))
			} else {
				truncated := strings.Join(lines[:maxLines], "\n")
				hint := "\n" + common.HelpStyle.Render("[m] read more...")
				rightPanels = append(rightPanels, common.RenderPanel("Description",
					common.ValueStyle.Render(truncated)+hint, rightW))
			}
		}

		{
			communityW := rightW / 2
			personalW := rightW - communityW

			var community strings.Builder
			if book.Rating != nil {
				ratingText := fmt.Sprintf("%.1f/5", *book.Rating)
				community.WriteString(common.ValueStyle.Render(ratingText))
				countsText := fmt.Sprintf(" (%d ratings)", book.RatingsCount)
				community.WriteString(lipgloss.NewStyle().Foreground(common.ColorSubtext).Render(countsText))
			} else {
				community.WriteString(common.ValueStyle.Render("No ratings yet"))
			}
			communityPanel := common.RenderPanel("Community", community.String(), communityW)

			var personal strings.Builder
			if m.userBook != nil && m.userBook.Rating != nil {
				ratingText := fmt.Sprintf("%.1f/5", *m.userBook.Rating)
				personal.WriteString(common.ValueStyle.Render(ratingText))
			} else {
				personal.WriteString(common.ValueStyle.Render("N/A"))
			}
			personalPanel := common.RenderPanel("Your Rating", personal.String(), personalW)

			ratingRow := lipgloss.JoinHorizontal(lipgloss.Top, communityPanel, personalPanel)
			rightPanels = append(rightPanels, ratingRow)
		}

		if len(m.reviews) > 0 {
			if m.reviewMode {
				reviewH := m.height - lipgloss.Height(lipgloss.JoinVertical(lipgloss.Left, rightPanels...)) - 6
				if reviewH < 8 {
					reviewH = 8
				}
				m.reviewList.SetSize(rightInner, reviewH)
				rightPanels = append(rightPanels,
					common.RenderActivePanel("Popular Reviews (v/esc to close)",
						m.reviewList.View(), rightW))
			} else {
				usedH := lipgloss.Height(lipgloss.JoinVertical(lipgloss.Left, rightPanels...))
				remainH := m.height - usedH - 4 // leave room for panel borders + help
				limit := remainH / 5
				if limit < 2 {
					limit = 2
				}
				if limit > len(m.reviews) {
					limit = len(m.reviews)
				}
				var rev strings.Builder
				for i := 0; i < limit; i++ {
					r := m.reviews[i]
					if i > 0 {
						rev.WriteString("\n" + common.ValueStyle.Render(strings.Repeat("-", rightInner)) + "\n")
					}
					header := common.LabelStyle.Render("@" + r.User.Username)
					if r.Rating != nil {
						header += "  " + common.RenderRatingBar(*r.Rating, 10)
					}
					if r.LikesCount > 0 {
						header += common.ValueStyle.Render(fmt.Sprintf("  %d likes", r.LikesCount))
					}
					rev.WriteString(header)

					if r.Review != nil && *r.Review != "" {
						text := stripHTML(*r.Review)
						if r.ReviewHasSpoilers {
							text = "[spoiler] " + text
						}
						if len(text) > 200 {
							text = text[:197] + "..."
						}
						wrapped := lipgloss.NewStyle().Width(rightInner).Render(text)
						rev.WriteString("\n" + common.ValueStyle.Render(wrapped))
					}
				}
				if len(m.reviews) > limit {
					rev.WriteString("\n" + common.HelpStyle.Render(
						fmt.Sprintf("[v] show all %d reviews", len(m.reviews))))
				}
				rightPanels = append(rightPanels, common.RenderPanel("Popular Reviews", rev.String(), rightW))
			}
		}
	}

	rightCol := lipgloss.JoinVertical(lipgloss.Left, rightPanels...)
	rightCell.SetContent(rightCol)

	body := m.flexBox.Render()

	switch m.mode {
	case modeStatusSelect:
		fg := m.renderStatusOverlay(fbW)
		body = overlay.Composite(fg, body, overlay.Center, overlay.Center, 0, 0)

	case modeRatingSelect:
		fg := m.renderRatingOverlay(fbW)
		body = overlay.Composite(fg, body, overlay.Center, overlay.Center, 0, 0)

	case modeReviewRead:
		fg := m.renderReviewOverlay(fbW)
		body = overlay.Composite(fg, body, overlay.Center, overlay.Center, 0, 0)

	case modeListSelect:
		fg := m.renderListOverlay(fbW)
		body = overlay.Composite(fg, body, overlay.Center, overlay.Center, 0, 0)

	case modeConfirm:
		fg := common.RenderConfirmOverlay(m.confirm.Message, m.confirm.Cursor, 50)
		body = overlay.Composite(fg, body, overlay.Center, overlay.Center, 0, 0)
	}

	return common.AppStyle.Render(body)
}

// renderStatusOverlay renders the status selection as an overlay panel.
func (m *Model) renderStatusOverlay(maxW int) string {
	w := 40
	if w > maxW-4 {
		w = maxW - 4
	}

	var sel strings.Builder
	sel.WriteString(common.LabelStyle.Render("Select Status:"))
	sel.WriteString("\n\n")
	for i, s := range api.AllStatuses() {
		cursor := "  "
		if i == m.cursor {
			cursor = "> "
		}
		sStyle := lipglossWithFg(common.StatusColor(int(s)))
		sel.WriteString(fmt.Sprintf("%s%s\n", cursor, sStyle.Render(s.String())))
	}
	sel.WriteString("\n")
	sel.WriteString(common.HelpStyle.Render("enter: select | esc: cancel"))

	return common.RenderActivePanel("Change Status", sel.String(), w)
}

// renderRatingOverlay renders the rating selection as an overlay panel.
func (m *Model) renderRatingOverlay(maxW int) string {
	w := 40
	if w > maxW-4 {
		w = maxW - 4
	}

	var sel strings.Builder
	sel.WriteString(common.LabelStyle.Render("Select Rating:"))
	sel.WriteString("\n\n")
	for i := 9; i >= 0; i-- {
		rating := float64(i+1) * 0.5
		cursor := "  "
		if i == m.cursor {
			cursor = "> "
		}
		sel.WriteString(fmt.Sprintf("%s%s\n", cursor, common.RenderRatingBar(rating, 15)))
	}
	sel.WriteString("\n")
	sel.WriteString(common.HelpStyle.Render("enter: select | esc: cancel"))

	return common.RenderActivePanel("Rate Book", sel.String(), w)
}

// renderReviewOverlay renders a full review in a viewport overlay panel.
func (m *Model) renderReviewOverlay(maxW int) string {
	w := maxW - 6
	if w < 40 {
		w = 40
	}
	if w > 80 {
		w = 80
	}

	vpH := m.height - 10
	if vpH < 5 {
		vpH = 5
	}
	m.reviewViewport.Width = w - 4
	m.reviewViewport.Height = vpH

	var content strings.Builder
	content.WriteString(m.reviewViewport.View())
	content.WriteString("\n\n")
	content.WriteString(common.HelpStyle.Render("j/k: scroll | esc: close"))

	return common.RenderActivePanel("Review", content.String(), w)
}

// HelpBindings returns page-specific keybindings for the global help bar.
func (m *Model) HelpBindings() []key.Binding {
	switch m.mode {
	case modeJournal:
		return []key.Binding{
			key.NewBinding(key.WithKeys("n"), key.WithHelp("n", "new entry")),
			key.NewBinding(key.WithKeys("d"), key.WithHelp("d", "delete")),
		}
	case modeJournalWrite:
		return []key.Binding{
			key.NewBinding(key.WithKeys("ctrl+s"), key.WithHelp("ctrl+s", "save")),
		}
	case modeStatusSelect, modeRatingSelect, modeListSelect:
		return []key.Binding{
			key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "select")),
		}
	case modeReviewRead:
		return []key.Binding{
			key.NewBinding(key.WithKeys("j"), key.WithHelp("j/k", "scroll")),
		}
	default:
		bindings := []key.Binding{}
		if m.userBook != nil {
			bindings = append(bindings,
				key.NewBinding(key.WithKeys("s"), key.WithHelp("s", "status")),
				key.NewBinding(key.WithKeys("r"), key.WithHelp("r", "rate")),
				key.NewBinding(key.WithKeys("w"), key.WithHelp("w", "review")),
				key.NewBinding(key.WithKeys("p"), key.WithHelp("p", "progress")),
				key.NewBinding(key.WithKeys("j"), key.WithHelp("j", "journal")),
			)
		} else {
			bindings = append(bindings,
				key.NewBinding(key.WithKeys("a"), key.WithHelp("a", "add to library")),
			)
		}
		bindings = append(bindings,
			key.NewBinding(key.WithKeys("l"), key.WithHelp("l", "add to list")),
		)
		if len(m.listBooks) > 0 {
			bindings = append(bindings,
				key.NewBinding(key.WithKeys("x"), key.WithHelp("x", "remove from list")),
				key.NewBinding(key.WithKeys("n"), key.WithHelp("n", "next")),
				key.NewBinding(key.WithKeys("N"), key.WithHelp("N", "prev")),
			)
		}
		return bindings
	}
}

// FullHelpBindings returns extra keybindings only shown in the expanded help view.
func (m *Model) FullHelpBindings() []key.Binding {
	if m.mode == modeDetail {
		return []key.Binding{
			key.NewBinding(key.WithKeys("m"), key.WithHelp("m", "more/less")),
			key.NewBinding(key.WithKeys("v"), key.WithHelp("v", "reviews")),
		}
	}
	return nil
}

// deduplicateTags removes duplicate tags by name, keeping the first occurrence.
func deduplicateTags(tags []api.TagItem) []api.TagItem {
	seen := make(map[string]struct{}, len(tags))
	out := make([]api.TagItem, 0, len(tags))
	for _, t := range tags {
		if _, ok := seen[t.Name]; !ok {
			seen[t.Name] = struct{}{}
			out = append(out, t)
		}
	}
	return out
}

func lipglossWithFg(c lipgloss.Color) lipgloss.Style {
	return lipgloss.NewStyle().Foreground(c)
}
