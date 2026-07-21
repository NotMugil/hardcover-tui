package queries

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/NotMugil/hardcover-tui/internal/api"
)

// GetActivities fetches the current user's own activities using gen.Client.
func GetActivities(ctx context.Context, c *api.Client, userID int, limit int) ([]api.Activity, error) {
	res, err := c.Gen.GetActivities(ctx, userID, limit)
	if err != nil {
		return nil, fmt.Errorf("query activities: %w", err)
	}

	out := make([]api.Activity, len(res.Activities))
	for i, a := range res.Activities {
		act := api.Activity{
			ID:               a.ID,
			Event:            a.Event,
			Data:             a.Data,
			BookID:           a.BookID,
			LikesCount:       a.LikesCount,
			PrivacySettingID: a.PrivacySettingID,
			CreatedAt:        parseRawString(a.CreatedAt),
			User: &api.ActivityUser{
				ID:       a.User.ID,
				Username: parseRawString(a.User.Username),
				Name:     a.User.Name,
			},
		}
		if a.Book != nil {
			b := &api.Book{
				ID:    a.Book.ID,
				Title: parseStringPtr(a.Book.Title),
			}
			if a.Book.Image != nil && a.Book.Image.URL != nil {
				b.Image = &api.Image{URL: *a.Book.Image.URL}
			}
			act.Book = b
		}
		out[i] = act
	}
	return out, nil
}

// GetFollowingActivities fetches the activity feed of users followed by the given user.
func GetFollowingActivities(ctx context.Context, c *api.Client, userID int, limit int) ([]api.Activity, error) {
	res, err := c.Gen.GetFollowingActivities(ctx, userID, limit)
	if err != nil {
		return nil, fmt.Errorf("query following activities: %w", err)
	}

	out := make([]api.Activity, len(res.Activities))
	for i, a := range res.Activities {
		act := api.Activity{
			ID:               a.ID,
			Event:            a.Event,
			Data:             a.Data,
			BookID:           a.BookID,
			LikesCount:       a.LikesCount,
			PrivacySettingID: a.PrivacySettingID,
			CreatedAt:        parseRawString(a.CreatedAt),
			User: &api.ActivityUser{
				ID:       a.User.ID,
				Username: parseRawString(a.User.Username),
				Name:     a.User.Name,
			},
		}
		if a.Book != nil {
			b := &api.Book{
				ID:    a.Book.ID,
				Title: parseStringPtr(a.Book.Title),
			}
			if a.Book.Image != nil && a.Book.Image.URL != nil {
				b.Image = &api.Image{URL: *a.Book.Image.URL}
			}
			act.Book = b
		}
		out[i] = act
	}
	return out, nil
}

// GetLists fetches the user's lists using gen.Client.
func GetLists(ctx context.Context, c *api.Client, userID int) ([]api.List, error) {
	res, err := c.Gen.GetLists(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("query lists: %w", err)
	}

	lists := make([]api.List, len(res.Lists))
	for i, l := range res.Lists {
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
			CreatedAt:        parseRawStringPtr(&l.CreatedAt),
			UpdatedAt:        parseRawStringPtr(&l.UpdatedAt),
		}
	}
	return lists, nil
}

// GetListBooks fetches books within a list using gen.Client.
func GetListBooks(ctx context.Context, c *api.Client, listID int) ([]api.ListBook, error) {
	res, err := c.Gen.GetListBooks(ctx, listID)
	if err != nil {
		return nil, fmt.Errorf("query list_books: %w", err)
	}

	books := make([]api.ListBook, len(res.ListBooks))
	for i, lb := range res.ListBooks {
		b := api.Book{
			ID:           lb.Book.ID,
			Title:        parseStringPtr(lb.Book.Title),
			Subtitle:     lb.Book.Subtitle,
			Description:  lb.Book.Description,
			Pages:        lb.Book.Pages,
			Rating:       parseRawFloat(&lb.Book.Rating),
			RatingsCount: lb.Book.ReviewsCount,
			ReviewsCount: lb.Book.ReviewsCount,
			UsersCount:   lb.Book.UsersCount,
			ReleaseYear:  lb.Book.ReleaseYear,
			Slug:         lb.Book.Slug,
			AudioSeconds: lb.Book.AudioSeconds,
		}
		if lb.Book.Image != nil && lb.Book.Image.URL != nil {
			b.Image = &api.Image{URL: *lb.Book.Image.URL}
		}
		for _, ct := range lb.Book.Contributions {
			if ct.Author != nil {
				b.Contributions = append(b.Contributions, api.Contribution{
					Author: api.Author{
						ID:   ct.Author.ID,
						Name: ct.Author.Name,
						Slug: parseStringPtr(ct.Author.Slug),
					},
				})
			}
		}

		books[i] = api.ListBook{
			ID:        lb.ID,
			ListID:    lb.ListID,
			BookID:    lb.BookID,
			Position:  lb.Position,
			DateAdded: parseRawStringPtr(&lb.DateAdded),
			Book:      b,
		}
	}
	return books, nil
}

// GetGoals fetches the user's reading goals using gen.Client.
func GetGoals(ctx context.Context, c *api.Client, userID int) ([]api.Goal, error) {
	res, err := c.Gen.GetGoals(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("query goals: %w", err)
	}

	goals := make([]api.Goal, len(res.Goals))
	for i, g := range res.Goals {
		var prog float64
		if g.Progress != nil {
			_ = json.Unmarshal(g.Progress, &prog)
		}
		goals[i] = api.Goal{
			ID:               g.ID,
			Goal:             g.Goal,
			Metric:           g.Metric,
			Progress:         prog,
			StartDate:        parseRawString(g.StartDate),
			EndDate:          parseRawString(g.EndDate),
			State:            g.State,
			Description:      g.Description,
			Archived:         g.Archived,
			CompletedAt:      parseRawStringPtr(&g.CompletedAt),
			PrivacySettingID: g.PrivacySettingID,
			UserID:           g.UserID,
		}
	}
	return goals, nil
}

// GetReadingJournals fetches the user's reading journal entries using gen.Client.
func GetReadingJournals(ctx context.Context, c *api.Client, userID int, limit int) ([]api.ReadingJournal, error) {
	res, err := c.Gen.GetReadingJournals(ctx, userID, limit)
	if err != nil {
		return nil, fmt.Errorf("query reading_journals: %w", err)
	}

	journals := make([]api.ReadingJournal, len(res.ReadingJournals))
	for i, j := range res.ReadingJournals {
		rj := api.ReadingJournal{
			ID:               parseRawInt(j.ID),
			Event:            parseStringPtr(j.Event),
			Entry:            j.Entry,
			ActionAt:         parseRawString(j.ActionAt),
			BookID:           j.BookID,
			EditionID:        j.EditionID,
			PrivacySettingID: j.PrivacySettingID,
			LikesCount:       j.LikesCount,
			CreatedAt:        parseRawString(j.CreatedAt),
			UpdatedAt:        parseRawString(j.UpdatedAt),
		}
		if j.Book != nil {
			b := &api.Book{
				ID:    j.Book.ID,
				Title: parseStringPtr(j.Book.Title),
			}
			if j.Book.Image != nil && j.Book.Image.URL != nil {
				b.Image = &api.Image{URL: *j.Book.Image.URL}
			}
			rj.Book = b
		}
		journals[i] = rj
	}
	return journals, nil
}

func parseRawInt(raw json.RawMessage) int {
	if len(raw) == 0 || string(raw) == "null" {
		return 0
	}
	var i int
	if err := json.Unmarshal(raw, &i); err == nil {
		return i
	}
	var s string
	if err := json.Unmarshal(raw, &s); err == nil {
		var parsed int
		if _, err := fmt.Sscanf(s, "%d", &parsed); err == nil {
			return parsed
		}
	}
	return 0
}
