package mutations

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/NotMugil/hardcover-tui/internal/api"
	"github.com/NotMugil/hardcover-tui/internal/api/gen"
)

// InsertList creates a new list using gen.Client.
func InsertList(ctx context.Context, c *api.Client, name, description string) (*api.List, error) {
	res, err := c.Gen.InsertList(ctx, name, &description, 1)
	if err != nil {
		return nil, fmt.Errorf("insert list: %w", err)
	}

	if res.InsertList.ID == nil {
		return nil, fmt.Errorf("insert list: failed")
	}

	return &api.List{
		ID:   *res.InsertList.ID,
		Name: name,
	}, nil
}

// UpdateList modifies an existing list using gen.Client.
func UpdateList(ctx context.Context, c *api.Client, listID int, name, description string, privacySettingID int) error {
	_, err := c.Gen.UpdateList(ctx, listID, name, &description, privacySettingID)
	return err
}

// DeleteList removes a list using gen.Client.
func DeleteList(ctx context.Context, c *api.Client, listID int) error {
	_, err := c.Gen.DeleteList(ctx, listID)
	return err
}

// InsertListBook adds a book to a list using gen.Client.
func InsertListBook(ctx context.Context, c *api.Client, listID, bookID int) error {
	_, err := c.Gen.InsertListBook(ctx, listID, bookID, 0)
	return err
}

// DeleteListBook removes a book from a list using gen.Client.
func DeleteListBook(ctx context.Context, c *api.Client, listBookID int) error {
	_, err := c.Gen.DeleteListBook(ctx, listBookID)
	return err
}

// InsertReadingJournal creates a new journal entry using gen.Client.
func InsertReadingJournal(ctx context.Context, c *api.Client, bookID int, event, entry, actionAt string) error {
	var rawAction json.RawMessage
	if actionAt != "" {
		rawAction, _ = json.Marshal(actionAt)
	}
	var tags []*gen.BasicTag
	_, err := c.Gen.InsertReadingJournal(ctx, bookID, event, &entry, rawAction, 1, tags)
	return err
}

// UpdateReadingJournal updates an existing journal entry using gen.Client.
func UpdateReadingJournal(ctx context.Context, c *api.Client, journalID int, entry string) error {
	_, err := c.Gen.UpdateReadingJournal(ctx, journalID, &entry)
	return err
}

// DeleteReadingJournal removes a journal entry using gen.Client.
func DeleteReadingJournal(ctx context.Context, c *api.Client, journalID int) error {
	_, err := c.Gen.DeleteReadingJournal(ctx, journalID)
	return err
}

// UpdateUserProfile updates the user's profile fields using gen.Client.
func UpdateUserProfile(ctx context.Context, c *api.Client, name, bio, location string) error {
	_, err := c.Gen.UpdateUserProfile(ctx, &name, &bio, &location)
	return err
}
