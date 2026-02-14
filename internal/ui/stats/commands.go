package stats

import (
	"context"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/NotMugil/hardcover-tui/internal/api/queries"
)

func (m *Model) loadStats() tea.Cmd {
	client := m.client
	user := m.user
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		goals, _ := queries.GetGoals(ctx, client, user.ID)
		counts, err := queries.GetUserBookStatusCounts(ctx, client, user.ID)
		if err != nil {
			return statsLoadedMsg{err: err}
		}
		userBooks, err := queries.GetUserBooksForStats(ctx, client, user.ID)
		if err != nil {
			return statsLoadedMsg{err: err}
		}
		readingHistory, _ := queries.GetReadingHistory(ctx, client, user.ID)
		return statsLoadedMsg{goals: goals, counts: counts, userBooks: userBooks, readingHistory: readingHistory, err: err}
	}
}
