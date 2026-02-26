package commands

import (
	"context"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/NotMugil/hardcover-tui/internal/api"
	"github.com/NotMugil/hardcover-tui/internal/api/mutations"
	"github.com/NotMugil/hardcover-tui/internal/api/queries"
)

// LoadLists fetches all user lists.
func LoadLists(d Deps, userID int) tea.Cmd {
	return Run(
		func(ctx context.Context) ([]api.List, error) {
			return queries.GetLists(ctx, d.Client, userID)
		},
		func(lists []api.List, err error) tea.Msg {
			return UserListsLoadedMsg{Lists: lists, Err: err}
		},
	)
}

// LoadListBooks fetches books in a specific list.
func LoadListBooks(d Deps, listID int) tea.Cmd {
	return Run(
		func(ctx context.Context) ([]api.ListBook, error) {
			return queries.GetListBooks(ctx, d.Client, listID)
		},
		func(books []api.ListBook, err error) tea.Msg {
			return ListBooksLoadedMsg{Books: books, Err: err}
		},
	)
}

// CreateList creates a new list.
func CreateList(d Deps, name string) tea.Cmd {
	return Run(
		func(ctx context.Context) (*api.List, error) {
			return mutations.InsertList(ctx, d.Client, name, "")
		},
		func(list *api.List, err error) tea.Msg {
			return ListCreatedMsg{List: list, Err: err}
		},
	)
}

// DeleteList removes a list.
func DeleteList(d Deps, listID int) tea.Cmd {
	return RunVoid(
		func(ctx context.Context) error {
			return mutations.DeleteList(ctx, d.Client, listID)
		},
		func(err error) tea.Msg { return ListDeletedMsg{Err: err} },
	)
}

// AddBookToList adds a book to a list.
func AddBookToList(d Deps, listID, bookID int, listName string) tea.Cmd {
	return RunVoid(
		func(ctx context.Context) error {
			return mutations.InsertListBook(ctx, d.Client, listID, bookID)
		},
		func(err error) tea.Msg { return BookAddedToListMsg{ListName: listName, Err: err} },
	)
}

// RemoveListBook removes a book from a list by list_book ID.
func RemoveListBook(d Deps, listBookID int) tea.Cmd {
	return RunVoid(
		func(ctx context.Context) error {
			return mutations.DeleteListBook(ctx, d.Client, listBookID)
		},
		func(err error) tea.Msg { return BookRemovedFromListMsg{Err: err} },
	)
}

// UpdateListPrivacy updates a list's name, description, and privacy setting.
func UpdateListPrivacy(d Deps, listID int, name, description string, privacySettingID int) tea.Cmd {
	return RunVoid(
		func(ctx context.Context) error {
			return mutations.UpdateList(ctx, d.Client, listID, name, description, privacySettingID)
		},
		func(err error) tea.Msg { return PrivacyUpdatedMsg{Err: err} },
	)
}
