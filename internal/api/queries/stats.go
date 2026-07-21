package queries

import (
	"context"
	"fmt"

	"github.com/NotMugil/hardcover-tui/internal/api"
)

// GetUserBooksForStats fetches user books with book metadata for stats using gen.Client.
func GetUserBooksForStats(ctx context.Context, c *api.Client, userID int) ([]api.StatsUserBook, error) {
	res, err := c.Gen.GetUserBooksForStats(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get user books for stats: %w", err)
	}

	result := make([]api.StatsUserBook, len(res.UserBooks))
	for i, ub := range res.UserBooks {
		sub := api.StatsUserBook{
			StatusID:       ub.StatusID,
			LiteraryTypeID: ub.Book.LiteraryTypeID,
		}

		for _, t := range ub.Book.Taggings {
			if t.Tag.TagCategoryID == api.TagCategoryGenre {
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

// GetReadingHistory fetches finished reads with dates, pages, and edition format for time-series charts using gen.Client.
func GetReadingHistory(ctx context.Context, c *api.Client, userID int) ([]api.ReadingHistoryEntry, error) {
	res, err := c.Gen.GetReadingHistory(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get reading history: %w", err)
	}

	var result []api.ReadingHistoryEntry
	for _, r := range res.UserBookReads {
		finAt := parseRawStringPtr(&r.FinishedAt)
		if finAt == nil {
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
				FinishedAt:    *finAt,
				Pages:         pages,
				EditionFormat: format,
			})
		}
	}
	return result, nil
}
