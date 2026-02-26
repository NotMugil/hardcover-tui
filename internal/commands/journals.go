package commands

import (
	"context"
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/NotMugil/hardcover-tui/internal/api"
	"github.com/NotMugil/hardcover-tui/internal/api/mutations"
	"github.com/NotMugil/hardcover-tui/internal/api/queries"
)

// LoadJournals fetches reading journal entries.
func LoadJournals(d Deps, userID, limit int) tea.Cmd {
	return Run(
		func(ctx context.Context) ([]api.ReadingJournal, error) {
			return queries.GetReadingJournals(ctx, d.Client, userID, limit)
		},
		func(journals []api.ReadingJournal, err error) tea.Msg {
			return JournalsLoadedMsg{Journals: journals, Err: err}
		},
	)
}

// SaveJournal creates a new journal entry.
func SaveJournal(d Deps, bookID int, entry string) tea.Cmd {
	return func() tea.Msg {
		if bookID == 0 {
			return JournalSavedMsg{Err: fmt.Errorf("no book selected")}
		}
		ctx, cancel := makeCtx()
		defer cancel()
		now := time.Now().Format("2006-01-02")
		err := mutations.InsertReadingJournal(ctx, d.Client, bookID, "note", entry, now)
		return JournalSavedMsg{Err: err}
	}
}

// DeleteJournal deletes a journal entry.
func DeleteJournal(d Deps, journalID int) tea.Cmd {
	return RunVoid(
		func(ctx context.Context) error {
			return mutations.DeleteReadingJournal(ctx, d.Client, journalID)
		},
		func(err error) tea.Msg { return JournalDeletedMsg{Err: err} },
	)
}
