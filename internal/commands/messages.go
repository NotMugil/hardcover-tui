package commands

import "github.com/NotMugil/hardcover-tui/internal/api"

// Book loading

type BookLoadedMsg struct {
	UserBook *api.UserBook
	Err      error
}

type BookFromBookIDMsg struct {
	Book     *api.Book
	UserBook *api.UserBook
	Err      error
}

// Book mutations

type StatusUpdatedMsg struct {
	Err error
}

type RatingUpdatedMsg struct {
	Err error
}

type BookAddedMsg struct {
	UserBook *api.UserBook
	Err      error
}

// Cover & tags & reviews

type CoverLoadedMsg struct {
	Art string
}

type TagsLoadedMsg struct {
	Genres []api.TagItem
	Err    error
}

type ReviewsLoadedMsg struct {
	Reviews []api.BookReview
	Err     error
}

// Journals

type JournalsLoadedMsg struct {
	Journals []api.ReadingJournal
	Err      error
}

type JournalSavedMsg struct {
	Err error
}

type JournalDeletedMsg struct {
	Err error
}

// Lists

type UserListsLoadedMsg struct {
	Lists []api.List
	Err   error
}

type ListBooksLoadedMsg struct {
	Books []api.ListBook
	Err   error
}

type ListCreatedMsg struct {
	List *api.List
	Err  error
}

type ListDeletedMsg struct {
	Err error
}

type BookAddedToListMsg struct {
	ListName string
	Err      error
}

type BookRemovedFromListMsg struct {
	Err error
}

type PrivacyUpdatedMsg struct {
	Err error
}

// Search

type SearchResultsMsg struct {
	Books []api.Book
	Err   error
}

// Stats

type StatsLoadedMsg struct {
	Goals          []api.Goal
	Counts         map[api.StatusID]int
	UserBooks      []api.StatsUserBook
	ReadingHistory []api.ReadingHistoryEntry
	Err            error
}

// Review

type ReviewSavedMsg struct {
	Err error
}

// Progress

type ProgressUpdatedMsg struct {
	Err error
}

// Profile

type ProfileUpdatedMsg struct {
	Err error
}

// Goals

type GoalsLoadedMsg struct {
	Goals []api.Goal
	Err   error
}
