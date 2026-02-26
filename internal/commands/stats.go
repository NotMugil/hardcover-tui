package commands

import (
	"context"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/NotMugil/hardcover-tui/internal/api"
	"github.com/NotMugil/hardcover-tui/internal/api/queries"
)

// LoadStats fetches all stats data (goals, counts, user books, reading history).
func LoadStats(d Deps) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := makeCtx()
		defer cancel()
		goals, _ := queries.GetGoals(ctx, d.Client, d.User.ID)
		counts, err := queries.GetUserBookStatusCounts(ctx, d.Client, d.User.ID)
		if err != nil {
			return StatsLoadedMsg{Err: err}
		}
		userBooks, err := queries.GetUserBooksForStats(ctx, d.Client, d.User.ID)
		if err != nil {
			return StatsLoadedMsg{Err: err}
		}
		readingHistory, _ := queries.GetReadingHistory(ctx, d.Client, d.User.ID)
		return StatsLoadedMsg{
			Goals:          goals,
			Counts:         counts,
			UserBooks:      userBooks,
			ReadingHistory: readingHistory,
		}
	}
}

// LoadGoals fetches reading goals for a user.
func LoadGoals(d Deps, userID int) tea.Cmd {
	return Run(
		func(ctx context.Context) ([]api.Goal, error) {
			return queries.GetGoals(ctx, d.Client, userID)
		},
		func(goals []api.Goal, err error) tea.Msg {
			return GoalsLoadedMsg{Goals: goals, Err: err}
		},
	)
}
