package queries

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	graphql "github.com/hasura/go-graphql-client"

	"github.com/NotMugil/hardcover-tui/internal/api"
)

// Search performs a book search using the Hardcover search API.
// Follows the approach from github.com/Kameleon21/oku: inline query values,
// input sanitization, and search fields/weights for better relevance.
func Search(ctx context.Context, c *api.Client, query string) ([]api.Book, error) {
	sanitized := strings.ReplaceAll(query, `\`, `\\`)
	sanitized = strings.ReplaceAll(sanitized, `"`, `\"`)

	gqlQuery := fmt.Sprintf(`query {
		search(query: "%s", query_type: "Book", per_page: 20, page: 1, fields: "title,author_names", weights: "7,3") {
			results
		}
	}`, sanitized)

	raw, err := c.ExecRaw(ctx, gqlQuery, nil)
	if err != nil {
		return nil, fmt.Errorf("search: %w", err)
	}

	var resp struct {
		Search struct {
			Results json.RawMessage `json:"results"`
		} `json:"search"`
	}
	if err := json.Unmarshal(raw, &resp); err != nil {
		return nil, fmt.Errorf("parse search response: %w", err)
	}

	if len(resp.Search.Results) == 0 || string(resp.Search.Results) == "null" {
		return nil, nil
	}

	results := resp.Search.Results

	if len(results) > 0 && results[0] == '"' {
		var inner string
		if err := json.Unmarshal(results, &inner); err == nil {
			results = json.RawMessage(inner)
		}
	}

	type searchHit struct {
		Document struct {
			ID           json.Number `json:"id"`
			Title        string      `json:"title"`
			Slug         string      `json:"slug"`
			AuthorNames  []string    `json:"author_names"`
			Pages        json.Number `json:"pages"`
			Rating       float64     `json:"rating"`
			UsersCount   int         `json:"users_count"`
			ReleaseYear  int         `json:"release_year"`
			Genres       []string    `json:"genres"`
			Description  string      `json:"description"`
			HasAudiobook bool        `json:"has_audiobook"`
			HasEbook     bool        `json:"has_ebook"`
		} `json:"document"`
	}
	type hitsContainer struct {
		Hits []searchHit `json:"hits"`
	}

	var allHits []searchHit

	var flat hitsContainer
	if err := json.Unmarshal(results, &flat); err == nil && len(flat.Hits) > 0 {
		allHits = flat.Hits
	}

	if len(allHits) == 0 {
		var groups []hitsContainer
		if err := json.Unmarshal(results, &groups); err == nil {
			for _, g := range groups {
				allHits = append(allHits, g.Hits...)
			}
		}
	}

	if len(allHits) == 0 {
		var wrapper struct {
			GroupedHits []hitsContainer `json:"grouped_hits"`
		}
		if err := json.Unmarshal(results, &wrapper); err == nil {
			for _, g := range wrapper.GroupedHits {
				allHits = append(allHits, g.Hits...)
			}
		}
	}

	if len(allHits) == 0 {
		return nil, nil
	}

	books := make([]api.Book, 0, len(allHits))
	for _, hit := range allHits {
		doc := hit.Document
		docID, _ := doc.ID.Int64()
		docPages, _ := doc.Pages.Int64()

		b := api.Book{
			ID:           int(docID),
			Title:        doc.Title,
			Slug:         &doc.Slug,
			UsersCount:   doc.UsersCount,
			HasAudiobook: doc.HasAudiobook,
			HasEbook:     doc.HasEbook,
		}
		if doc.Rating > 0 {
			r := doc.Rating
			b.Rating = &r
		}
		if docPages > 0 {
			p := int(docPages)
			b.Pages = &p
		}
		if doc.ReleaseYear > 0 {
			ry := doc.ReleaseYear
			b.ReleaseYear = &ry
		}
		if doc.Description != "" {
			d := doc.Description
			b.Description = &d
		}
		for _, name := range doc.AuthorNames {
			b.Contributions = append(b.Contributions, api.Contribution{
				Author: api.Author{Name: name},
			})
		}
		for _, g := range doc.Genres {
			b.Genres = append(b.Genres, api.TagItem{Name: g})
		}
		books = append(books, b)
	}
	return books, nil
}

// GetBookByID fetches a book by its primary key (book ID, not user_book ID).
func GetBookByID(ctx context.Context, c *api.Client, bookID int) (*api.Book, error) {
	var q struct {
		Book *struct {
			ID             int        `graphql:"id"`
			Title          string     `graphql:"title"`
			Subtitle       *string    `graphql:"subtitle"`
			Description    *string    `graphql:"description"`
			Pages          *int       `graphql:"pages"`
			Rating         *float64   `graphql:"rating"`
			RatingsCount   int        `graphql:"ratings_count"`
			ReviewsCount   int        `graphql:"reviews_count"`
			UsersCount     int        `graphql:"users_count"`
			ReleaseYear    *int       `graphql:"release_year"`
			Slug           *string    `graphql:"slug"`
			AudioSeconds   *int       `graphql:"audio_seconds"`
			LiteraryTypeID *int       `graphql:"literary_type_id"`
			Image          *api.Image `graphql:"image"`
			Contributions  []struct {
				Author struct {
					ID   int    `graphql:"id"`
					Name string `graphql:"name"`
					Slug string `graphql:"slug"`
				} `graphql:"author"`
			} `graphql:"contributions"`
		} `graphql:"books_by_pk(id: $id)"`
	}

	vars := map[string]interface{}{
		"id": graphql.Int(bookID),
	}

	if err := c.Query(ctx, &q, vars); err != nil {
		return nil, fmt.Errorf("query books_by_pk: %w", err)
	}
	if q.Book == nil {
		return nil, fmt.Errorf("book %d not found", bookID)
	}

	b := &api.Book{
		ID:             q.Book.ID,
		Title:          q.Book.Title,
		Subtitle:       q.Book.Subtitle,
		Description:    q.Book.Description,
		Pages:          q.Book.Pages,
		Rating:         q.Book.Rating,
		RatingsCount:   q.Book.RatingsCount,
		ReviewsCount:   q.Book.ReviewsCount,
		UsersCount:     q.Book.UsersCount,
		ReleaseYear:    q.Book.ReleaseYear,
		Slug:           q.Book.Slug,
		AudioSeconds:   q.Book.AudioSeconds,
		LiteraryTypeID: q.Book.LiteraryTypeID,
		Image:          q.Book.Image,
	}
	for _, ct := range q.Book.Contributions {
		b.Contributions = append(b.Contributions, api.Contribution{
			Author: api.Author{
				ID:   ct.Author.ID,
				Name: ct.Author.Name,
				Slug: ct.Author.Slug,
			},
		})
	}
	return b, nil
}

// GetBookTags fetches genres, moods, and content warnings for a book via ExecRaw.
func GetBookTags(ctx context.Context, c *api.Client, bookID int) (genres, moods, contentWarnings []api.TagItem, err error) {
	const gqlQuery = `query ($bookId: Int!) {
		books_by_pk(id: $bookId) {
			taggings {
				tag {
					tag
					tag_category_id
				}
			}
		}
	}`

	vars := map[string]any{
		"bookId": bookID,
	}

	raw, e := c.ExecRaw(ctx, gqlQuery, vars)
	if e != nil {
		return nil, nil, nil, fmt.Errorf("get book tags: %w", e)
	}

	var resp struct {
		BooksByPK *struct {
			Taggings []struct {
				Tag struct {
					Tag           string `json:"tag"`
					TagCategoryID int    `json:"tag_category_id"`
				} `json:"tag"`
			} `json:"taggings"`
		} `json:"books_by_pk"`
	}
	if e := json.Unmarshal(raw, &resp); e != nil {
		return nil, nil, nil, fmt.Errorf("parse book tags: %w", e)
	}

	if resp.BooksByPK == nil {
		return nil, nil, nil, nil
	}

	for _, t := range resp.BooksByPK.Taggings {
		item := api.TagItem{Name: t.Tag.Tag}
		switch t.Tag.TagCategoryID {
		case api.TagCategoryGenre:
			genres = append(genres, item)
		case api.TagCategoryMood:
			moods = append(moods, item)
		case api.TagCategoryContentWarning:
			contentWarnings = append(contentWarnings, item)
		}
	}
	return genres, moods, contentWarnings, nil
}

// GetBookReviews fetches popular community reviews for a book.
func GetBookReviews(ctx context.Context, c *api.Client, bookID, limit int) ([]api.BookReview, error) {
	const gqlQuery = `query ($bookId: Int!, $limit: Int!) {
		user_books(
			where: {book_id: {_eq: $bookId}, has_review: {_eq: true}}
			order_by: {likes_count: desc}
			limit: $limit
		) {
			id
			rating
			review
			review_has_spoilers
			likes_count
			created_at
			user {
				id
				username
				name
			}
		}
	}`

	vars := map[string]any{
		"bookId": bookID,
		"limit":  limit,
	}

	raw, err := c.ExecRaw(ctx, gqlQuery, vars)
	if err != nil {
		return nil, fmt.Errorf("get book reviews: %w", err)
	}

	var resp struct {
		UserBooks []struct {
			ID                int      `json:"id"`
			Rating            *float64 `json:"rating"`
			Review            *string  `json:"review"`
			ReviewHasSpoilers bool     `json:"review_has_spoilers"`
			LikesCount        int      `json:"likes_count"`
			CreatedAt         string   `json:"created_at"`
			User              struct {
				ID       int     `json:"id"`
				Username string  `json:"username"`
				Name     *string `json:"name"`
			} `json:"user"`
		} `json:"user_books"`
	}
	if err := json.Unmarshal(raw, &resp); err != nil {
		return nil, fmt.Errorf("parse book reviews: %w", err)
	}

	reviews := make([]api.BookReview, len(resp.UserBooks))
	for i, ub := range resp.UserBooks {
		reviews[i] = api.BookReview{
			ID:                ub.ID,
			Rating:            ub.Rating,
			Review:            ub.Review,
			ReviewHasSpoilers: ub.ReviewHasSpoilers,
			LikesCount:        ub.LikesCount,
			CreatedAt:         ub.CreatedAt,
			User: api.ReviewUser{
				ID:       ub.User.ID,
				Username: ub.User.Username,
				Name:     ub.User.Name,
			},
		}
	}
	return reviews, nil
}
