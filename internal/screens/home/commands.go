package home

import (
	"context"
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/NotMugil/hardcover-tui/internal/api"
	"github.com/NotMugil/hardcover-tui/internal/api/queries"
	"github.com/NotMugil/hardcover-tui/internal/common"
)

func (m *Model) tickClock() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return clockTickMsg(t)
	})
}

func (m *Model) loadInitial() tea.Cmd {
	client := m.deps.Client
	user := m.deps.User
	filter := m.filter
	page := m.page
	pageSize := m.pageSize
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		u, uErr := queries.GetMe(ctx, client)
		if uErr == nil && u != nil {
			user = u
		}

		if user == nil || user.ID <= 0 {
			return booksLoadedMsg{err: fmt.Errorf("user profile not loaded")}
		}

		var statusID *int
		if filter > 0 {
			s := filter
			statusID = &s
		}
		books, err := queries.GetUserBooks(ctx, client, user.ID, statusID, pageSize, page*pageSize)
		if err != nil {
			return booksLoadedMsg{err: err}
		}

		reading, _ := queries.GetCurrentlyReading(ctx, client, user.ID)

		var avatarArt string
		if user.ImageURL() != "" {
			art, artErr := common.RenderImage(user.ImageURL(), 12, 6)
			if artErr == nil {
				avatarArt = art
			}
		}

		return booksLoadedMsg{user: user, books: books, reading: reading, avatarArt: avatarArt}
	}
}

func (m *Model) loadBooksOnly() tea.Cmd {
	client := m.deps.Client
	user := m.deps.User
	filter := m.filter
	page := m.page
	pageSize := m.pageSize
	return func() tea.Msg {
		if user == nil || user.ID <= 0 {
			return booksOnlyLoadedMsg{err: fmt.Errorf("user profile not loaded")}
		}
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		var statusID *int
		if filter > 0 {
			s := filter
			statusID = &s
		}
		books, err := queries.GetUserBooks(ctx, client, user.ID, statusID, pageSize, page*pageSize)
		return booksOnlyLoadedMsg{books: books, err: err}
	}
}

func (m *Model) loadActivities() tea.Cmd {
	client := m.deps.Client
	user := m.deps.User
	af := m.activityFilter
	return func() tea.Msg {
		if user == nil || user.ID <= 0 {
			return activitiesLoadedMsg{err: fmt.Errorf("user profile not loaded")}
		}
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		var activities []api.Activity
		var err error
		if af == activityFilterFollowing {
			activities, err = queries.GetFollowingActivities(ctx, client, user.ID, 30)
		} else {
			activities, err = queries.GetActivities(ctx, client, user.ID, 30)
		}
		return activitiesLoadedMsg{activities: activities, err: err}
	}
}

// scheduleFilterLoad waits a short delay before triggering the actual load.
// If the user keeps pressing f/F, only the last filter position loads.
func (m *Model) scheduleFilterLoad() tea.Cmd {
	f := m.filter
	return tea.Tick(300*time.Millisecond, func(time.Time) tea.Msg {
		return filterSettledMsg{filter: f}
	})
}
