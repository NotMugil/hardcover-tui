package lists

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/NotMugil/hardcover-tui/internal/api"
	"github.com/NotMugil/hardcover-tui/internal/common"
)

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case listSettledMsg:
		if msg.listID == m.pendingListID {
			m.booksLoading = true
			return m, tea.Batch(m.spinner.Tick, m.loadListBooks(msg.listID))
		}
		return m, nil

	case listsLoadedMsg:
		m.loading = false
		if msg.err != nil {
			m.err = msg.err
			return m, nil
		}
		m.lists = msg.lists
		items := make([]list.Item, len(m.lists))
		for i, l := range m.lists {
			items[i] = listItem{data: l}
		}
		m.list.SetItems(items)
		if len(m.lists) > 0 {
			m.booksLoading = true
			return m, m.loadListBooks(m.lists[0].ID)
		}
		return m, nil

	case listBooksLoadedMsg:
		m.booksLoading = false
		if msg.err != nil {
			m.err = msg.err
			return m, nil
		}
		m.listBooks = msg.books
		items := make([]list.Item, len(m.listBooks))
		for i, lb := range m.listBooks {
			items[i] = bookListItem{data: lb}
		}
		m.bookList.SetItems(items)
		return m, nil

	case listCreatedMsg:
		m.loading = false
		m.mode = modeNormal
		m.nameInput.SetValue("")
		m.nameInput.Blur()
		if msg.err != nil {
			m.err = msg.err
			return m, common.NotifyCmd(common.NotifyError, msg.err.Error())
		}
		return m, tea.Batch(m.loadLists(), common.NotifyCmd(common.NotifySuccess, "List created"))

	case listDeletedMsg:
		m.loading = false
		if msg.err != nil {
			m.err = msg.err
			return m, common.NotifyCmd(common.NotifyError, msg.err.Error())
		}
		return m, tea.Batch(m.loadLists(), common.NotifyCmd(common.NotifySuccess, "List deleted"))

	case searchBooksMsg:
		m.searching = false
		if msg.err != nil {
			m.addErr = msg.err
			return m, nil
		}
		m.searchResults = msg.books
		rows := make([]table.Row, len(msg.books))
		for i, b := range msg.books {
			rating := ""
			if b.Rating != nil {
				rating = fmt.Sprintf("%.1f", *b.Rating)
			}
			rows[i] = table.Row{b.Title, b.Authors(), rating}
		}
		m.searchTable.SetRows(rows)
		m.searchInput.Blur()
		m.searchTable.Focus()
		return m, nil

	case bookAddedToListMsg:
		if msg.err != nil {
			m.addErr = msg.err
			return m, common.NotifyCmd(common.NotifyError, msg.err.Error())
		}
		m.addSuccess = true
		if item, ok := m.list.SelectedItem().(listItem); ok {
			m.booksLoading = true
			return m, tea.Batch(m.spinner.Tick, m.loadListBooks(item.data.ID), common.NotifyCmd(common.NotifySuccess, "Book added to list"))
		}
		return m, common.NotifyCmd(common.NotifySuccess, "Book added to list")

	case bookRemovedFromListMsg:
		if msg.err != nil {
			m.err = msg.err
			return m, common.NotifyCmd(common.NotifyError, msg.err.Error())
		}
		if item, ok := m.list.SelectedItem().(listItem); ok {
			m.booksLoading = true
			return m, tea.Batch(m.spinner.Tick, m.loadListBooks(item.data.ID), common.NotifyCmd(common.NotifySuccess, "Book removed from list"))
		}
		return m, common.NotifyCmd(common.NotifySuccess, "Book removed from list")

	case privacyUpdatedMsg:
		if msg.err != nil {
			m.err = msg.err
			return m, common.NotifyCmd(common.NotifyError, msg.err.Error())
		}
		return m, tea.Batch(m.loadLists(), common.NotifyCmd(common.NotifySuccess, "Privacy updated"))

	case spinner.TickMsg:
		if m.loading || m.booksLoading {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}

	case tea.KeyMsg:
		if m.mode == modeCreate {
			switch msg.String() {
			case "enter":
				name := strings.TrimSpace(m.nameInput.Value())
				if name == "" {
					return m, nil
				}
				m.loading = true
				return m, tea.Batch(m.spinner.Tick, m.createList(name))
			case "esc":
				m.mode = modeNormal
				m.nameInput.SetValue("")
				m.nameInput.Blur()
				return m, nil
			}
			var cmd tea.Cmd
			m.nameInput, cmd = m.nameInput.Update(msg)
			return m, cmd
		}

		if m.mode == modePrivacy {
			privacyOptions := api.AllPrivacySettings()
			switch strings.ToLower(msg.String()) {
			case "esc":
				m.mode = modeNormal
				return m, nil
			case "up", "k":
				if m.privacyCursor > 0 {
					m.privacyCursor--
				}
			case "down", "j":
				if m.privacyCursor < len(privacyOptions)-1 {
					m.privacyCursor++
				}
			case "enter":
				if m.privacyCursor >= 0 && m.privacyCursor < len(privacyOptions) {
					selected := int(privacyOptions[m.privacyCursor])
					if item, ok := m.list.SelectedItem().(listItem); ok {
						l := item.data
						desc := ""
						if l.Description != nil {
							desc = *l.Description
						}
						m.mode = modeNormal
						return m, m.updateListPrivacy(l.ID, l.Name, desc, selected)
					}
				}
			}
			return m, nil
		}

		if m.mode == modeAddBook {
			switch msg.String() {
			case "esc":
				m.mode = modeNormal
				m.searchInput.SetValue("")
				m.searchInput.Blur()
				m.searchTable.SetRows([]table.Row{})
				m.searchResults = nil
				m.addSuccess = false
				m.addErr = nil
				return m, nil
			case "enter":
				if m.searchInput.Focused() {
					query := strings.TrimSpace(m.searchInput.Value())
					if query == "" {
						return m, nil
					}
					m.searching = true
					m.addSuccess = false
					m.addErr = nil
					return m, tea.Batch(m.spinner.Tick, m.doSearchBooks(query))
				}
				if len(m.searchResults) > 0 {
					cursor := m.searchTable.Cursor()
					if cursor >= 0 && cursor < len(m.searchResults) {
						book := m.searchResults[cursor]
						if item, ok := m.list.SelectedItem().(listItem); ok {
							m.addSuccess = false
							m.addErr = nil
							return m, m.addBookToSelectedList(item.data.ID, book.ID)
						}
					}
				}
				return m, nil
			case "tab":
				if m.searchInput.Focused() {
					m.searchInput.Blur()
					m.searchTable.Focus()
				} else {
					m.searchTable.Blur()
					m.searchInput.Focus()
					return m, textinput.Blink
				}
				return m, nil
			}
			if m.searchInput.Focused() {
				var cmd tea.Cmd
				m.searchInput, cmd = m.searchInput.Update(msg)
				return m, cmd
			}
			var cmd tea.Cmd
			m.searchTable, cmd = m.searchTable.Update(msg)
			return m, cmd
		}

		if m.mode == modeConfirm {
			confirmed, _ := m.confirm.HandleKey(msg.String())
			if !m.confirm.Active {
				m.mode = modeNormal
				if confirmed {
					switch m.confirm.Action {
					case "delete-list":
						m.loading = true
						return m, tea.Batch(m.spinner.Tick, m.deleteList(m.confirmItemID))
					case "remove-book":
						return m, m.removeBookFromList(m.confirmItemID)
					}
				}
			}
			return m, nil
		}

		if m.loading {
			return m, nil
		}

		if m.list.FilterState() == list.Filtering {
			var cmd tea.Cmd
			m.list, cmd = m.list.Update(msg)
			return m, cmd
		}

		if m.focusRight {
			switch strings.ToLower(msg.String()) {
			case "esc":
				m.focusRight = false
				return m, nil
			case "enter":
				if item, ok := m.bookList.SelectedItem().(bookListItem); ok {
					bookID := item.data.Book.ID
					entries := make([]ListBookEntry, len(m.listBooks))
					selectedIdx := 0
					for i, lb := range m.listBooks {
						entries[i] = ListBookEntry{BookID: lb.BookID, Title: lb.Book.Title}
						if lb.BookID == bookID {
							selectedIdx = i
						}
					}
					listName := "List"
					listID := 0
					if sel, ok := m.list.SelectedItem().(listItem); ok {
						listName = sel.data.Name
						listID = sel.data.ID
					}
					return m, func() tea.Msg {
						return NavigateToBookFromListMsg{
							BookID:    bookID,
							ListBooks: entries,
							ListIndex: selectedIdx,
							ListID:    listID,
							ListName:  listName,
						}
					}
				}
			case "x":
				if item, ok := m.bookList.SelectedItem().(bookListItem); ok {
					m.confirm = common.NewConfirm(
						fmt.Sprintf("Remove \"%s\" from this list?", item.data.Book.Title),
						"remove-book",
					)
					m.confirmItemID = item.data.ID
					m.mode = modeConfirm
					return m, nil
				}
			}
			var cmd tea.Cmd
			m.bookList, cmd = m.bookList.Update(msg)
			return m, cmd
		}

		switch strings.ToLower(msg.String()) {
		case "enter":
			if len(m.listBooks) > 0 {
				m.focusRight = true
				return m, nil
			}
		case "n":
			m.mode = modeCreate
			m.nameInput.Focus()
			return m, textinput.Blink
		case "a":
			m.mode = modeAddBook
			m.addSuccess = false
			m.addErr = nil
			m.searchInput.SetValue("")
			m.searchTable.SetRows([]table.Row{})
			m.searchResults = nil
			m.searchInput.Focus()
			return m, textinput.Blink
		case "d":
			if item, ok := m.list.SelectedItem().(listItem); ok {
				m.confirm = common.NewConfirm(
					fmt.Sprintf("Delete list \"%s\"? This cannot be undone.", item.data.Name),
					"delete-list",
				)
				m.confirmItemID = item.data.ID
				m.mode = modeConfirm
				return m, nil
			}
		case "p":
			if item, ok := m.list.SelectedItem().(listItem); ok {
				m.mode = modePrivacy
				m.privacyCursor = item.data.PrivacySettingID - 1
				if m.privacyCursor < 0 {
					m.privacyCursor = 0
				}
				return m, nil
			}
		case "j", "down":
			items := m.list.Items()
			if len(items) > 0 && m.list.Index() == len(items)-1 {
				m.list.Select(0)
			} else {
				var cmd tea.Cmd
				m.list, cmd = m.list.Update(msg)
				_ = cmd
			}
			if item, ok := m.list.SelectedItem().(listItem); ok {
				m.pendingListID = item.data.ID
				m.booksLoading = true
				m.listBooks = nil
				m.bookList.SetItems(nil)
				return m, tea.Batch(m.spinner.Tick, tea.Tick(250*time.Millisecond, func(t time.Time) tea.Msg {
					return listSettledMsg{listID: item.data.ID}
				}))
			}
			return m, nil
		case "k", "up":
			items := m.list.Items()
			if len(items) > 0 && m.list.Index() == 0 {
				m.list.Select(len(items) - 1)
			} else {
				var cmd tea.Cmd
				m.list, cmd = m.list.Update(msg)
				_ = cmd
			}
			if item, ok := m.list.SelectedItem().(listItem); ok {
				m.pendingListID = item.data.ID
				m.booksLoading = true
				m.listBooks = nil
				m.bookList.SetItems(nil)
				return m, tea.Batch(m.spinner.Tick, tea.Tick(250*time.Millisecond, func(t time.Time) tea.Msg {
					return listSettledMsg{listID: item.data.ID}
				}))
			}
			return m, nil
		}
	}

	if !m.loading && m.mode == modeNormal {
		if m.focusRight {
			var cmd tea.Cmd
			m.bookList, cmd = m.bookList.Update(msg)
			return m, cmd
		}
		var cmd tea.Cmd
		m.list, cmd = m.list.Update(msg)
		return m, cmd
	}

	return m, nil
}
