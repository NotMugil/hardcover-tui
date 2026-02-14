package queries

import (
	"context"
	"encoding/json"
	"fmt"

	"hardcover-tui/internal/api"
)

// GetUserBooksForStats fetches user books with book metadata for stats.
// Uses taggings with distinct_on to deduplicate and tag_category_id to classify genres/moods.
// Also fetches edition format from user_book_reads for format breakdown stats.
func GetUserBooksForStats(ctx context.Context, c *api.Client, userID int) ([]api.StatsUserBook, error) {
	const gqlQuery = `query ($userId: Int!) {
		user_books(where: {user_id: {_eq: $userId}}, limit: 500) {
			status_id
			book {
				literary_type_id
				taggings(distinct_on: [tag_id]) {
					tag {
						tag
						tag_category_id
					}
				}
			}
			user_book_reads(limit: 1, order_by: {id: desc}) {
				edition {
					edition_format
				}
			}
		}
	}`

	vars := map[string]any{
		"userId": userID,
	}

	raw, err := c.ExecRaw(ctx, gqlQuery, vars)
	if err != nil {
		return nil, fmt.Errorf("get user books for stats: %w", err)
	}

	var resp struct {
		UserBooks []struct {
			StatusID int `json:"status_id"`
			Book     struct {
				LiteraryTypeID *int `json:"literary_type_id"`
				Taggings       []struct {
					Tag struct {
						Tag           string `json:"tag"`
						TagCategoryID int    `json:"tag_category_id"`
					} `json:"tag"`
				} `json:"taggings"`
			} `json:"book"`
			UserBookReads []struct {
				Edition *struct {
					EditionFormat *string `json:"edition_format"`
				} `json:"edition"`
			} `json:"user_book_reads"`
		} `json:"user_books"`
	}
	if err := json.Unmarshal(raw, &resp); err != nil {
		return nil, fmt.Errorf("parse user books for stats: %w", err)
	}

	result := make([]api.StatsUserBook, len(resp.UserBooks))
	for i, ub := range resp.UserBooks {
		sub := api.StatsUserBook{
			StatusID:       ub.StatusID,
			LiteraryTypeID: ub.Book.LiteraryTypeID,
		}

		for _, t := range ub.Book.Taggings {
			switch t.Tag.TagCategoryID {
			case api.TagCategoryGenre:
				sub.Genres = append(sub.Genres, t.Tag.Tag)
			}
		}

		if len(ub.UserBookReads) > 0 && ub.UserBookReads[0].Edition != nil && ub.UserBookReads[0].Edition.EditionFormat != nil {
			sub.EditionFormat = ub.UserBookReads[0].Edition.EditionFormat
		}

		result[i] = sub
	}
	return result, nil
}

// GetReadingHistory fetches finished reads with dates, pages, and edition format for time-series charts.
func GetReadingHistory(ctx context.Context, c *api.Client, userID int) ([]api.ReadingHistoryEntry, error) {
	const gqlQuery = `query ($userId: Int!) {
		user_book_reads(
			where: {
				user_book: {user_id: {_eq: $userId}},
				finished_at: {_is_null: false}
			},
			order_by: {finished_at: asc},
			limit: 1000
		) {
			finished_at
			progress_pages
			edition {
				edition_format
				pages
			}
		}
	}`

	vars := map[string]any{
		"userId": userID,
	}

	raw, err := c.ExecRaw(ctx, gqlQuery, vars)
	if err != nil {
		return nil, fmt.Errorf("get reading history: %w", err)
	}

	var resp struct {
		UserBookReads []struct {
			FinishedAt    *string `json:"finished_at"`
			ProgressPages *int    `json:"progress_pages"`
			Edition       *struct {
				EditionFormat *string `json:"edition_format"`
				Pages         *int    `json:"pages"`
			} `json:"edition"`
		} `json:"user_book_reads"`
	}
	if err := json.Unmarshal(raw, &resp); err != nil {
		return nil, fmt.Errorf("parse reading history: %w", err)
	}

	var result []api.ReadingHistoryEntry
	for _, r := range resp.UserBookReads {
		if r.FinishedAt == nil {
			continue
		}

		pages := 0
		if r.ProgressPages != nil && *r.ProgressPages > 0 {
			pages = *r.ProgressPages
		} else if r.Edition != nil && r.Edition.Pages != nil {
			pages = *r.Edition.Pages
		}

		format := "physical" // default
		if r.Edition != nil && r.Edition.EditionFormat != nil {
			f := *r.Edition.EditionFormat
			switch f {
			case "audiobook", "audio_cd", "audio_cassette":
				format = "audiobook"
			case "ebook", "kindle_edition":
				format = "ebook"
			default:
				format = "physical"
			}
		}

		if pages > 0 {
			result = append(result, api.ReadingHistoryEntry{
				FinishedAt:    *r.FinishedAt,
				Pages:         pages,
				EditionFormat: format,
			})
		}
	}
	return result, nil
}
