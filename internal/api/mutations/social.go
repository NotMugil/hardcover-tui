package mutations

import (
	"context"
	"fmt"

	graphql "github.com/hasura/go-graphql-client"

	"hardcover-tui/internal/api"
)

// InsertList creates a new list.
func InsertList(ctx context.Context, c *api.Client, name, description string) (*api.List, error) {
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

	return &api.List{
		ID:   *m.InsertList.ID,
		Name: name,
	}, nil
}

// UpdateList modifies an existing list.
func UpdateList(ctx context.Context, c *api.Client, listID int, name, description string, privacySettingID int) error {
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
func DeleteList(ctx context.Context, c *api.Client, listID int) error {
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
func InsertListBook(ctx context.Context, c *api.Client, listID, bookID int) error {
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
func DeleteListBook(ctx context.Context, c *api.Client, listBookID int) error {
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
func InsertReadingJournal(ctx context.Context, c *api.Client, bookID int, event, entry, actionAt string) error {
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
func UpdateReadingJournal(ctx context.Context, c *api.Client, journalID int, entry string) error {
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
func DeleteReadingJournal(ctx context.Context, c *api.Client, journalID int) error {
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
func UpdateUserProfile(ctx context.Context, c *api.Client, name, bio, location string) error {
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
