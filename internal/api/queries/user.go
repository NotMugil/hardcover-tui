package queries

import (
	"context"
	"fmt"
	"time"

	graphql "github.com/hasura/go-graphql-client"

	"github.com/NotMugil/hardcover-tui/internal/api"
)

// GetMe fetches the authenticated user's profile.
func GetMe(ctx context.Context, c *api.Client) (*api.User, error) {
	var q struct {
		Me []struct {
			ID                 int        `graphql:"id"`
			Username           string     `graphql:"username"`
			Name               *string    `graphql:"name"`
			Bio                *string    `graphql:"bio"`
			Location           *string    `graphql:"location"`
			Link               *string    `graphql:"link"`
			Flair              *string    `graphql:"flair"`
			BooksCount         int        `graphql:"books_count"`
			FollowersCount     int        `graphql:"followers_count"`
			FollowedUsersCount int        `graphql:"followed_users_count"`
			Pro                bool       `graphql:"pro"`
			PronounPersonal    string     `graphql:"pronoun_personal"`
			PronounPossessive  string     `graphql:"pronoun_possessive"`
			Image              *api.Image `graphql:"image"`
			CreatedAt          localTime  `graphql:"created_at"`
		} `graphql:"me"`
	}

	if err := c.Query(ctx, &q, nil); err != nil {
		return nil, fmt.Errorf("query me: %w", err)
	}
	if len(q.Me) == 0 {
		return nil, fmt.Errorf("not authenticated or no user found")
	}

	me := q.Me[0]
	return &api.User{
		ID:                 me.ID,
		Username:           me.Username,
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
		Image:              me.Image,
		CreatedAt:          me.CreatedAt.Time,
	}, nil
}

// GetUserBooks fetches the user's books with optional status filter.
func GetUserBooks(ctx context.Context, c *api.Client, userID int, statusID *int, limit, offset int) ([]api.UserBook, error) {
	var q struct {
		UserBooks []struct {
			ID            int          `graphql:"id"`
			BookID        int          `graphql:"book_id"`
			StatusID      int          `graphql:"status_id"`
			Rating        *float64     `graphql:"rating"`
			Review        *string      `graphql:"review"`
			HasReview     bool         `graphql:"has_review"`
			DateAdded     string       `graphql:"date_added"`
			ReadCount     int          `graphql:"read_count"`
			Owned         bool         `graphql:"owned"`
			Starred       bool         `graphql:"starred"`
			LikesCount    int          `graphql:"likes_count"`
			CreatedAt     string       `graphql:"created_at"`
			Book          bookFragment `graphql:"book"`
			UserBookReads []ubReadFrag `graphql:"user_book_reads"`
		} `graphql:"user_books(where: $where, order_by: {updated_at: desc}, limit: $limit, offset: $offset)"`
	}

	where := map[string]interface{}{
		"user_id": map[string]interface{}{"_eq": userID},
	}
	if statusID != nil {
		where["status_id"] = map[string]interface{}{"_eq": *statusID}
	}

	vars := map[string]interface{}{
		"where":  user_books_bool_exp(where),
		"limit":  graphql.Int(limit),
		"offset": graphql.Int(offset),
	}

	if err := c.Query(ctx, &q, vars); err != nil {
		return nil, fmt.Errorf("query user_books: %w", err)
	}

	books := make([]api.UserBook, len(q.UserBooks))
	for i, ub := range q.UserBooks {
		books[i] = api.UserBook{
			ID:            ub.ID,
			BookID:        ub.BookID,
			StatusID:      ub.StatusID,
			Rating:        ub.Rating,
			Review:        ub.Review,
			HasReview:     ub.HasReview,
			DateAdded:     ub.DateAdded,
			ReadCount:     ub.ReadCount,
			Owned:         ub.Owned,
			Starred:       ub.Starred,
			LikesCount:    ub.LikesCount,
			CreatedAt:     ub.CreatedAt,
			Book:          ub.Book.toBook(),
			UserBookReads: toReads(ub.UserBookReads),
		}
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
	var q struct {
		UserBook *struct {
			ID                int          `graphql:"id"`
			BookID            int          `graphql:"book_id"`
			StatusID          int          `graphql:"status_id"`
			Rating            *float64     `graphql:"rating"`
			Review            *string      `graphql:"review"`
			ReviewHasSpoilers bool         `graphql:"review_has_spoilers"`
			HasReview         bool         `graphql:"has_review"`
			DateAdded         string       `graphql:"date_added"`
			ReadCount         int          `graphql:"read_count"`
			Owned             bool         `graphql:"owned"`
			Starred           bool         `graphql:"starred"`
			LikesCount        int          `graphql:"likes_count"`
			CreatedAt         string       `graphql:"created_at"`
			PrivateNotes      *string      `graphql:"private_notes"`
			PrivacySettingID  int          `graphql:"privacy_setting_id"`
			Book              bookFragment `graphql:"book"`
			UserBookReads     []ubReadFrag `graphql:"user_book_reads"`
		} `graphql:"user_books_by_pk(id: $id)"`
	}

	vars := map[string]interface{}{
		"id": graphql.Int(id),
	}

	if err := c.Query(ctx, &q, vars); err != nil {
		return nil, fmt.Errorf("query user_books_by_pk: %w", err)
	}
	if q.UserBook == nil {
		return nil, fmt.Errorf("user book %d not found", id)
	}

	ub := q.UserBook
	return &api.UserBook{
		ID:                ub.ID,
		BookID:            ub.BookID,
		StatusID:          ub.StatusID,
		Rating:            ub.Rating,
		Review:            ub.Review,
		ReviewHasSpoilers: ub.ReviewHasSpoilers,
		HasReview:         ub.HasReview,
		DateAdded:         ub.DateAdded,
		ReadCount:         ub.ReadCount,
		Owned:             ub.Owned,
		Starred:           ub.Starred,
		LikesCount:        ub.LikesCount,
		CreatedAt:         ub.CreatedAt,
		PrivateNotes:      ub.PrivateNotes,
		PrivacySettingID:  ub.PrivacySettingID,
		Book:              ub.Book.toBook(),
		UserBookReads:     toReads(ub.UserBookReads),
	}, nil
}

// GetUserBookByBookID fetches a user's relationship with a specific book.
// Returns nil (no error) if the user doesn't have this book in their library.
func GetUserBookByBookID(ctx context.Context, c *api.Client, userID, bookID int) (*api.UserBook, error) {
	var q struct {
		UserBooks []struct {
			ID            int          `graphql:"id"`
			BookID        int          `graphql:"book_id"`
			StatusID      int          `graphql:"status_id"`
			Rating        *float64     `graphql:"rating"`
			Review        *string      `graphql:"review"`
			HasReview     bool         `graphql:"has_review"`
			DateAdded     string       `graphql:"date_added"`
			ReadCount     int          `graphql:"read_count"`
			Owned         bool         `graphql:"owned"`
			Starred       bool         `graphql:"starred"`
			LikesCount    int          `graphql:"likes_count"`
			CreatedAt     string       `graphql:"created_at"`
			Book          bookFragment `graphql:"book"`
			UserBookReads []ubReadFrag `graphql:"user_book_reads"`
		} `graphql:"user_books(where: {user_id: {_eq: $userID}, book_id: {_eq: $bookID}}, limit: 1)"`
	}

	vars := map[string]interface{}{
		"userID": graphql.Int(userID),
		"bookID": graphql.Int(bookID),
	}

	if err := c.Query(ctx, &q, vars); err != nil {
		return nil, fmt.Errorf("query user_books by book_id: %w", err)
	}
	if len(q.UserBooks) == 0 {
		return nil, nil
	}

	ub := q.UserBooks[0]
	return &api.UserBook{
		ID:            ub.ID,
		BookID:        ub.BookID,
		StatusID:      ub.StatusID,
		Rating:        ub.Rating,
		Review:        ub.Review,
		HasReview:     ub.HasReview,
		DateAdded:     ub.DateAdded,
		ReadCount:     ub.ReadCount,
		Owned:         ub.Owned,
		Starred:       ub.Starred,
		LikesCount:    ub.LikesCount,
		CreatedAt:     ub.CreatedAt,
		Book:          ub.Book.toBook(),
		UserBookReads: toReads(ub.UserBookReads),
	}, nil
}

// GetUserBookStats fetches aggregate statistics for a user's books.
func GetUserBookStats(ctx context.Context, c *api.Client, userID int) (*api.UserBookAggregate, error) {
	var q struct {
		Agg struct {
			Aggregate struct {
				Count int `graphql:"count"`
				Avg   struct {
					Rating *float64 `graphql:"rating"`
				} `graphql:"avg"`
			} `graphql:"aggregate"`
		} `graphql:"user_books_aggregate(where: {user_id: {_eq: $userID}})"`
	}

	vars := map[string]interface{}{
		"userID": graphql.Int(userID),
	}

	if err := c.Query(ctx, &q, vars); err != nil {
		return nil, fmt.Errorf("query user_books_aggregate: %w", err)
	}

	agg := &api.UserBookAggregate{}
	agg.Aggregate.Count = q.Agg.Aggregate.Count
	agg.Aggregate.Avg.Rating = q.Agg.Aggregate.Avg.Rating
	return agg, nil
}

// GetUserBookStatusCounts returns the number of books per status for a user.
func GetUserBookStatusCounts(ctx context.Context, c *api.Client, userID int) (map[api.StatusID]int, error) {
	counts := make(map[api.StatusID]int)
	for _, s := range api.AllStatuses() {
		sid := int(s)
		books, err := GetUserBooks(ctx, c, userID, &sid, 0, 0)
		if err != nil {
			return nil, err
		}
		counts[s] = len(books)
	}
	return counts, nil
}

// --- internal fragment types for queries ---

type bookFragment struct {
	ID            int        `graphql:"id"`
	Title         string     `graphql:"title"`
	Subtitle      *string    `graphql:"subtitle"`
	Description   *string    `graphql:"description"`
	Pages         *int       `graphql:"pages"`
	Rating        *float64   `graphql:"rating"`
	RatingsCount  int        `graphql:"ratings_count"`
	ReviewsCount  int        `graphql:"reviews_count"`
	UsersCount    int        `graphql:"users_count"`
	ReleaseYear   *int       `graphql:"release_year"`
	Slug          *string    `graphql:"slug"`
	AudioSeconds  *int       `graphql:"audio_seconds"`
	Image         *api.Image `graphql:"image"`
	Contributions []struct {
		Author struct {
			ID   int    `graphql:"id"`
			Name string `graphql:"name"`
			Slug string `graphql:"slug"`
		} `graphql:"author"`
	} `graphql:"contributions"`
}

func (bf bookFragment) toBook() api.Book {
	b := api.Book{
		ID:           bf.ID,
		Title:        bf.Title,
		Subtitle:     bf.Subtitle,
		Description:  bf.Description,
		Pages:        bf.Pages,
		Rating:       bf.Rating,
		RatingsCount: bf.RatingsCount,
		ReviewsCount: bf.ReviewsCount,
		UsersCount:   bf.UsersCount,
		ReleaseYear:  bf.ReleaseYear,
		Slug:         bf.Slug,
		AudioSeconds: bf.AudioSeconds,
		Image:        bf.Image,
	}
	for _, ct := range bf.Contributions {
		b.Contributions = append(b.Contributions, api.Contribution{
			Author: api.Author{
				ID:   ct.Author.ID,
				Name: ct.Author.Name,
				Slug: ct.Author.Slug,
			},
		})
	}
	return b
}

type ubReadFrag struct {
	ID              int     `graphql:"id"`
	StartedAt       *string `graphql:"started_at"`
	FinishedAt      *string `graphql:"finished_at"`
	ProgressPages   *int    `graphql:"progress_pages"`
	ProgressSeconds *int    `graphql:"progress_seconds"`
	EditionID       *int    `graphql:"edition_id"`
}

func toReads(frags []ubReadFrag) []api.UserBookRead {
	reads := make([]api.UserBookRead, len(frags))
	for i, f := range frags {
		reads[i] = api.UserBookRead{
			ID:              f.ID,
			StartedAt:       f.StartedAt,
			FinishedAt:      f.FinishedAt,
			ProgressPages:   f.ProgressPages,
			ProgressSeconds: f.ProgressSeconds,
			EditionID:       f.EditionID,
		}
	}
	return reads
}

// localTime helps parse the flexible timestamp format from the API.
type localTime struct {
	time.Time
}

func (t *localTime) UnmarshalJSON(data []byte) error {
	s := string(data)
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
			t.Time = parsed
			return nil
		}
	}
	return fmt.Errorf("unable to parse time: %s", s)
}

// user_books_bool_exp is a marker type for the user_books where clause.
// The go-graphql-client derives the GraphQL variable type from the Go type name,
// so this must match Hasura's expected input type exactly.
type user_books_bool_exp map[string]interface{}
