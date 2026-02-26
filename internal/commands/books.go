package commands

import (
	"context"
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/NotMugil/hardcover-tui/internal/api"
	"github.com/NotMugil/hardcover-tui/internal/api/mutations"
	"github.com/NotMugil/hardcover-tui/internal/api/queries"
	"github.com/NotMugil/hardcover-tui/internal/common"
)

// LoadBookByPK loads a UserBook by its primary key.
func LoadBookByPK(d Deps, ubID int) tea.Cmd {
	return Run(
		func(ctx context.Context) (*api.UserBook, error) {
			return queries.GetUserBookByPK(ctx, d.Client, ubID)
		},
		func(ub *api.UserBook, err error) tea.Msg {
			return BookLoadedMsg{UserBook: ub, Err: err}
		},
	)
}

// LoadBookByBookID loads a Book by its book ID, plus optionally the user's UserBook.
func LoadBookByBookID(d Deps, bookID int) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := makeCtx()
		defer cancel()
		book, err := queries.GetBookByID(ctx, d.Client, bookID)
		if err != nil {
			return BookFromBookIDMsg{Err: err}
		}
		ub, _ := queries.GetUserBookByBookID(ctx, d.Client, d.User.ID, bookID)
		return BookFromBookIDMsg{Book: book, UserBook: ub}
	}
}

// UpdateStatus changes a book's reading status.
func UpdateStatus(d Deps, ubID, statusID int) tea.Cmd {
	return RunVoid(
		func(ctx context.Context) error {
			return mutations.UpdateUserBookStatus(ctx, d.Client, ubID, statusID)
		},
		func(err error) tea.Msg { return StatusUpdatedMsg{Err: err} },
	)
}

// UpdateRating changes a book's rating.
func UpdateRating(d Deps, ubID int, rating float64) tea.Cmd {
	return RunVoid(
		func(ctx context.Context) error {
			return mutations.UpdateUserBookRating(ctx, d.Client, ubID, rating)
		},
		func(err error) tea.Msg { return RatingUpdatedMsg{Err: err} },
	)
}

// AddToLibrary inserts a new user_book.
func AddToLibrary(d Deps, bookID, statusID int) tea.Cmd {
	return Run(
		func(ctx context.Context) (*api.UserBook, error) {
			return mutations.InsertUserBook(ctx, d.Client, bookID, statusID)
		},
		func(ub *api.UserBook, err error) tea.Msg {
			return BookAddedMsg{UserBook: ub, Err: err}
		},
	)
}

// LoadCover renders a book cover image to terminal art.
func LoadCover(url string, w, h int) tea.Cmd {
	return func() tea.Msg {
		art, err := common.RenderImage(url, w, h)
		if err != nil {
			return CoverLoadedMsg{}
		}
		return CoverLoadedMsg{Art: art}
	}
}

// LoadTags fetches genre tags for a book.
func LoadTags(d Deps, bookID int) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		genres, _, _, err := queries.GetBookTags(ctx, d.Client, bookID)
		return TagsLoadedMsg{Genres: genres, Err: err}
	}
}

// LoadReviews fetches reviews for a book.
func LoadReviews(d Deps, bookID, limit int) tea.Cmd {
	return Run(
		func(ctx context.Context) ([]api.BookReview, error) {
			return queries.GetBookReviews(ctx, d.Client, bookID, limit)
		},
		func(r []api.BookReview, err error) tea.Msg {
			return ReviewsLoadedMsg{Reviews: r, Err: err}
		},
	)
}

// SearchBooks searches for books by query string.
func SearchBooks(d Deps, query string) tea.Cmd {
	return Run(
		func(ctx context.Context) ([]api.Book, error) {
			return queries.Search(ctx, d.Client, query)
		},
		func(books []api.Book, err error) tea.Msg {
			return SearchResultsMsg{Books: books, Err: err}
		},
	)
}

// SaveReview updates the review for a user book.
func SaveReview(d Deps, ubID int, review string, hasSpoilers bool) tea.Cmd {
	return RunVoid(
		func(ctx context.Context) error {
			return mutations.UpdateUserBookReview(ctx, d.Client, ubID, review, hasSpoilers)
		},
		func(err error) tea.Msg { return ReviewSavedMsg{Err: err} },
	)
}

// UpdateProgress updates reading progress.
func UpdateProgress(d Deps, readID int, progressPages *int) tea.Cmd {
	return RunVoid(
		func(ctx context.Context) error {
			return mutations.UpdateUserBookRead(ctx, d.Client, readID, progressPages)
		},
		func(err error) tea.Msg { return ProgressUpdatedMsg{Err: err} },
	)
}

// UpdateProgressDates updates started_at and finished_at on a read entry.
func UpdateProgressDates(d Deps, readID int, startedAt, finishedAt *string) tea.Cmd {
	return RunVoid(
		func(ctx context.Context) error {
			return mutations.UpdateUserBookReadDates(ctx, d.Client, readID, startedAt, finishedAt)
		},
		func(err error) tea.Msg { return ProgressUpdatedMsg{Err: err} },
	)
}

// UpdateProgressWithDates updates reading progress pages and optionally dates.
func UpdateProgressWithDates(d Deps, readID int, pages int, startedAt, finishedAt *string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := makeCtx()
		defer cancel()
		err := mutations.UpdateUserBookRead(ctx, d.Client, readID, &pages)
		if err != nil {
			return ProgressUpdatedMsg{Err: err}
		}
		if startedAt != nil || finishedAt != nil {
			err = mutations.UpdateUserBookReadDates(ctx, d.Client, readID, startedAt, finishedAt)
		}
		return ProgressUpdatedMsg{Err: err}
	}
}

// UpdateProfile updates the user's profile fields.
func UpdateProfile(d Deps, name, bio, location string) tea.Cmd {
	return RunVoid(
		func(ctx context.Context) error {
			return mutations.UpdateUserProfile(ctx, d.Client, name, bio, location)
		},
		func(err error) tea.Msg { return ProfileUpdatedMsg{Err: err} },
	)
}

// DeleteJournalAndReload deletes a journal entry and reloads the list.
func DeleteJournalAndReload(d Deps, journalID int) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := makeCtx()
		defer cancel()
		err := mutations.DeleteReadingJournal(ctx, d.Client, journalID)
		if err != nil {
			return JournalDeletedMsg{Err: err}
		}
		journals, err := queries.GetReadingJournals(ctx, d.Client, d.User.ID, 20)
		return JournalsLoadedMsg{Journals: journals, Err: err}
	}
}

// RemoveBookFromList finds and removes a book from a list by book ID.
func RemoveBookFromList(d Deps, listID, bookID int) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := makeCtx()
		defer cancel()
		books, err := queries.GetListBooks(ctx, d.Client, listID)
		if err != nil {
			return BookRemovedFromListMsg{Err: err}
		}
		for _, lb := range books {
			if lb.BookID == bookID {
				err = mutations.DeleteListBook(ctx, d.Client, lb.ID)
				return BookRemovedFromListMsg{Err: err}
			}
		}
		return BookRemovedFromListMsg{Err: fmt.Errorf("book not found in list")}
	}
}
