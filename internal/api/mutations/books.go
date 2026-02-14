package mutations

import (
	"context"
	"fmt"

	graphql "github.com/hasura/go-graphql-client"

	"github.com/NotMugil/hardcover-tui/internal/api"
)

// Numeric is a custom GraphQL scalar that maps to Hardcover's "numeric" type.
// The standard graphql.Float maps to "Float" which causes type mismatch errors.
type Numeric float64

// GetGraphQLType implements the go-graphql-client GraphQLType interface.
func (n Numeric) GetGraphQLType() string { return "numeric" }

// Date is a custom GraphQL scalar that maps to Hardcover's "date" type.
// The standard graphql.String maps to "String" which causes type mismatch errors.
type Date string

// GetGraphQLType implements the go-graphql-client GraphQLType interface.
func (d Date) GetGraphQLType() string { return "date" }

// InsertUserBook adds a book to the user's library.
func InsertUserBook(ctx context.Context, c *api.Client, bookID, statusID int) (*api.UserBook, error) {
	var m struct {
		InsertUserBook struct {
			ID    *int    `graphql:"id"`
			Error *string `graphql:"error"`
		} `graphql:"insert_user_book(object: {book_id: $bookId, status_id: $statusId})"`
	}

	vars := map[string]interface{}{
		"bookId":   graphql.Int(bookID),
		"statusId": graphql.Int(statusID),
	}

	if err := c.Mutate(ctx, &m, vars); err != nil {
		return nil, fmt.Errorf("insert user_book: %w", err)
	}

	if m.InsertUserBook.ID == nil {
		errMsg := "unknown error"
		if m.InsertUserBook.Error != nil {
			errMsg = *m.InsertUserBook.Error
		}
		return nil, fmt.Errorf("insert user_book: %s", errMsg)
	}

	return &api.UserBook{
		ID:       *m.InsertUserBook.ID,
		BookID:   bookID,
		StatusID: statusID,
	}, nil
}

// UpdateUserBookStatus changes the reading status of a user book.
func UpdateUserBookStatus(ctx context.Context, c *api.Client, userBookID, statusID int) error {
	var m struct {
		UpdateUserBook struct {
			ID    *int    `graphql:"id"`
			Error *string `graphql:"error"`
		} `graphql:"update_user_book(id: $id, object: {status_id: $statusId})"`
	}

	vars := map[string]interface{}{
		"id":       graphql.Int(userBookID),
		"statusId": graphql.Int(statusID),
	}

	return c.Mutate(ctx, &m, vars)
}

// UpdateUserBookRating updates the rating for a user book.
func UpdateUserBookRating(ctx context.Context, c *api.Client, userBookID int, rating float64) error {
	var m struct {
		UpdateUserBook struct {
			ID    *int    `graphql:"id"`
			Error *string `graphql:"error"`
		} `graphql:"update_user_book(id: $id, object: {rating: $rating})"`
	}

	vars := map[string]interface{}{
		"id":     graphql.Int(userBookID),
		"rating": Numeric(rating),
	}

	return c.Mutate(ctx, &m, vars)
}

// UpdateUserBookReview updates the review for a user book.
func UpdateUserBookReview(ctx context.Context, c *api.Client, userBookID int, review string, hasSpoilers bool) error {
	var m struct {
		UpdateUserBook struct {
			ID    *int    `graphql:"id"`
			Error *string `graphql:"error"`
		} `graphql:"update_user_book(id: $id, object: {review_raw: $review, review_has_spoilers: $spoilers})"`
	}

	vars := map[string]interface{}{
		"id":       graphql.Int(userBookID),
		"review":   graphql.String(review),
		"spoilers": graphql.Boolean(hasSpoilers),
	}

	return c.Mutate(ctx, &m, vars)
}

// DeleteUserBook removes a book from the user's library.
func DeleteUserBook(ctx context.Context, c *api.Client, userBookID int) error {
	var m struct {
		DeleteUserBook struct {
			ID *int `graphql:"id"`
		} `graphql:"delete_user_book(id: $id)"`
	}

	vars := map[string]interface{}{
		"id": graphql.Int(userBookID),
	}

	return c.Mutate(ctx, &m, vars)
}

// InsertUserBookRead creates a new read-through entry.
func InsertUserBookRead(ctx context.Context, c *api.Client, userBookID int, startedAt, finishedAt *string) error {
	var m struct {
		InsertUserBookRead struct {
			ID    *int    `graphql:"id"`
			Error *string `graphql:"error"`
		} `graphql:"insert_user_book_read(user_book_id: $userBookId, user_book_read: {started_at: $startedAt, progress_pages: $progressPages})"`
	}

	vars := map[string]interface{}{
		"userBookId":    graphql.Int(userBookID),
		"startedAt":     (*graphql.String)(nil),
		"progressPages": (*graphql.Int)(nil),
	}
	if startedAt != nil {
		s := graphql.String(*startedAt)
		vars["startedAt"] = &s
	}

	return c.Mutate(ctx, &m, vars)
}

// UpdateUserBookRead updates a read-through entry.
func UpdateUserBookRead(ctx context.Context, c *api.Client, readID int, progressPages *int) error {
	var m struct {
		UpdateUserBookRead struct {
			ID    *int    `graphql:"id"`
			Error *string `graphql:"error"`
		} `graphql:"update_user_book_read(id: $id, object: {progress_pages: $progressPages})"`
	}

	vars := map[string]interface{}{
		"id":            graphql.Int(readID),
		"progressPages": (*graphql.Int)(nil),
	}
	if progressPages != nil {
		p := graphql.Int(*progressPages)
		vars["progressPages"] = &p
	}

	return c.Mutate(ctx, &m, vars)
}

// UpdateUserBookReadDates updates started_at and finished_at on a read entry.
func UpdateUserBookReadDates(ctx context.Context, c *api.Client, readID int, startedAt, finishedAt *string) error {
	var m struct {
		UpdateUserBookRead struct {
			ID    *int    `graphql:"id"`
			Error *string `graphql:"error"`
		} `graphql:"update_user_book_read(id: $id, object: {started_at: $startedAt, finished_at: $finishedAt})"`
	}

	vars := map[string]interface{}{
		"id":         graphql.Int(readID),
		"startedAt":  (*Date)(nil),
		"finishedAt": (*Date)(nil),
	}
	if startedAt != nil {
		s := Date(*startedAt)
		vars["startedAt"] = &s
	}
	if finishedAt != nil {
		f := Date(*finishedAt)
		vars["finishedAt"] = &f
	}

	return c.Mutate(ctx, &m, vars)
}
