package api

import (
	"context"
	"fmt"

	graphql "github.com/hasura/go-graphql-client"
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
func (c *Client) InsertUserBook(ctx context.Context, bookID, statusID int) (*UserBook, error) {
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

	return &UserBook{
		ID:       *m.InsertUserBook.ID,
		BookID:   bookID,
		StatusID: statusID,
	}, nil
}

// UpdateUserBookStatus changes the reading status of a user book.
func (c *Client) UpdateUserBookStatus(ctx context.Context, userBookID, statusID int) error {
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
func (c *Client) UpdateUserBookRating(ctx context.Context, userBookID int, rating float64) error {
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
func (c *Client) UpdateUserBookReview(ctx context.Context, userBookID int, review string, hasSpoilers bool) error {
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
func (c *Client) DeleteUserBook(ctx context.Context, userBookID int) error {
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
func (c *Client) InsertUserBookRead(ctx context.Context, userBookID int, startedAt, finishedAt *string) error {
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
func (c *Client) UpdateUserBookRead(ctx context.Context, readID int, progressPages *int) error {
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
func (c *Client) UpdateUserBookReadDates(ctx context.Context, readID int, startedAt, finishedAt *string) error {
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

// InsertList creates a new list.
func (c *Client) InsertList(ctx context.Context, name, description string) (*List, error) {
	var m struct {
		InsertList struct {
			ID     *int      `graphql:"id"`
			Errors *[]string `graphql:"errors"`
		} `graphql:"insert_list(object: {name: $name, description: $description, privacy_setting_id: $privacySettingId})"`
	}

	vars := map[string]interface{}{
		"name":             graphql.String(name),
		"description":      graphql.String(description),
		"privacySettingId": graphql.Int(1),
	}

	if err := c.Mutate(ctx, &m, vars); err != nil {
		return nil, fmt.Errorf("insert list: %w", err)
	}

	if m.InsertList.ID == nil {
		return nil, fmt.Errorf("insert list: failed")
	}

	return &List{
		ID:   *m.InsertList.ID,
		Name: name,
	}, nil
}

// UpdateList modifies an existing list.
func (c *Client) UpdateList(ctx context.Context, listID int, name, description string, privacySettingID int) error {
	var m struct {
		UpdateList struct {
			ID     *int      `graphql:"id"`
			Errors *[]string `graphql:"errors"`
		} `graphql:"update_list(id: $id, object: {name: $name, description: $description, privacy_setting_id: $privacySettingId})"`
	}

	vars := map[string]interface{}{
		"id":               graphql.Int(listID),
		"name":             graphql.String(name),
		"description":      graphql.String(description),
		"privacySettingId": graphql.Int(privacySettingID),
	}

	return c.Mutate(ctx, &m, vars)
}

// DeleteList removes a list.
func (c *Client) DeleteList(ctx context.Context, listID int) error {
	var m struct {
		DeleteList struct {
			Success bool `graphql:"success"`
		} `graphql:"delete_list(id: $id)"`
	}

	vars := map[string]interface{}{
		"id": graphql.Int(listID),
	}

	return c.Mutate(ctx, &m, vars)
}

// InsertListBook adds a book to a list.
func (c *Client) InsertListBook(ctx context.Context, listID, bookID int) error {
	var m struct {
		InsertListBook struct {
			ID *int `graphql:"id"`
		} `graphql:"insert_list_book(object: {list_id: $listId, book_id: $bookId, position: $position})"`
	}

	vars := map[string]interface{}{
		"listId":   graphql.Int(listID),
		"bookId":   graphql.Int(bookID),
		"position": graphql.Int(0),
	}

	return c.Mutate(ctx, &m, vars)
}

// DeleteListBook removes a book from a list.
func (c *Client) DeleteListBook(ctx context.Context, listBookID int) error {
	var m struct {
		DeleteListBook struct {
			ID *int `graphql:"id"`
		} `graphql:"delete_list_book(id: $id)"`
	}

	vars := map[string]interface{}{
		"id": graphql.Int(listBookID),
	}

	return c.Mutate(ctx, &m, vars)
}

// InsertReadingJournal creates a new journal entry.
func (c *Client) InsertReadingJournal(ctx context.Context, bookID int, event, entry, actionAt string) error {
	var m struct {
		InsertReadingJournal struct {
			ID     *int      `graphql:"id"`
			Errors *[]string `graphql:"errors"`
		} `graphql:"insert_reading_journal(object: {book_id: $bookId, event: $event, entry: $entry, privacy_setting_id: $privacySettingId, tags: []})"`
	}

	vars := map[string]interface{}{
		"bookId":           graphql.Int(bookID),
		"event":            graphql.String(event),
		"entry":            graphql.String(entry),
		"privacySettingId": graphql.Int(1),
	}

	return c.Mutate(ctx, &m, vars)
}

// UpdateReadingJournal updates an existing journal entry.
func (c *Client) UpdateReadingJournal(ctx context.Context, journalID int, entry string) error {
	var m struct {
		UpdateReadingJournal struct {
			ID     *int      `graphql:"id"`
			Errors *[]string `graphql:"errors"`
		} `graphql:"update_reading_journal(id: $id, object: {entry: $entry})"`
	}

	vars := map[string]interface{}{
		"id":    graphql.Int(journalID),
		"entry": graphql.String(entry),
	}

	return c.Mutate(ctx, &m, vars)
}

// DeleteReadingJournal removes a journal entry.
func (c *Client) DeleteReadingJournal(ctx context.Context, journalID int) error {
	var m struct {
		DeleteReadingJournal struct {
			ID int `graphql:"id"`
		} `graphql:"delete_reading_journal(id: $id)"`
	}

	vars := map[string]interface{}{
		"id": graphql.Int(journalID),
	}

	return c.Mutate(ctx, &m, vars)
}

// UpdateUserProfile updates the user's profile fields.
func (c *Client) UpdateUserProfile(ctx context.Context, name, bio, location string) error {
	var m struct {
		UpdateUser struct {
			ID     *int      `graphql:"id"`
			Errors *[]string `graphql:"errors"`
		} `graphql:"update_user(user: {name: $name, bio: $bio, location: $location})"`
	}

	vars := map[string]interface{}{
		"name":     graphql.String(name),
		"bio":      graphql.String(bio),
		"location": graphql.String(location),
	}

	return c.Mutate(ctx, &m, vars)
}
