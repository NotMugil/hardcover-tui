package bookdetail

import (
	"context"
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/NotMugil/hardcover-tui/internal/api/mutations"
	"github.com/NotMugil/hardcover-tui/internal/api/queries"
	"github.com/NotMugil/hardcover-tui/internal/common"
)

func (m *Model) loadBook() tea.Cmd {
	client := m.client
	id := m.bookID
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		ub, err := queries.GetUserBookByPK(ctx, client, id)
		return bookLoadedMsg{userBook: ub, err: err}
	}
}

func (m *Model) loadBookByBookID() tea.Cmd {
	client := m.client
	user := m.user
	bookID := m.bookID
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		book, err := queries.GetBookByID(ctx, client, bookID)
		if err != nil {
			return bookFromBookIDMsg{err: err}
		}
		ub, _ := queries.GetUserBookByBookID(ctx, client, user.ID, bookID)
		return bookFromBookIDMsg{book: book, userBook: ub}
	}
}

func (m *Model) loadCover(url string) tea.Cmd {
	return func() tea.Msg {
		art, err := common.RenderImage(url, 32, 16)
		if err != nil {
			return coverLoadedMsg{}
		}
		return coverLoadedMsg{art: art}
	}
}

func (m *Model) loadTags(bookID int) tea.Cmd {
	client := m.client
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		genres, _, _, err := queries.GetBookTags(ctx, client, bookID)
		return tagsLoadedMsg{genres: genres, err: err}
	}
}

func (m *Model) loadReviews(bookID int) tea.Cmd {
	client := m.client
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		reviews, err := queries.GetBookReviews(ctx, client, bookID, 5)
		return reviewsLoadedMsg{reviews: reviews, err: err}
	}
}

func (m *Model) updateStatus(statusID int) tea.Cmd {
	client := m.client
	ubID := m.userBook.ID
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		err := mutations.UpdateUserBookStatus(ctx, client, ubID, statusID)
		return statusUpdatedMsg{err: err}
	}
}

func (m *Model) updateRating(rating float64) tea.Cmd {
	client := m.client
	ubID := m.userBook.ID
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		err := mutations.UpdateUserBookRating(ctx, client, ubID, rating)
		return ratingUpdatedMsg{err: err}
	}
}

func (m *Model) addToLibrary(bookID int, statusID int) tea.Cmd {
	client := m.client
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		ub, err := mutations.InsertUserBook(ctx, client, bookID, statusID)
		return bookAddedMsg{userBook: ub, err: err}
	}
}

func (m *Model) loadJournals() tea.Cmd {
	client := m.client
	user := m.user
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		journals, err := queries.GetReadingJournals(ctx, client, user.ID, 20)
		return journalsLoadedMsg{journals: journals, err: err}
	}
}

func (m *Model) saveJournalEntry(entry string) tea.Cmd {
	client := m.client
	ub := m.userBook
	return func() tea.Msg {
		if ub == nil {
			return journalSavedMsg{err: fmt.Errorf("no book selected")}
		}
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		now := time.Now().Format("2006-01-02")
		err := mutations.InsertReadingJournal(ctx, client, ub.BookID, "note", entry, now)
		return journalSavedMsg{err: err}
	}
}

func (m *Model) deleteJournalEntry(id int) tea.Cmd {
	client := m.client
	user := m.user
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		err := mutations.DeleteReadingJournal(ctx, client, id)
		if err != nil {
			return journalDeletedMsg{err: err}
		}
		journals, err := queries.GetReadingJournals(ctx, client, user.ID, 20)
		return journalsLoadedMsg{journals: journals, err: err}
	}
}

func (m *Model) loadUserLists() tea.Cmd {
	client := m.client
	user := m.user
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		lists, err := queries.GetLists(ctx, client, user.ID)
		return userListsLoadedMsg{lists: lists, err: err}
	}
}

func (m *Model) addBookToList(listID int, listName string, bookID int) tea.Cmd {
	client := m.client
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		err := mutations.InsertListBook(ctx, client, listID, bookID)
		return bookAddedToListMsg{listName: listName, err: err}
	}
}

func (m *Model) removeBookFromCurrentList() tea.Cmd {
	client := m.client
	listID := m.listID
	bookID := m.bookID
	if m.book != nil {
		bookID = m.book.ID
	} else if m.userBook != nil {
		bookID = m.userBook.BookID
	}
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		books, err := queries.GetListBooks(ctx, client, listID)
		if err != nil {
			return bookRemovedFromListMsg{err: err}
		}
		for _, lb := range books {
			if lb.BookID == bookID {
				err = mutations.DeleteListBook(ctx, client, lb.ID)
				return bookRemovedFromListMsg{err: err}
			}
		}
		return bookRemovedFromListMsg{err: fmt.Errorf("book not found in list")}
	}
}
