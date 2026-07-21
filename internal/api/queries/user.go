package queries

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/NotMugil/hardcover-tui/internal/api"
	"github.com/NotMugil/hardcover-tui/internal/api/gen"
)

// GetMe fetches the authenticated user's profile.
func GetMe(ctx context.Context, c *api.Client) (*api.User, error) {
	res, err := c.Gen.GetMe(ctx)
	if err != nil {
		return nil, fmt.Errorf("query me: %w", err)
	}
	if len(res.Me) == 0 {
		return nil, fmt.Errorf("not authenticated or no user found")
	}

	me := res.Me[0]
	u := &api.User{
		ID:                 me.ID,
		Username:           parseRawString(me.Username),
		Name:               me.Name,
		Bio:                me.Bio,
		Location:           me.Location,
		Link:               me.Link,
		Flair:              me.Flair,
		BooksCount:         me.BooksCount,
		FollowersCount:     me.FollowersCount,
		FollowedUsersCount: me.FollowedUsersCount,
		Pro:                me.Pro,
		PronounPersonal:    me.PronounPersonal,
		PronounPossessive:  me.PronounPossessive,
		CreatedAt:          parseRawTime(me.CreatedAt),
	}
	if me.Image != nil && me.Image.URL != nil && *me.Image.URL != "" {
		u.Image = &api.Image{URL: *me.Image.URL}
	} else if len(me.CachedImage) > 0 {
		var ci struct {
			URL string `json:"url"`
		}
		if err := json.Unmarshal(me.CachedImage, &ci); err == nil && ci.URL != "" {
			u.Image = &api.Image{URL: ci.URL}
		}
	}
	return u, nil
}

// GetUserBooks fetches the user's books with optional status filter.
func GetUserBooks(ctx context.Context, c *api.Client, userID int, statusID *int, limit, offset int) ([]api.UserBook, error) {
	if userID <= 0 {
		return nil, fmt.Errorf("invalid user ID: %d", userID)
	}
	where := &gen.UserBooksBoolExp{
		UserID: &gen.IntComparisonExp{Eq: &userID},
	}
	if statusID != nil {
		where.StatusID = &gen.IntComparisonExp{Eq: statusID}
	}

	res, err := c.Gen.GetUserBooks(ctx, where, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("query user_books: %w", err)
	}

	books := make([]api.UserBook, len(res.UserBooks))
	for i, ub := range res.UserBooks {
		books[i] = mapUserBookFromGetUserBooks(ub)
	}
	return books, nil
}

// GetCurrentlyReading fetches the user's currently reading books.
func GetCurrentlyReading(ctx context.Context, c *api.Client, userID int) ([]api.UserBook, error) {
	status := int(api.StatusCurrentlyReading)
	return GetUserBooks(ctx, c, userID, &status, 20, 0)
}

// GetUserBookByPK fetches a single user_book by primary key.
func GetUserBookByPK(ctx context.Context, c *api.Client, id int) (*api.UserBook, error) {
	res, err := c.Gen.GetUserBookByPk(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("query user_books_by_pk: %w", err)
	}
	if res.UserBooksByPk == nil {
		return nil, fmt.Errorf("user book %d not found", id)
	}

	ub := res.UserBooksByPk
	b := mapBookFromGetUserBookByPk(&ub.Book)

	reads := make([]api.UserBookRead, len(ub.UserBookReads))
	for i, r := range ub.UserBookReads {
		reads[i] = api.UserBookRead{
			ID:              r.ID,
			StartedAt:       parseRawStringPtr(&r.StartedAt),
			FinishedAt:      parseRawStringPtr(&r.FinishedAt),
			ProgressPages:   r.ProgressPages,
			ProgressSeconds: r.ProgressSeconds,
			EditionID:       r.EditionID,
		}
	}

	return &api.UserBook{
		ID:                ub.ID,
		BookID:            ub.BookID,
		StatusID:          ub.StatusID,
		Rating:            parseRawFloat(&ub.Rating),
		Review:            ub.Review,
		ReviewHasSpoilers: ub.ReviewHasSpoilers,
		HasReview:         ub.HasReview,
		DateAdded:         parseRawString(ub.DateAdded),
		ReadCount:         ub.ReadCount,
		Owned:             ub.Owned,
		Starred:           ub.Starred,
		LikesCount:        ub.LikesCount,
		CreatedAt:         parseRawString(ub.CreatedAt),
		PrivateNotes:      ub.PrivateNotes,
		PrivacySettingID:  ub.PrivacySettingID,
		Book:              b,
		UserBookReads:     reads,
	}, nil
}

// GetUserBookByBookID fetches a user's relationship with a specific book.
func GetUserBookByBookID(ctx context.Context, c *api.Client, userID, bookID int) (*api.UserBook, error) {
	res, err := c.Gen.GetUserBookByBookID(ctx, userID, bookID)
	if err != nil {
		return nil, fmt.Errorf("query user_books by book_id: %w", err)
	}
	if len(res.UserBooks) == 0 {
		return nil, nil
	}

	ub := res.UserBooks[0]
	b := mapBookFromGetUserBookByBookID(&ub.Book)

	reads := make([]api.UserBookRead, len(ub.UserBookReads))
	for i, r := range ub.UserBookReads {
		reads[i] = api.UserBookRead{
			ID:              r.ID,
			StartedAt:       parseRawStringPtr(&r.StartedAt),
			FinishedAt:      parseRawStringPtr(&r.FinishedAt),
			ProgressPages:   r.ProgressPages,
			ProgressSeconds: r.ProgressSeconds,
			EditionID:       r.EditionID,
		}
	}

	return &api.UserBook{
		ID:            ub.ID,
		BookID:        ub.BookID,
		StatusID:      ub.StatusID,
		Rating:        parseRawFloat(&ub.Rating),
		Review:        ub.Review,
		HasReview:     ub.HasReview,
		DateAdded:     parseRawString(ub.DateAdded),
		ReadCount:     ub.ReadCount,
		Owned:         ub.Owned,
		Starred:       ub.Starred,
		LikesCount:    ub.LikesCount,
		CreatedAt:     parseRawString(ub.CreatedAt),
		Book:          b,
		UserBookReads: reads,
	}, nil
}

// GetUserBookStats fetches aggregate statistics for a user's books.
func GetUserBookStats(ctx context.Context, c *api.Client, userID int) (*api.UserBookAggregate, error) {
	res, err := c.Gen.GetUserBookAggregate(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("query user_books_aggregate: %w", err)
	}

	agg := &api.UserBookAggregate{}
	if res.UserBooksAggregate.Aggregate != nil {
		agg.Aggregate.Count = res.UserBooksAggregate.Aggregate.Count
		if res.UserBooksAggregate.Aggregate.Avg != nil {
			agg.Aggregate.Avg.Rating = res.UserBooksAggregate.Aggregate.Avg.Rating
		}
	}
	return agg, nil
}

// GetUserBookStatusCounts returns the number of books per status for a user in a single query.
func GetUserBookStatusCounts(ctx context.Context, c *api.Client, userID int) (map[api.StatusID]int, error) {
	res, err := c.Gen.GetUserBookStatusIDs(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("query user_book status_ids: %w", err)
	}

	counts := make(map[api.StatusID]int)
	for _, s := range api.AllStatuses() {
		counts[s] = 0
	}

	for _, ub := range res.UserBooks {
		counts[api.StatusID(ub.StatusID)]++
	}

	return counts, nil
}

// --- Internal Mapping Helpers ---

func mapUserBookFromGetUserBooks(ub *gen.GetUserBooks_UserBooks) api.UserBook {
	b := mapBookFromGetUserBooks(&ub.Book)
	reads := make([]api.UserBookRead, len(ub.UserBookReads))
	for i, r := range ub.UserBookReads {
		reads[i] = api.UserBookRead{
			ID:              r.ID,
			StartedAt:       parseRawStringPtr(&r.StartedAt),
			FinishedAt:      parseRawStringPtr(&r.FinishedAt),
			ProgressPages:   r.ProgressPages,
			ProgressSeconds: r.ProgressSeconds,
			EditionID:       r.EditionID,
		}
	}

	return api.UserBook{
		ID:            ub.ID,
		BookID:        ub.BookID,
		StatusID:      ub.StatusID,
		Rating:        parseRawFloat(&ub.Rating),
		Review:        ub.Review,
		HasReview:     ub.HasReview,
		DateAdded:     parseRawString(ub.DateAdded),
		ReadCount:     ub.ReadCount,
		Owned:         ub.Owned,
		Starred:       ub.Starred,
		LikesCount:    ub.LikesCount,
		CreatedAt:     parseRawString(ub.CreatedAt),
		Book:          b,
		UserBookReads: reads,
	}
}

func mapBookFromGetUserBooks(gb *gen.GetUserBooks_UserBooks_Book) api.Book {
	b := api.Book{
		ID:           gb.ID,
		Title:        parseStringPtr(gb.Title),
		Subtitle:     gb.Subtitle,
		Description:  gb.Description,
		Pages:        gb.Pages,
		Rating:       parseRawFloat(&gb.Rating),
		RatingsCount: gb.RatingsCount,
		ReviewsCount: gb.ReviewsCount,
		UsersCount:   gb.UsersCount,
		ReleaseYear:  gb.ReleaseYear,
		Slug:         gb.Slug,
		AudioSeconds: gb.AudioSeconds,
	}
	if gb.Image != nil && gb.Image.URL != nil {
		b.Image = &api.Image{URL: *gb.Image.URL}
	}
	for _, ct := range gb.Contributions {
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
	return b
}

func mapBookFromGetUserBookByPk(gb *gen.GetUserBookByPK_UserBooksByPk_Book) api.Book {
	b := api.Book{
		ID:           gb.ID,
		Title:        parseStringPtr(gb.Title),
		Subtitle:     gb.Subtitle,
		Description:  gb.Description,
		Pages:        gb.Pages,
		Rating:       parseRawFloat(&gb.Rating),
		RatingsCount: gb.RatingsCount,
		ReviewsCount: gb.ReviewsCount,
		UsersCount:   gb.UsersCount,
		ReleaseYear:  gb.ReleaseYear,
		Slug:         gb.Slug,
		AudioSeconds: gb.AudioSeconds,
	}
	if gb.Image != nil && gb.Image.URL != nil {
		b.Image = &api.Image{URL: *gb.Image.URL}
	}
	for _, ct := range gb.Contributions {
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
	return b
}

func mapBookFromGetUserBookByBookID(gb *gen.GetUserBookByBookID_UserBooks_Book) api.Book {
	b := api.Book{
		ID:           gb.ID,
		Title:        parseStringPtr(gb.Title),
		Subtitle:     gb.Subtitle,
		Description:  gb.Description,
		Pages:        gb.Pages,
		Rating:       parseRawFloat(&gb.Rating),
		RatingsCount: gb.RatingsCount,
		ReviewsCount: gb.ReviewsCount,
		UsersCount:   gb.UsersCount,
		ReleaseYear:  gb.ReleaseYear,
		Slug:         gb.Slug,
		AudioSeconds: gb.AudioSeconds,
	}
	if gb.Image != nil && gb.Image.URL != nil {
		b.Image = &api.Image{URL: *gb.Image.URL}
	}
	for _, ct := range gb.Contributions {
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
	return b
}

func parseRawFloat(raw *json.RawMessage) *float64 {
	if raw == nil || len(*raw) == 0 || string(*raw) == "null" {
		return nil
	}
	var f float64
	if err := json.Unmarshal(*raw, &f); err == nil {
		return &f
	}
	var s string
	if err := json.Unmarshal(*raw, &s); err == nil {
		var parsed float64
		if _, err := fmt.Sscanf(s, "%f", &parsed); err == nil {
			return &parsed
		}
	}
	return nil
}

func parseRawString(raw json.RawMessage) string {
	if len(raw) == 0 || string(raw) == "null" {
		return ""
	}
	var s string
	if err := json.Unmarshal(raw, &s); err == nil {
		return s
	}
	return string(raw)
}

func parseRawStringPtr(raw *json.RawMessage) *string {
	if raw == nil || len(*raw) == 0 || string(*raw) == "null" {
		return nil
	}
	s := parseRawString(*raw)
	return &s
}

func parseRawTime(raw json.RawMessage) time.Time {
	s := parseRawString(raw)
	if s == "" {
		return time.Time{}
	}
	if len(s) >= 2 && s[0] == '"' && s[len(s)-1] == '"' {
		s = s[1 : len(s)-1]
	}
	for _, format := range []string{
		time.RFC3339Nano,
		time.RFC3339,
		"2006-01-02T15:04:05",
		"2006-01-02",
	} {
		if parsed, err := time.Parse(format, s); err == nil {
			return parsed
		}
	}
	return time.Time{}
}

func parseStringPtr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
