package queries

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/NotMugil/hardcover-tui/internal/api"
)

// Search performs a book search using the Hardcover search API with safe GraphQL variable binding
// and resilient Typesense response unmarshaling via gen.Client.
func Search(ctx context.Context, c *api.Client, query string) ([]api.Book, error) {
	cleanQuery := strings.TrimSpace(query)
	if cleanQuery == "" {
		return nil, nil
	}

	cacheKey := "search:" + strings.ToLower(cleanQuery)
	if val, ok := c.Cache.Get(cacheKey); ok {
		if cached, valid := val.([]api.Book); valid {
			return cached, nil
		}
	}

	searchCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	res, err := c.Gen.SearchBooks(searchCtx, cleanQuery, 20)
	if err != nil {
		return nil, fmt.Errorf("search API request failed: %w", err)
	}

	if res.Search == nil || len(res.Search.Results) == 0 {
		return nil, nil
	}

	books, err := ParseSearchResultsContainer(res.Search.Results)
	if err == nil && len(books) > 0 {
		c.Cache.Set(cacheKey, books, 3*time.Minute)
	}
	return books, err
}

// ParseSearchResults extracts and normalizes Book records from raw Search GraphQL JSON responses.
func ParseSearchResults(raw []byte) ([]api.Book, error) {
	var resp struct {
		Search struct {
			Results json.RawMessage `json:"results"`
		} `json:"search"`
	}
	if err := json.Unmarshal(raw, &resp); err != nil {
		return nil, fmt.Errorf("parse search response container: %w", err)
	}

	return ParseSearchResultsContainer(resp.Search.Results)
}

// ParseSearchResultsContainer parses raw Typesense search result bytes.
func ParseSearchResultsContainer(results json.RawMessage) ([]api.Book, error) {
	if len(results) == 0 || string(results) == "null" {
		return nil, nil
	}

	// If results is double-encoded JSON string, unquote it
	if len(results) > 0 && results[0] == '"' {
		var inner string
		if err := json.Unmarshal(results, &inner); err == nil {
			results = json.RawMessage(inner)
		}
	}

	type searchHitDoc struct {
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
	}

	type searchHit struct {
		Document searchHitDoc `json:"document"`
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

// GetBookByID fetches a book by its primary key using gen.Client with caching.
func GetBookByID(ctx context.Context, c *api.Client, bookID int) (*api.Book, error) {
	cacheKey := fmt.Sprintf("book:%d", bookID)
	if val, ok := c.Cache.Get(cacheKey); ok {
		if cached, valid := val.(*api.Book); valid {
			return cached, nil
		}
	}

	res, err := c.Gen.GetBookByID(ctx, bookID)
	if err != nil {
		return nil, fmt.Errorf("query books_by_pk: %w", err)
	}
	if res.BooksByPk == nil {
		return nil, fmt.Errorf("book %d not found", bookID)
	}

	bk := res.BooksByPk
	b := &api.Book{
		ID:             bk.ID,
		Title:          parseStringPtr(bk.Title),
		Subtitle:       bk.Subtitle,
		Description:    bk.Description,
		Pages:          bk.Pages,
		Rating:         parseRawFloat(&bk.Rating),
		RatingsCount:   bk.RatingsCount,
		ReviewsCount:   bk.ReviewsCount,
		UsersCount:     bk.UsersCount,
		ReleaseYear:    bk.ReleaseYear,
		Slug:           bk.Slug,
		AudioSeconds:   bk.AudioSeconds,
		LiteraryTypeID: bk.LiteraryTypeID,
	}
	if bk.Image != nil && bk.Image.URL != nil {
		b.Image = &api.Image{URL: *bk.Image.URL}
	}
	for _, ct := range bk.Contributions {
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

	c.Cache.Set(cacheKey, b, 10*time.Minute)
	return b, nil
}

// BookTagsResult holds cached genre, mood, and content warning tags.
type BookTagsResult struct {
	Genres          []api.TagItem
	Moods           []api.TagItem
	ContentWarnings []api.TagItem
}

// GetBookTags fetches genres, moods, and content warnings for a book using gen.Client with caching.
func GetBookTags(ctx context.Context, c *api.Client, bookID int) (genres, moods, contentWarnings []api.TagItem, err error) {
	cacheKey := fmt.Sprintf("book_tags:%d", bookID)
	if val, ok := c.Cache.Get(cacheKey); ok {
		if cached, valid := val.(BookTagsResult); valid {
			return cached.Genres, cached.Moods, cached.ContentWarnings, nil
		}
	}

	res, err := c.Gen.GetBookTags(ctx, bookID)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("get book tags: %w", err)
	}

	if res.BooksByPk == nil {
		return nil, nil, nil, nil
	}

	for _, t := range res.BooksByPk.Taggings {
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

	c.Cache.Set(cacheKey, BookTagsResult{
		Genres:          genres,
		Moods:           moods,
		ContentWarnings: contentWarnings,
	}, 15*time.Minute)

	return genres, moods, contentWarnings, nil
}

// GetBookReviews fetches popular community reviews for a book using gen.Client with caching.
func GetBookReviews(ctx context.Context, c *api.Client, bookID, limit int) ([]api.BookReview, error) {
	cacheKey := fmt.Sprintf("book_reviews:%d:%d", bookID, limit)
	if val, ok := c.Cache.Get(cacheKey); ok {
		if cached, valid := val.([]api.BookReview); valid {
			return cached, nil
		}
	}

	res, err := c.Gen.GetBookReviews(ctx, bookID, limit)
	if err != nil {
		return nil, fmt.Errorf("get book reviews: %w", err)
	}

	reviews := make([]api.BookReview, len(res.UserBooks))
	for i, ub := range res.UserBooks {
		reviews[i] = api.BookReview{
			ID:                ub.ID,
			Rating:            parseRawFloat(&ub.Rating),
			Review:            ub.Review,
			ReviewHasSpoilers: ub.ReviewHasSpoilers,
			LikesCount:        ub.LikesCount,
			CreatedAt:         parseRawString(ub.CreatedAt),
			User: api.ReviewUser{
				ID:       ub.User.ID,
				Username: parseRawString(ub.User.Username),
				Name:     ub.User.Name,
			},
		}
	}

	c.Cache.Set(cacheKey, reviews, 5*time.Minute)
	return reviews, nil
}


