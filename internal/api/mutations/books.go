package mutations

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/NotMugil/hardcover-tui/internal/api"
)

// InsertUserBook adds a book to the user's library using gen.Client.
func InsertUserBook(ctx context.Context, c *api.Client, bookID, statusID int) (*api.UserBook, error) {
	res, err := c.Gen.InsertUserBook(ctx, bookID, statusID)
	if err != nil {
		return nil, fmt.Errorf("insert user_book: %w", err)
	}

	if res.InsertUserBook.ID == nil {
		errMsg := "unknown error"
		if res.InsertUserBook.Error != nil {
			errMsg = *res.InsertUserBook.Error
		}
		return nil, fmt.Errorf("insert user_book: %s", errMsg)
	}

	return &api.UserBook{
		ID:       *res.InsertUserBook.ID,
		BookID:   bookID,
		StatusID: statusID,
	}, nil
}

// UpdateUserBookStatus changes the reading status of a user book using gen.Client.
func UpdateUserBookStatus(ctx context.Context, c *api.Client, userBookID, statusID int) error {
	_, err := c.Gen.UpdateUserBookStatus(ctx, userBookID, statusID)
	return err
}

// UpdateUserBookRating updates the rating for a user book using gen.Client.
func UpdateUserBookRating(ctx context.Context, c *api.Client, userBookID int, rating float64) error {
	rawRating := json.RawMessage(fmt.Sprintf("%f", rating))
	_, err := c.Gen.UpdateUserBookRating(ctx, userBookID, rawRating)
	return err
}

// UpdateUserBookReview updates the review for a user book using gen.Client.
func UpdateUserBookReview(ctx context.Context, c *api.Client, userBookID int, review string, hasSpoilers bool) error {
	rawReview, _ := json.Marshal(review)
	_, err := c.Gen.UpdateUserBookReview(ctx, userBookID, rawReview, hasSpoilers)
	return err
}

// DeleteUserBook removes a book from the user's library using gen.Client.
func DeleteUserBook(ctx context.Context, c *api.Client, userBookID int) error {
	_, err := c.Gen.DeleteUserBook(ctx, userBookID)
	return err
}

// InsertUserBookRead creates a new read-through entry using gen.Client.
func InsertUserBookRead(ctx context.Context, c *api.Client, userBookID int, startedAt, finishedAt *string) error {
	var rawStarted json.RawMessage
	if startedAt != nil {
		rawStarted, _ = json.Marshal(*startedAt)
	}
	_, err := c.Gen.InsertUserBookRead(ctx, userBookID, rawStarted, nil)
	return err
}

// UpdateUserBookRead updates a read-through entry using gen.Client.
func UpdateUserBookRead(ctx context.Context, c *api.Client, readID int, progressPages *int) error {
	_, err := c.Gen.UpdateUserBookRead(ctx, readID, progressPages)
	return err
}

// UpdateUserBookReadDates updates started_at and finished_at on a read entry using gen.Client.
func UpdateUserBookReadDates(ctx context.Context, c *api.Client, readID int, startedAt, finishedAt *string) error {
	var rawStarted, rawFinished json.RawMessage
	if startedAt != nil {
		rawStarted, _ = json.Marshal(*startedAt)
	}
	if finishedAt != nil {
		rawFinished, _ = json.Marshal(*finishedAt)
	}
	_, err := c.Gen.UpdateUserBookReadDates(ctx, readID, rawStarted, rawFinished)
	return err
}
