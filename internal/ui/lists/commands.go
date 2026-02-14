package lists

import (
	"context"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"hardcover-tui/internal/api/mutations"
	"hardcover-tui/internal/api/queries"
)

func (m *Model) loadLists() tea.Cmd {
	client := m.client
	user := m.user
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		lists, err := queries.GetLists(ctx, client, user.ID)
		return listsLoadedMsg{lists: lists, err: err}
	}
}

func (m *Model) loadListBooks(listID int) tea.Cmd {
	client := m.client
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		books, err := queries.GetListBooks(ctx, client, listID)
		return listBooksLoadedMsg{books: books, err: err}
	}
}

func (m *Model) createList(name string) tea.Cmd {
	client := m.client
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		list, err := mutations.InsertList(ctx, client, name, "")
		return listCreatedMsg{list: list, err: err}
	}
}

func (m *Model) deleteList(id int) tea.Cmd {
	client := m.client
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		err := mutations.DeleteList(ctx, client, id)
		return listDeletedMsg{err: err}
	}
}

func (m *Model) doSearchBooks(query string) tea.Cmd {
	client := m.client
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		books, err := queries.Search(ctx, client, query)
		return searchBooksMsg{books: books, err: err}
	}
}

func (m *Model) addBookToSelectedList(listID, bookID int) tea.Cmd {
	client := m.client
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		err := mutations.InsertListBook(ctx, client, listID, bookID)
		return bookAddedToListMsg{err: err}
	}
}

func (m *Model) removeBookFromList(listBookID int) tea.Cmd {
	client := m.client
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		err := mutations.DeleteListBook(ctx, client, listBookID)
		return bookRemovedFromListMsg{err: err}
	}
}

func (m *Model) updateListPrivacy(listID int, name, description string, privacySettingID int) tea.Cmd {
	client := m.client
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		err := mutations.UpdateList(ctx, client, listID, name, description, privacySettingID)
		return privacyUpdatedMsg{err: err}
	}
}
