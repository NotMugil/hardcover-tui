package queries

import (
	"context"
	"encoding/json"
	"fmt"

	graphql "github.com/hasura/go-graphql-client"

	"hardcover-tui/internal/api"
)

// GetActivities fetches recent user activities.
func GetActivities(ctx context.Context, c *api.Client, userID int, limit int) ([]api.Activity, error) {
	var q struct {
		Activities []struct {
			ID               int    `graphql:"id"`
			Event            string `graphql:"event"`
			BookID           *int   `graphql:"book_id"`
			LikesCount       int    `graphql:"likes_count"`
			PrivacySettingID int    `graphql:"privacy_setting_id"`
			CreatedAt        string `graphql:"created_at"`
			Book             *struct {
				ID    int        `graphql:"id"`
				Title string     `graphql:"title"`
				Image *api.Image `graphql:"image"`
			} `graphql:"book"`
			User struct {
				ID       int     `graphql:"id"`
				Username string  `graphql:"username"`
				Name     *string `graphql:"name"`
			} `graphql:"user"`
		} `graphql:"activities(where: {user_id: {_eq: $userID}}, order_by: {created_at: desc}, limit: $limit)"`
	}

	vars := map[string]interface{}{
		"userID": graphql.Int(userID),
		"limit":  graphql.Int(limit),
	}

	if err := c.Query(ctx, &q, vars); err != nil {
		return nil, fmt.Errorf("query activities: %w", err)
	}

	activities := make([]api.Activity, len(q.Activities))
	for i, a := range q.Activities {
		act := api.Activity{
			ID:               a.ID,
			Event:            a.Event,
			BookID:           a.BookID,
			LikesCount:       a.LikesCount,
			PrivacySettingID: a.PrivacySettingID,
			CreatedAt:        a.CreatedAt,
			User: &api.ActivityUser{
				ID:       a.User.ID,
				Username: a.User.Username,
				Name:     a.User.Name,
			},
		}
		if a.Book != nil {
			b := &api.Book{
				ID:    a.Book.ID,
				Title: a.Book.Title,
				Image: a.Book.Image,
			}
			act.Book = b
		}
		activities[i] = act
	}
	return activities, nil
}

// GetFollowingActivities fetches activities from users the given user follows.
// Uses a two-step approach: first get followed user IDs, then query their activities.
func GetFollowingActivities(ctx context.Context, c *api.Client, userID int, limit int) ([]api.Activity, error) {
	const followsQuery = `query {
		me {
			followed_users {
				followed_user_id
			}
		}
	}`

	raw, err := c.ExecRaw(ctx, followsQuery, nil)
	if err != nil {
		return nil, fmt.Errorf("query followed users: %w", err)
	}

	var followResp struct {
		Me []struct {
			FollowedUsers []struct {
				FollowedUserID int `json:"followed_user_id"`
			} `json:"followed_users"`
		} `json:"me"`
	}
	if err := json.Unmarshal(raw, &followResp); err != nil {
		return nil, fmt.Errorf("parse followed users: %w", err)
	}

	if len(followResp.Me) == 0 || len(followResp.Me[0].FollowedUsers) == 0 {
		return nil, nil
	}

	var followedIDs []int
	for _, fu := range followResp.Me[0].FollowedUsers {
		followedIDs = append(followedIDs, fu.FollowedUserID)
	}

	followedJSON, _ := json.Marshal(followedIDs)
	activitiesQuery := fmt.Sprintf(`query {
		activities(
			where: {user_id: {_in: %s}},
			order_by: {created_at: desc},
			limit: %d
		) {
			id
			event
			book_id
			likes_count
			privacy_setting_id
			created_at
			book {
				id
				title
				image { url }
			}
			user {
				id
				username
				name
			}
		}
	}`, string(followedJSON), limit)

	raw, err = c.ExecRaw(ctx, activitiesQuery, nil)
	if err != nil {
		return nil, fmt.Errorf("query following activities: %w", err)
	}

	var actResp struct {
		Activities []struct {
			ID               int    `json:"id"`
			Event            string `json:"event"`
			BookID           *int   `json:"book_id"`
			LikesCount       int    `json:"likes_count"`
			PrivacySettingID int    `json:"privacy_setting_id"`
			CreatedAt        string `json:"created_at"`
			Book             *struct {
				ID    int    `json:"id"`
				Title string `json:"title"`
				Image *struct {
					URL string `json:"url"`
				} `json:"image"`
			} `json:"book"`
			User struct {
				ID       int     `json:"id"`
				Username string  `json:"username"`
				Name     *string `json:"name"`
			} `json:"user"`
		} `json:"activities"`
	}
	if err := json.Unmarshal(raw, &actResp); err != nil {
		return nil, fmt.Errorf("parse following activities: %w", err)
	}

	var all []api.Activity
	for _, a := range actResp.Activities {
		act := api.Activity{
			ID:               a.ID,
			Event:            a.Event,
			BookID:           a.BookID,
			LikesCount:       a.LikesCount,
			PrivacySettingID: a.PrivacySettingID,
			CreatedAt:        a.CreatedAt,
			User: &api.ActivityUser{
				ID:       a.User.ID,
				Username: a.User.Username,
				Name:     a.User.Name,
			},
		}
		if a.Book != nil {
			b := &api.Book{
				ID:    a.Book.ID,
				Title: a.Book.Title,
			}
			if a.Book.Image != nil {
				b.Image = &api.Image{URL: a.Book.Image.URL}
			}
			act.Book = b
		}
		all = append(all, act)
	}

	return all, nil
}

// GetLists fetches the user's lists.
func GetLists(ctx context.Context, c *api.Client, userID int) ([]api.List, error) {
	var q struct {
		Lists []struct {
			ID               int     `graphql:"id"`
			Name             string  `graphql:"name"`
			Description      *string `graphql:"description"`
			BooksCount       int     `graphql:"books_count"`
			LikesCount       int     `graphql:"likes_count"`
			Public           bool    `graphql:"public"`
			Ranked           bool    `graphql:"ranked"`
			PrivacySettingID int     `graphql:"privacy_setting_id"`
			Slug             *string `graphql:"slug"`
			UserID           int     `graphql:"user_id"`
			CreatedAt        *string `graphql:"created_at"`
			UpdatedAt        *string `graphql:"updated_at"`
		} `graphql:"lists(where: {user_id: {_eq: $userID}}, order_by: {updated_at: desc})"`
	}

	vars := map[string]interface{}{
		"userID": graphql.Int(userID),
	}

	if err := c.Query(ctx, &q, vars); err != nil {
		return nil, fmt.Errorf("query lists: %w", err)
	}

	lists := make([]api.List, len(q.Lists))
	for i, l := range q.Lists {
		lists[i] = api.List{
			ID:               l.ID,
			Name:             l.Name,
			Description:      l.Description,
			BooksCount:       l.BooksCount,
			LikesCount:       l.LikesCount,
			Public:           l.Public,
			Ranked:           l.Ranked,
			PrivacySettingID: l.PrivacySettingID,
			Slug:             l.Slug,
			UserID:           l.UserID,
			CreatedAt:        l.CreatedAt,
			UpdatedAt:        l.UpdatedAt,
		}
	}
	return lists, nil
}

// GetListBooks fetches books within a list.
func GetListBooks(ctx context.Context, c *api.Client, listID int) ([]api.ListBook, error) {
	var q struct {
		ListBooks []struct {
			ID        int          `graphql:"id"`
			ListID    int          `graphql:"list_id"`
			BookID    int          `graphql:"book_id"`
			Position  *int         `graphql:"position"`
			DateAdded *string      `graphql:"date_added"`
			Book      bookFragment `graphql:"book"`
		} `graphql:"list_books(where: {list_id: {_eq: $listID}}, order_by: {position: asc})"`
	}

	vars := map[string]interface{}{
		"listID": graphql.Int(listID),
	}

	if err := c.Query(ctx, &q, vars); err != nil {
		return nil, fmt.Errorf("query list_books: %w", err)
	}

	books := make([]api.ListBook, len(q.ListBooks))
	for i, lb := range q.ListBooks {
		books[i] = api.ListBook{
			ID:        lb.ID,
			ListID:    lb.ListID,
			BookID:    lb.BookID,
			Position:  lb.Position,
			DateAdded: lb.DateAdded,
			Book:      lb.Book.toBook(),
		}
	}
	return books, nil
}

// GetGoals fetches the user's reading goals.
func GetGoals(ctx context.Context, c *api.Client, userID int) ([]api.Goal, error) {
	var q struct {
		Goals []struct {
			ID               int     `graphql:"id"`
			Goal             int     `graphql:"goal"`
			Metric           string  `graphql:"metric"`
			Progress         float64 `graphql:"progress"`
			StartDate        string  `graphql:"start_date"`
			EndDate          string  `graphql:"end_date"`
			State            string  `graphql:"state"`
			Description      *string `graphql:"description"`
			Archived         bool    `graphql:"archived"`
			CompletedAt      *string `graphql:"completed_at"`
			PrivacySettingID *int    `graphql:"privacy_setting_id"`
			UserID           int     `graphql:"user_id"`
		} `graphql:"goals(where: {user_id: {_eq: $userID}, archived: {_eq: false}}, order_by: {start_date: desc})"`
	}

	vars := map[string]interface{}{
		"userID": graphql.Int(userID),
	}

	if err := c.Query(ctx, &q, vars); err != nil {
		return nil, fmt.Errorf("query goals: %w", err)
	}

	goals := make([]api.Goal, len(q.Goals))
	for i, g := range q.Goals {
		goals[i] = api.Goal{
			ID:               g.ID,
			Goal:             g.Goal,
			Metric:           g.Metric,
			Progress:         g.Progress,
			StartDate:        g.StartDate,
			EndDate:          g.EndDate,
			State:            g.State,
			Description:      g.Description,
			Archived:         g.Archived,
			CompletedAt:      g.CompletedAt,
			PrivacySettingID: g.PrivacySettingID,
			UserID:           g.UserID,
		}
	}
	return goals, nil
}

// GetReadingJournals fetches the user's reading journal entries.
func GetReadingJournals(ctx context.Context, c *api.Client, userID int, limit int) ([]api.ReadingJournal, error) {
	var q struct {
		Journals []struct {
			ID               int     `graphql:"id"`
			Event            string  `graphql:"event"`
			Entry            *string `graphql:"entry"`
			ActionAt         string  `graphql:"action_at"`
			BookID           *int    `graphql:"book_id"`
			EditionID        *int    `graphql:"edition_id"`
			PrivacySettingID int     `graphql:"privacy_setting_id"`
			LikesCount       int     `graphql:"likes_count"`
			CreatedAt        string  `graphql:"created_at"`
			UpdatedAt        string  `graphql:"updated_at"`
			Book             *struct {
				ID    int        `graphql:"id"`
				Title string     `graphql:"title"`
				Image *api.Image `graphql:"image"`
			} `graphql:"book"`
		} `graphql:"reading_journals(where: {user_id: {_eq: $userID}}, order_by: {action_at: desc}, limit: $limit)"`
	}

	vars := map[string]interface{}{
		"userID": graphql.Int(userID),
		"limit":  graphql.Int(limit),
	}

	if err := c.Query(ctx, &q, vars); err != nil {
		return nil, fmt.Errorf("query reading_journals: %w", err)
	}

	journals := make([]api.ReadingJournal, len(q.Journals))
	for i, j := range q.Journals {
		rj := api.ReadingJournal{
			ID:               j.ID,
			Event:            j.Event,
			Entry:            j.Entry,
			ActionAt:         j.ActionAt,
			BookID:           j.BookID,
			EditionID:        j.EditionID,
			PrivacySettingID: j.PrivacySettingID,
			LikesCount:       j.LikesCount,
			CreatedAt:        j.CreatedAt,
			UpdatedAt:        j.UpdatedAt,
		}
		if j.Book != nil {
			b := &api.Book{
				ID:    j.Book.ID,
				Title: j.Book.Title,
				Image: j.Book.Image,
			}
			rj.Book = b
		}
		journals[i] = rj
	}
	return journals, nil
}
