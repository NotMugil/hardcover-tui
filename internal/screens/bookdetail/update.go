package bookdetail

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/NotMugil/hardcover-tui/internal/api"
	"github.com/NotMugil/hardcover-tui/internal/commands"
	"github.com/NotMugil/hardcover-tui/internal/common"
	"github.com/NotMugil/hardcover-tui/internal/components"
)

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case commands.BookLoadedMsg:
		m.loading = false
		if msg.Err != nil {
			m.err = msg.Err
			m.coverLoading = false
			m.tagsLoading = false
			m.reviewsLoading = false
			return m, nil
		}
		m.userBook = msg.UserBook
		var cmds []tea.Cmd
		if msg.UserBook != nil {
			if msg.UserBook.Book.CoverURL() != "" {
				cmds = append(cmds, commands.LoadCover(msg.UserBook.Book.CoverURL(), 32, 16))
			} else {
				m.coverLoading = false
			}
			cmds = append(cmds, commands.LoadTags(m.deps, msg.UserBook.Book.ID))
			cmds = append(cmds, commands.LoadReviews(m.deps, msg.UserBook.Book.ID, 5))
		} else {
			m.coverLoading = false
			m.tagsLoading = false
			m.reviewsLoading = false
		}
		if len(cmds) > 0 {
			return m, tea.Batch(cmds...)
		}
		return m, nil

	case commands.BookFromBookIDMsg:
		m.loading = false
		if msg.Err != nil {
			m.err = msg.Err
			m.coverLoading = false
			m.tagsLoading = false
			m.reviewsLoading = false
			return m, nil
		}
		m.book = msg.Book
		if msg.UserBook != nil {
			m.userBook = msg.UserBook
		}
		var coverURL string
		var bid int
		if msg.UserBook != nil {
			coverURL = msg.UserBook.Book.CoverURL()
			bid = msg.UserBook.Book.ID
		} else if msg.Book != nil {
			coverURL = msg.Book.CoverURL()
			bid = msg.Book.ID
		}
		var cmds []tea.Cmd
		if coverURL != "" {
			cmds = append(cmds, commands.LoadCover(coverURL, 32, 16))
		} else {
			m.coverLoading = false
		}
		if bid > 0 {
			cmds = append(cmds, commands.LoadTags(m.deps, bid))
			cmds = append(cmds, commands.LoadReviews(m.deps, bid, 5))
		} else {
			m.tagsLoading = false
			m.reviewsLoading = false
		}
		if len(cmds) > 0 {
			return m, tea.Batch(cmds...)
		}
		return m, nil

	case commands.CoverLoadedMsg:
		m.coverArt = msg.Art
		m.coverLoading = false
		return m, nil

	case commands.TagsLoadedMsg:
		if msg.Err == nil {
			m.genres = deduplicateTags(msg.Genres)
		}
		m.tagsLoading = false
		return m, nil

	case commands.ReviewsLoadedMsg:
		if msg.Err == nil {
			m.reviews = msg.Reviews
			items := make([]list.Item, len(m.reviews))
			for i, r := range m.reviews {
				items[i] = reviewItem{data: r}
			}
			m.reviewList.SetItems(items)
		}
		m.reviewsLoading = false
		return m, nil

	case commands.StatusUpdatedMsg:
		m.loading = false
		m.mode = modeDetail
		if msg.Err != nil {
			m.err = msg.Err
			return m, components.NotifyCmd(components.NotifyError, msg.Err.Error())
		}
		if m.userBook != nil {
			m.bookID = m.userBook.ID
			m.loading = true
			return m, tea.Batch(m.spinner.Tick, commands.LoadBookByPK(m.deps, m.bookID), components.NotifyCmd(components.NotifySuccess, "Status updated"))
		}
		return m, components.NotifyCmd(components.NotifySuccess, "Status updated")

	case commands.RatingUpdatedMsg:
		m.loading = false
		m.mode = modeDetail
		if msg.Err != nil {
			m.err = msg.Err
			return m, components.NotifyCmd(components.NotifyError, msg.Err.Error())
		}
		if m.userBook != nil {
			m.bookID = m.userBook.ID
			m.loading = true
			return m, tea.Batch(m.spinner.Tick, commands.LoadBookByPK(m.deps, m.bookID), components.NotifyCmd(components.NotifySuccess, "Rating updated"))
		}
		return m, components.NotifyCmd(components.NotifySuccess, "Rating updated")

	case commands.BookAddedMsg:
		m.loading = false
		if msg.Err != nil {
			m.err = msg.Err
			return m, components.NotifyCmd(components.NotifyError, msg.Err.Error())
		}
		m.userBook = msg.UserBook
		return m, components.NotifyCmd(components.NotifySuccess, "Book added to library")

	case spinner.TickMsg:
		if m.loading || m.journalLoading {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}

	case commands.JournalsLoadedMsg:
		m.journalLoading = false
		if msg.Err != nil {
			m.journalErr = msg.Err
			return m, nil
		}
		m.journals = msg.Journals
		items := make([]list.Item, len(m.journals))
		for i, j := range m.journals {
			items[i] = journalItem{data: j}
		}
		m.journalList.SetItems(items)
		return m, nil

	case commands.JournalSavedMsg:
		m.journalLoading = false
		if msg.Err != nil {
			m.journalErr = msg.Err
			return m, components.NotifyCmd(components.NotifyError, msg.Err.Error())
		}
		m.journalSuccess = true
		m.journalTA.SetValue("")
		m.mode = modeJournal
		m.journalTA.Blur()
		return m, tea.Batch(commands.LoadJournals(m.deps, m.deps.User.ID, 20), components.NotifyCmd(components.NotifySuccess, "Journal entry saved"))

	case commands.JournalDeletedMsg:
		m.journalLoading = false
		if msg.Err != nil {
			m.journalErr = msg.Err
			return m, components.NotifyCmd(components.NotifyError, msg.Err.Error())
		}
		return m, tea.Batch(commands.LoadJournals(m.deps, m.deps.User.ID, 20), components.NotifyCmd(components.NotifySuccess, "Journal entry deleted"))

	case commands.UserListsLoadedMsg:
		m.listLoading = false
		if msg.Err != nil {
			m.listErr = msg.Err
			return m, nil
		}
		m.userLists = msg.Lists
		m.listCursor = 0
		m.mode = modeListSelect
		return m, nil

	case commands.BookAddedToListMsg:
		m.listLoading = false
		if msg.Err != nil {
			m.listErr = msg.Err
			return m, components.NotifyCmd(components.NotifyError, msg.Err.Error())
		}
		m.listSuccess = true
		m.mode = modeDetail
		return m, components.NotifyCmd(components.NotifySuccess, "Added to list")

	case commands.BookRemovedFromListMsg:
		m.loading = false
		if msg.Err != nil {
			m.err = msg.Err
			return m, components.NotifyCmd(components.NotifyError, msg.Err.Error())
		}
		if m.listIndex < len(m.listBooks) {
			m.listBooks = append(m.listBooks[:m.listIndex], m.listBooks[m.listIndex+1:]...)
		}
		if len(m.listBooks) == 0 {
			return m, components.NotifyCmd(components.NotifySuccess, "Removed from "+m.listName)
		}
		if m.listIndex >= len(m.listBooks) {
			m.listIndex = len(m.listBooks) - 1
		}
		m.switchToListBook(m.listIndex)
		return m, tea.Batch(m.spinner.Tick, commands.LoadBookByBookID(m.deps, m.bookID), components.NotifyCmd(components.NotifySuccess, "Removed from "+m.listName))

	case tea.KeyMsg:
		if m.loading {
			return m, nil
		}

		switch m.mode {
		case modeConfirm:
			confirmed, _ := m.confirm.HandleKey(msg.String())
			if !m.confirm.Active {
				if confirmed {
					switch m.confirm.Action {
					case "delete-journal":
						m.mode = modeJournal
						m.journalLoading = true
						return m, tea.Batch(m.spinner.Tick, commands.DeleteJournalAndReload(m.deps, m.confirmItemID))
					case "remove-from-list":
						m.mode = modeDetail
						m.loading = true
						bookID := m.bookID
						if m.book != nil {
							bookID = m.book.ID
						} else if m.userBook != nil {
							bookID = m.userBook.BookID
						}
						return m, tea.Batch(m.spinner.Tick, commands.RemoveBookFromList(m.deps, m.listID, bookID))
					}
				} else {
					m.mode = m.confirmReturn
				}
			}
			return m, nil
		case modeStatusSelect:
			return m.updateStatusSelect(msg)
		case modeRatingSelect:
			return m.updateRatingSelect(msg)
		case modeJournal:
			return m.updateJournal(msg)
		case modeJournalWrite:
			return m.updateJournalWrite(msg)
		case modeReviewRead:
			return m.updateReviewRead(msg)
		case modeListSelect:
			return m.updateListSelect(msg)
		default:
			return m.updateDetail(msg)
		}
	}
	return m, nil
}

func (m *Model) updateDetail(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.reviewMode {
		k := strings.ToLower(msg.String())
		switch k {
		case "esc", "v":
			m.reviewMode = false
			return m, nil
		case "enter":
			if item, ok := m.reviewList.SelectedItem().(reviewItem); ok {
				m.selectedReview = &item.data
				m.mode = modeReviewRead
				m.reviewViewport = viewport.New(m.getWidth()-10, m.height-12)
				m.reviewViewport.Style = common.ValueStyle
				var content strings.Builder
				content.WriteString(common.LabelStyle.Render("@" + item.data.User.Username))
				if item.data.Rating != nil {
					content.WriteString("  " + common.RenderRatingBar(*item.data.Rating, 15))
				}
				if item.data.LikesCount > 0 {
					content.WriteString(common.ValueStyle.Render(fmt.Sprintf("  %d likes", item.data.LikesCount)))
				}
				content.WriteString("\n\n")
				if item.data.Review != nil && *item.data.Review != "" {
					text := stripHTML(*item.data.Review)
					if item.data.ReviewHasSpoilers {
						text = "[SPOILER]\n\n" + text
					}
					wrapped := lipgloss.NewStyle().Width(m.getWidth() - 14).Render(text)
					content.WriteString(wrapped)
				}
				m.reviewViewport.SetContent(content.String())
				return m, nil
			}
		}
		var cmd tea.Cmd
		m.reviewList, cmd = m.reviewList.Update(msg)
		return m, cmd
	}

	k := strings.ToLower(msg.String())
	switch k {
	case "m":
		m.descExpanded = !m.descExpanded
	case "v":
		if len(m.reviews) > 0 {
			m.reviewMode = true
			return m, nil
		}
	case "s":
		if m.userBook != nil {
			m.mode = modeStatusSelect
			m.cursor = m.userBook.StatusID - 1
		}
	case "r":
		if m.userBook != nil {
			m.mode = modeRatingSelect
			if m.userBook.Rating != nil {
				m.cursor = int(*m.userBook.Rating*2) - 1
			} else {
				m.cursor = 0
			}
		}
	case "w":
		if m.userBook != nil {
			return m, func() tea.Msg {
				return NavigateToReviewMsg{UserBook: m.userBook}
			}
		}
	case "p":
		if m.userBook != nil {
			return m, func() tea.Msg {
				return NavigateToProgressMsg{UserBook: m.userBook}
			}
		}
	case "j":
		if m.userBook != nil && !m.journalLoading {
			m.mode = modeJournal
			m.journalLoading = true
			m.journalErr = nil
			m.journalSuccess = false
			return m, tea.Batch(m.spinner.Tick, commands.LoadJournals(m.deps, m.deps.User.ID, 20))
		}
	case "a":
		if m.userBook == nil && m.mode == modeDetail {
			m.mode = modeStatusSelect
			m.cursor = 0
			return m, nil
		}
	case "l":
		if !m.listLoading && m.mode == modeDetail {
			m.listLoading = true
			m.listSuccess = false
			m.listErr = nil
			return m, tea.Batch(m.spinner.Tick, commands.LoadLists(m.deps, m.deps.User.ID))
		}
	case "x":
		if len(m.listBooks) > 0 && m.listID > 0 {
			bookTitle := ""
			if m.book != nil {
				bookTitle = m.book.Title
			} else if m.userBook != nil {
				bookTitle = m.userBook.Book.Title
			}
			m.confirm = components.NewConfirm(
				fmt.Sprintf("Remove \"%s\" from %s?", bookTitle, m.listName),
				"remove-from-list",
			)
			m.confirmReturn = modeDetail
			m.mode = modeConfirm
			return m, nil
		}
	case "n":
		if len(m.listBooks) > 0 && m.listIndex < len(m.listBooks)-1 {
			m.listIndex++
			m.switchToListBook(m.listIndex)
			return m, tea.Batch(m.spinner.Tick, commands.LoadBookByBookID(m.deps, m.bookID))
		}
	case "shift+n":
		if len(m.listBooks) > 0 && m.listIndex > 0 {
			m.listIndex--
			m.switchToListBook(m.listIndex)
			return m, tea.Batch(m.spinner.Tick, commands.LoadBookByBookID(m.deps, m.bookID))
		}
	}
	return m, nil
}

func (m *Model) updateStatusSelect(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	statuses := api.AllStatuses()
	k := strings.ToLower(msg.String())
	switch k {
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < len(statuses)-1 {
			m.cursor++
		}
	case "enter":
		if m.cursor < len(statuses) {
			selectedStatus := int(statuses[m.cursor])
			if m.userBook == nil {
				bid := m.bookID
				if m.book != nil {
					bid = m.book.ID
				}
				if bid > 0 {
					m.loading = true
					return m, tea.Batch(m.spinner.Tick, commands.AddToLibrary(m.deps, bid, selectedStatus))
				}
			} else {
				m.loading = true
				return m, tea.Batch(m.spinner.Tick, commands.UpdateStatus(m.deps, m.userBook.ID, selectedStatus))
			}
		}
	case "esc":
		m.mode = modeDetail
	}
	return m, nil
}

func (m *Model) updateRatingSelect(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	k := strings.ToLower(msg.String())
	switch k {
	case "up", "k":
		if m.cursor < 9 {
			m.cursor++
		}
	case "down", "j":
		if m.cursor > 0 {
			m.cursor--
		}
	case "enter":
		rating := float64(m.cursor+1) * 0.5
		m.loading = true
		return m, tea.Batch(m.spinner.Tick, commands.UpdateRating(m.deps, m.userBook.ID, rating))
	case "esc":
		m.mode = modeDetail
	}
	return m, nil
}

func (m *Model) updateJournal(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	k := strings.ToLower(msg.String())
	switch k {
	case "esc":
		m.mode = modeDetail
		return m, nil
	case "n":
		m.mode = modeJournalWrite
		m.journalSuccess = false
		m.journalTA.Focus()
		return m, textarea.Blink
	case "d":
		if item, ok := m.journalList.SelectedItem().(journalItem); ok {
			m.confirm = components.NewConfirm("Delete this journal entry?", "delete-journal")
			m.confirmItemID = item.data.ID
			m.confirmReturn = modeJournal
			m.mode = modeConfirm
			return m, nil
		}
	}
	var cmd tea.Cmd
	m.journalList, cmd = m.journalList.Update(msg)
	return m, cmd
}

func (m *Model) updateJournalWrite(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+s":
		entry := strings.TrimSpace(m.journalTA.Value())
		if entry == "" {
			return m, nil
		}
		m.journalLoading = true
		m.journalErr = nil
		m.journalSuccess = false
		return m, tea.Batch(m.spinner.Tick, commands.SaveJournal(m.deps, m.userBook.BookID, entry))
	case "esc":
		m.mode = modeJournal
		m.journalTA.Blur()
		return m, nil
	}
	var cmd tea.Cmd
	m.journalTA, cmd = m.journalTA.Update(msg)
	return m, cmd
}

func (m *Model) updateReviewRead(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	k := strings.ToLower(msg.String())
	switch k {
	case "esc", "q":
		m.mode = modeDetail
		m.selectedReview = nil
		return m, nil
	}
	var cmd tea.Cmd
	m.reviewViewport, cmd = m.reviewViewport.Update(msg)
	return m, cmd
}

func (m *Model) updateListSelect(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	k := strings.ToLower(msg.String())
	switch k {
	case "esc":
		m.mode = modeDetail
		m.listErr = nil
		return m, nil
	case "up", "k":
		if m.listCursor > 0 {
			m.listCursor--
		}
	case "down", "j":
		if m.listCursor < len(m.userLists)-1 {
			m.listCursor++
		}
	case "enter":
		if m.listCursor >= 0 && m.listCursor < len(m.userLists) {
			selected := m.userLists[m.listCursor]
			bid := m.bookID
			if m.book != nil {
				bid = m.book.ID
			} else if m.userBook != nil {
				bid = m.userBook.BookID
			}
			if bid > 0 {
				m.listLoading = true
				return m, tea.Batch(m.spinner.Tick, commands.AddBookToList(m.deps, selected.ID, bid, selected.Name))
			}
		}
	}
	return m, nil
}
