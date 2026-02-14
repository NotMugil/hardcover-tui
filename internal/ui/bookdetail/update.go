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
	"github.com/NotMugil/hardcover-tui/internal/common"
)

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case bookLoadedMsg:
		m.loading = false
		if msg.err != nil {
			m.err = msg.err
			return m, nil
		}
		m.userBook = msg.userBook
		var cmds []tea.Cmd
		if msg.userBook != nil {
			if msg.userBook.Book.CoverURL() != "" {
				cmds = append(cmds, m.loadCover(msg.userBook.Book.CoverURL()))
			}
			cmds = append(cmds, m.loadTags(msg.userBook.Book.ID))
			cmds = append(cmds, m.loadReviews(msg.userBook.Book.ID))
		}
		if len(cmds) > 0 {
			return m, tea.Batch(cmds...)
		}
		return m, nil

	case bookFromBookIDMsg:
		m.loading = false
		if msg.err != nil {
			m.err = msg.err
			return m, nil
		}
		m.book = msg.book
		if msg.userBook != nil {
			m.userBook = msg.userBook
		}
		var coverURL string
		var bid int
		if msg.userBook != nil {
			coverURL = msg.userBook.Book.CoverURL()
			bid = msg.userBook.Book.ID
		} else if msg.book != nil {
			coverURL = msg.book.CoverURL()
			bid = msg.book.ID
		}
		var cmds []tea.Cmd
		if coverURL != "" {
			cmds = append(cmds, m.loadCover(coverURL))
		}
		if bid > 0 {
			cmds = append(cmds, m.loadTags(bid))
			cmds = append(cmds, m.loadReviews(bid))
		}
		if len(cmds) > 0 {
			return m, tea.Batch(cmds...)
		}
		return m, nil

	case coverLoadedMsg:
		m.coverArt = msg.art
		return m, nil

	case tagsLoadedMsg:
		if msg.err == nil {
			m.genres = deduplicateTags(msg.genres)
		}
		return m, nil

	case reviewsLoadedMsg:
		if msg.err == nil {
			m.reviews = msg.reviews
			items := make([]list.Item, len(m.reviews))
			for i, r := range m.reviews {
				items[i] = reviewItem{data: r}
			}
			m.reviewList.SetItems(items)
		}
		return m, nil

	case statusUpdatedMsg:
		m.loading = false
		m.mode = modeDetail
		if msg.err != nil {
			m.err = msg.err
			return m, common.NotifyCmd(common.NotifyError, msg.err.Error())
		}
		if m.userBook != nil {
			m.bookID = m.userBook.ID
			m.loading = true
			return m, tea.Batch(m.spinner.Tick, m.loadBook(), common.NotifyCmd(common.NotifySuccess, "Status updated"))
		}
		return m, common.NotifyCmd(common.NotifySuccess, "Status updated")

	case ratingUpdatedMsg:
		m.loading = false
		m.mode = modeDetail
		if msg.err != nil {
			m.err = msg.err
			return m, common.NotifyCmd(common.NotifyError, msg.err.Error())
		}
		if m.userBook != nil {
			m.bookID = m.userBook.ID
			m.loading = true
			return m, tea.Batch(m.spinner.Tick, m.loadBook(), common.NotifyCmd(common.NotifySuccess, "Rating updated"))
		}
		return m, common.NotifyCmd(common.NotifySuccess, "Rating updated")

	case bookAddedMsg:
		m.loading = false
		if msg.err != nil {
			m.err = msg.err
			return m, common.NotifyCmd(common.NotifyError, msg.err.Error())
		}
		m.userBook = msg.userBook
		return m, common.NotifyCmd(common.NotifySuccess, "Book added to library")

	case spinner.TickMsg:
		if m.loading || m.journalLoading {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}

	case journalsLoadedMsg:
		m.journalLoading = false
		if msg.err != nil {
			m.journalErr = msg.err
			return m, nil
		}
		m.journals = msg.journals
		items := make([]list.Item, len(m.journals))
		for i, j := range m.journals {
			items[i] = journalItem{data: j}
		}
		m.journalList.SetItems(items)
		return m, nil

	case journalSavedMsg:
		m.journalLoading = false
		if msg.err != nil {
			m.journalErr = msg.err
			return m, common.NotifyCmd(common.NotifyError, msg.err.Error())
		}
		m.journalSuccess = true
		m.journalTA.SetValue("")
		m.mode = modeJournal
		m.journalTA.Blur()
		return m, tea.Batch(m.loadJournals(), common.NotifyCmd(common.NotifySuccess, "Journal entry saved"))

	case journalDeletedMsg:
		m.journalLoading = false
		if msg.err != nil {
			m.journalErr = msg.err
			return m, common.NotifyCmd(common.NotifyError, msg.err.Error())
		}
		return m, tea.Batch(m.loadJournals(), common.NotifyCmd(common.NotifySuccess, "Journal entry deleted"))

	case userListsLoadedMsg:
		m.listLoading = false
		if msg.err != nil {
			m.listErr = msg.err
			return m, nil
		}
		m.userLists = msg.lists
		m.listCursor = 0
		m.mode = modeListSelect
		return m, nil

	case bookAddedToListMsg:
		m.listLoading = false
		if msg.err != nil {
			m.listErr = msg.err
			return m, common.NotifyCmd(common.NotifyError, msg.err.Error())
		}
		m.listSuccess = true
		m.mode = modeDetail
		return m, common.NotifyCmd(common.NotifySuccess, "Added to list")

	case bookRemovedFromListMsg:
		m.loading = false
		if msg.err != nil {
			m.err = msg.err
			return m, common.NotifyCmd(common.NotifyError, msg.err.Error())
		}
		if m.listIndex < len(m.listBooks) {
			m.listBooks = append(m.listBooks[:m.listIndex], m.listBooks[m.listIndex+1:]...)
		}
		if len(m.listBooks) == 0 {
			return m, common.NotifyCmd(common.NotifySuccess, "Removed from "+m.listName)
		}
		if m.listIndex >= len(m.listBooks) {
			m.listIndex = len(m.listBooks) - 1
		}
		m.switchToListBook(m.listIndex)
		return m, tea.Batch(m.spinner.Tick, m.loadBookByBookID(), common.NotifyCmd(common.NotifySuccess, "Removed from "+m.listName))

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
						return m, tea.Batch(m.spinner.Tick, m.deleteJournalEntry(m.confirmItemID))
					case "remove-from-list":
						m.mode = modeDetail
						m.loading = true
						return m, tea.Batch(m.spinner.Tick, m.removeBookFromCurrentList())
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
		switch msg.String() {
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

	switch msg.String() {
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
			return m, tea.Batch(m.spinner.Tick, m.loadJournals())
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
			return m, tea.Batch(m.spinner.Tick, m.loadUserLists())
		}
	case "x":
		if len(m.listBooks) > 0 && m.listID > 0 {
			bookTitle := ""
			if m.book != nil {
				bookTitle = m.book.Title
			} else if m.userBook != nil {
				bookTitle = m.userBook.Book.Title
			}
			m.confirm = common.NewConfirm(
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
			return m, tea.Batch(m.spinner.Tick, m.loadBookByBookID())
		}
	case "N":
		if len(m.listBooks) > 0 && m.listIndex > 0 {
			m.listIndex--
			m.switchToListBook(m.listIndex)
			return m, tea.Batch(m.spinner.Tick, m.loadBookByBookID())
		}
	}
	return m, nil
}

func (m *Model) updateStatusSelect(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	statuses := api.AllStatuses()
	switch msg.String() {
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
					return m, tea.Batch(m.spinner.Tick, m.addToLibrary(bid, selectedStatus))
				}
			} else {
				m.loading = true
				return m, tea.Batch(m.spinner.Tick, m.updateStatus(selectedStatus))
			}
		}
	case "esc":
		m.mode = modeDetail
	}
	return m, nil
}

func (m *Model) updateRatingSelect(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
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
		return m, tea.Batch(m.spinner.Tick, m.updateRating(rating))
	case "esc":
		m.mode = modeDetail
	}
	return m, nil
}

func (m *Model) updateJournal(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
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
			m.confirm = common.NewConfirm("Delete this journal entry?", "delete-journal")
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
		return m, tea.Batch(m.spinner.Tick, m.saveJournalEntry(entry))
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
	switch msg.String() {
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
	switch msg.String() {
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
				return m, tea.Batch(m.spinner.Tick, m.addBookToList(selected.ID, selected.Name, bid))
			}
		}
	}
	return m, nil
}
