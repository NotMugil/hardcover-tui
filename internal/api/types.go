package api

import (
	"encoding/json"
	"time"
)

// StatusID represents a reading status.
type StatusID int

const (
	StatusWantToRead       StatusID = 1
	StatusCurrentlyReading StatusID = 2
	StatusRead             StatusID = 3
	StatusPaused           StatusID = 4
	StatusDidNotFinish     StatusID = 5
	StatusIgnored          StatusID = 6
)

func (s StatusID) String() string {
	switch s {
	case StatusWantToRead:
		return "Want to Read"
	case StatusCurrentlyReading:
		return "Currently Reading"
	case StatusRead:
		return "Read"
	case StatusPaused:
		return "Paused"
	case StatusDidNotFinish:
		return "Did Not Finish"
	case StatusIgnored:
		return "Ignored"
	default:
		return "Unknown"
	}
}

// AllStatuses returns all reading statuses in order.
func AllStatuses() []StatusID {
	return []StatusID{
		StatusWantToRead,
		StatusCurrentlyReading,
		StatusRead,
		StatusPaused,
		StatusDidNotFinish,
		StatusIgnored,
	}
}

// PrivacySettingID represents privacy levels.
type PrivacySettingID int

const (
	PrivacyPublic    PrivacySettingID = 1
	PrivacyFollowers PrivacySettingID = 2
	PrivacyPrivate   PrivacySettingID = 3
)

func (p PrivacySettingID) String() string {
	switch p {
	case PrivacyPublic:
		return "Public"
	case PrivacyFollowers:
		return "Followers Only"
	case PrivacyPrivate:
		return "Private"
	default:
		return "Unknown"
	}
}

// AllPrivacySettings returns all privacy settings in order.
func AllPrivacySettings() []PrivacySettingID {
	return []PrivacySettingID{PrivacyPublic, PrivacyFollowers, PrivacyPrivate}
}

// Image represents a Hardcover image.
type Image struct {
	URL string `json:"url" graphql:"url"`
}

// Author represents a book author.
type Author struct {
	ID   int    `json:"id" graphql:"id"`
	Name string `json:"name" graphql:"name"`
	Slug string `json:"slug" graphql:"slug"`
}

// Contribution represents an author's contribution to a book.
type Contribution struct {
	Author Author `json:"author" graphql:"author"`
}

// TagItem represents a genre, mood, or content warning tag.
type TagItem struct {
	Name string `json:"name" graphql:"tag"`
}

// Book represents a Hardcover book.
type Book struct {
	ID             int            `json:"id" graphql:"id"`
	Title          string         `json:"title" graphql:"title"`
	Subtitle       *string        `json:"subtitle" graphql:"subtitle"`
	Description    *string        `json:"description" graphql:"description"`
	Pages          *int           `json:"pages" graphql:"pages"`
	Rating         *float64       `json:"rating" graphql:"rating"`
	RatingsCount   int            `json:"ratings_count" graphql:"ratings_count"`
	ReviewsCount   int            `json:"reviews_count" graphql:"reviews_count"`
	UsersCount     int            `json:"users_count" graphql:"users_count"`
	ReleaseYear    *int           `json:"release_year" graphql:"release_year"`
	ReleaseDate    *string        `json:"release_date" graphql:"release_date"`
	Slug           *string        `json:"slug" graphql:"slug"`
	Contributions  []Contribution `json:"contributions" graphql:"contributions"`
	Image          *Image         `json:"image" graphql:"image"`
	AudioSeconds   *int           `json:"audio_seconds" graphql:"audio_seconds"`
	BookCategoryID int            `json:"book_category_id" graphql:"book_category_id"`
	LiteraryTypeID *int           `json:"literary_type_id" graphql:"literary_type_id"`
	HasAudiobook   bool           `json:"-"`
	HasEbook       bool           `json:"-"`
	Genres         []TagItem      `json:"-"`
}

// FormatIndicator returns a short string indicating available formats.
func (b Book) FormatIndicator() string {
	var parts []string
	if b.Pages != nil && *b.Pages > 0 {
		parts = append(parts, "P")
	}
	if b.HasEbook {
		parts = append(parts, "E")
	}
	if b.HasAudiobook {
		parts = append(parts, "A")
	}
	if len(parts) == 0 {
		return "-"
	}
	result := ""
	for i, p := range parts {
		if i > 0 {
			result += "/"
		}
		result += p
	}
	return result
}

// Authors returns a comma-separated string of author names.
func (b Book) Authors() string {
	if len(b.Contributions) == 0 {
		return "Unknown"
	}
	names := ""
	for i, c := range b.Contributions {
		if i > 0 {
			names += ", "
		}
		names += c.Author.Name
	}
	return names
}

// CoverURL returns the book cover URL, or empty string if none.
func (b Book) CoverURL() string {
	if b.Image != nil {
		return b.Image.URL
	}
	return ""
}

// UserBookRead represents a single read-through of a book.
type UserBookRead struct {
	ID              int     `json:"id" graphql:"id"`
	StartedAt       *string `json:"started_at" graphql:"started_at"`
	FinishedAt      *string `json:"finished_at" graphql:"finished_at"`
	ProgressPages   *int    `json:"progress_pages" graphql:"progress_pages"`
	ProgressSeconds *int    `json:"progress_seconds" graphql:"progress_seconds"`
	EditionID       *int    `json:"edition_id" graphql:"edition_id"`
}

// UserBook represents the relationship between a user and a book.
type UserBook struct {
	ID                      int            `json:"id" graphql:"id"`
	BookID                  int            `json:"book_id" graphql:"book_id"`
	StatusID                int            `json:"status_id" graphql:"status_id"`
	Rating                  *float64       `json:"rating" graphql:"rating"`
	Review                  *string        `json:"review" graphql:"review"`
	ReviewHTML              *string        `json:"review_html" graphql:"review_html"`
	ReviewHasSpoilers       bool           `json:"review_has_spoilers" graphql:"review_has_spoilers"`
	HasReview               bool           `json:"has_review" graphql:"has_review"`
	DateAdded               string         `json:"date_added" graphql:"date_added"`
	FirstReadDate           *string        `json:"first_read_date" graphql:"first_read_date"`
	LastReadDate            *string        `json:"last_read_date" graphql:"last_read_date"`
	FirstStartedReadingDate *string        `json:"first_started_reading_date" graphql:"first_started_reading_date"`
	ReadCount               int            `json:"read_count" graphql:"read_count"`
	Owned                   bool           `json:"owned" graphql:"owned"`
	PrivacySettingID        int            `json:"privacy_setting_id" graphql:"privacy_setting_id"`
	PrivateNotes            *string        `json:"private_notes" graphql:"private_notes"`
	Starred                 bool           `json:"starred" graphql:"starred"`
	EditionID               *int           `json:"edition_id" graphql:"edition_id"`
	LikesCount              int            `json:"likes_count" graphql:"likes_count"`
	CreatedAt               string         `json:"created_at" graphql:"created_at"`
	UpdatedAt               *string        `json:"updated_at" graphql:"updated_at"`
	Book                    Book           `json:"book" graphql:"book"`
	UserBookReads           []UserBookRead `json:"user_book_reads" graphql:"user_book_reads"`
}

// Status returns the StatusID enum for this user book.
func (ub UserBook) Status() StatusID {
	return StatusID(ub.StatusID)
}

// User represents a Hardcover user.
type User struct {
	ID                 int       `json:"id" graphql:"id"`
	Username           string    `json:"username" graphql:"username"`
	Name               *string   `json:"name" graphql:"name"`
	Bio                *string   `json:"bio" graphql:"bio"`
	Location           *string   `json:"location" graphql:"location"`
	Link               *string   `json:"link" graphql:"link"`
	Flair              *string   `json:"flair" graphql:"flair"`
	BooksCount         int       `json:"books_count" graphql:"books_count"`
	FollowersCount     int       `json:"followers_count" graphql:"followers_count"`
	FollowedUsersCount int       `json:"followed_users_count" graphql:"followed_users_count"`
	Pro                bool      `json:"pro" graphql:"pro"`
	PronounPersonal    string    `json:"pronoun_personal" graphql:"pronoun_personal"`
	PronounPossessive  string    `json:"pronoun_possessive" graphql:"pronoun_possessive"`
	Image              *Image    `json:"image" graphql:"image"`
	CreatedAt          time.Time `json:"created_at" graphql:"created_at"`
}

// DisplayName returns the user's display name, falling back to username.
func (u User) DisplayName() string {
	if u.Name != nil && *u.Name != "" {
		return *u.Name
	}
	return u.Username
}

// ImageURL returns the user's profile image URL, or empty string if none.
func (u User) ImageURL() string {
	if u.Image != nil {
		return u.Image.URL
	}
	return ""
}

// List represents a user-created book list.
type List struct {
	ID               int     `json:"id" graphql:"id"`
	Name             string  `json:"name" graphql:"name"`
	Description      *string `json:"description" graphql:"description"`
	BooksCount       int     `json:"books_count" graphql:"books_count"`
	LikesCount       int     `json:"likes_count" graphql:"likes_count"`
	FollowersCount   *int    `json:"followers_count" graphql:"followers_count"`
	Public           bool    `json:"public" graphql:"public"`
	Ranked           bool    `json:"ranked" graphql:"ranked"`
	PrivacySettingID int     `json:"privacy_setting_id" graphql:"privacy_setting_id"`
	Slug             *string `json:"slug" graphql:"slug"`
	UserID           int     `json:"user_id" graphql:"user_id"`
	CreatedAt        *string `json:"created_at" graphql:"created_at"`
	UpdatedAt        *string `json:"updated_at" graphql:"updated_at"`
}

// ListBook represents a book entry within a list.
type ListBook struct {
	ID        int     `json:"id" graphql:"id"`
	ListID    int     `json:"list_id" graphql:"list_id"`
	BookID    int     `json:"book_id" graphql:"book_id"`
	Position  *int    `json:"position" graphql:"position"`
	DateAdded *string `json:"date_added" graphql:"date_added"`
	Book      Book    `json:"book" graphql:"book"`
}

// ReadingJournal represents a reading journal entry.
type ReadingJournal struct {
	ID               int     `json:"id" graphql:"id"`
	Event            string  `json:"event" graphql:"event"`
	Entry            *string `json:"entry" graphql:"entry"`
	ActionAt         string  `json:"action_at" graphql:"action_at"`
	BookID           *int    `json:"book_id" graphql:"book_id"`
	EditionID        *int    `json:"edition_id" graphql:"edition_id"`
	PrivacySettingID int     `json:"privacy_setting_id" graphql:"privacy_setting_id"`
	LikesCount       int     `json:"likes_count" graphql:"likes_count"`
	CreatedAt        string  `json:"created_at" graphql:"created_at"`
	UpdatedAt        string  `json:"updated_at" graphql:"updated_at"`
	Book             *Book   `json:"book" graphql:"book"`
}

// Goal represents a reading goal.
type Goal struct {
	ID               int     `json:"id" graphql:"id"`
	Goal             int     `json:"goal" graphql:"goal"`
	Metric           string  `json:"metric" graphql:"metric"`
	Progress         float64 `json:"progress" graphql:"progress"`
	StartDate        string  `json:"start_date" graphql:"start_date"`
	EndDate          string  `json:"end_date" graphql:"end_date"`
	State            string  `json:"state" graphql:"state"`
	Description      *string `json:"description" graphql:"description"`
	Archived         bool    `json:"archived" graphql:"archived"`
	CompletedAt      *string `json:"completed_at" graphql:"completed_at"`
	PrivacySettingID *int    `json:"privacy_setting_id" graphql:"privacy_setting_id"`
	UserID           int     `json:"user_id" graphql:"user_id"`
}

// PercentComplete returns the goal's progress as a percentage (0.0-1.0).
func (g Goal) PercentComplete() float64 {
	if g.Goal == 0 {
		return 0
	}
	pct := g.Progress / float64(g.Goal)
	if pct > 1 {
		return 1
	}
	return pct
}

// DisplayName returns a human-readable name for the goal.
func (g Goal) DisplayName() string {
	if g.Description != nil && *g.Description != "" {
		return *g.Description
	}
	year := "????"
	if len(g.StartDate) >= 4 {
		year = g.StartDate[:4]
	}
	return year + " Reading Goal"
}

// DaysRemaining returns the number of days until the goal's end date.
// Returns 0 if the end date is in the past or can't be parsed.
func (g Goal) DaysRemaining() int {
	end, err := time.Parse("2006-01-02", g.EndDate)
	if err != nil {
		return 0
	}
	days := int(time.Until(end).Hours() / 24)
	if days < 0 {
		return 0
	}
	return days
}

// Remaining returns goal - progress, floored at 0.
func (g Goal) Remaining() int {
	r := g.Goal - int(g.Progress)
	if r < 0 {
		return 0
	}
	return r
}

// Activity represents a user activity event.
type Activity struct {
	ID               int              `json:"id" graphql:"id"`
	Event            string           `json:"event" graphql:"event"`
	Data             json.RawMessage  `json:"data" graphql:"data"`
	BookID           *int             `json:"book_id" graphql:"book_id"`
	LikesCount       int              `json:"likes_count" graphql:"likes_count"`
	PrivacySettingID int              `json:"privacy_setting_id" graphql:"privacy_setting_id"`
	CreatedAt        string           `json:"created_at" graphql:"created_at"`
	Book             *Book            `json:"book" graphql:"book"`
	User             *ActivityUser    `json:"user" graphql:"user"`
}

// ActivityDataUserBook holds parsed data for UserBookActivity events.
type ActivityDataUserBook struct {
	Rating   *string `json:"rating"`
	Review   *string `json:"review"`
	StatusID *int    `json:"statusId"`
}

// ActivityDataGoal holds parsed data for GoalActivity events.
type ActivityDataGoal struct {
	Goal            int     `json:"goal"`
	Metric          string  `json:"metric"`
	Progress        float64 `json:"progress"`
	PercentComplete float64 `json:"percentComplete"`
	Description     string  `json:"description"`
}

// ActivityDataList holds parsed data for ListActivity events.
type ActivityDataList struct {
	Name       string `json:"name"`
	BooksCount int    `json:"booksCount"`
}

// ActivityDataPrompt holds parsed data for PromptActivity events.
type ActivityDataPrompt struct {
	Question string `json:"question"`
}

// ActivityParsedData contains the parsed data payload of an activity.
type ActivityParsedData struct {
	UserBook *ActivityDataUserBook `json:"userBook,omitempty"`
	Goal     *ActivityDataGoal     `json:"goal,omitempty"`
	List     *ActivityDataList     `json:"list,omitempty"`
	Prompt   *ActivityDataPrompt   `json:"prompt,omitempty"`
}

// ParseData parses the JSON data field into a structured type.
func (a Activity) ParseData() ActivityParsedData {
	var d ActivityParsedData
	if len(a.Data) > 0 {
		_ = json.Unmarshal(a.Data, &d)
	}
	return d
}

// ActivityUser is a lightweight user attached to an activity.
type ActivityUser struct {
	ID       int     `json:"id" graphql:"id"`
	Username string  `json:"username" graphql:"username"`
	Name     *string `json:"name" graphql:"name"`
}

// DisplayName returns the activity user's display name.
func (u ActivityUser) DisplayName() string {
	if u.Name != nil && *u.Name != "" {
		return *u.Name
	}
	return u.Username
}

// UserBookAggregate represents aggregate stats for user books.
type UserBookAggregate struct {
	Aggregate struct {
		Count int `json:"count" graphql:"count"`
		Avg   struct {
			Rating *float64 `json:"rating" graphql:"rating"`
		} `json:"avg" graphql:"avg"`
	} `json:"aggregate" graphql:"aggregate"`
}

// BookReview represents a community review of a book.
type BookReview struct {
	ID                int        `json:"id"`
	Rating            *float64   `json:"rating"`
	Review            *string    `json:"review"`
	ReviewHasSpoilers bool       `json:"review_has_spoilers"`
	LikesCount        int        `json:"likes_count"`
	CreatedAt         string     `json:"created_at"`
	User              ReviewUser `json:"user"`
}

// ReviewUser represents basic user info attached to a review.
type ReviewUser struct {
	ID       int     `json:"id"`
	Username string  `json:"username"`
	Name     *string `json:"name"`
}

// DisplayName returns the reviewer's display name or username.
func (u ReviewUser) DisplayName() string {
	if u.Name != nil && *u.Name != "" {
		return *u.Name
	}
	return u.Username
}

// StatsUserBook is a lightweight user_book for stats queries.
type StatsUserBook struct {
	StatusID       int
	LiteraryTypeID *int
	EditionFormat  *string
	Genres         []string
}

// LiteraryType constants.
const (
	LiteraryTypeFiction    = 1
	LiteraryTypeNonfiction = 2
)

// Tag category IDs from the Hardcover API.
const (
	TagCategoryGenre          = 1
	TagCategoryContentWarning = 3
	TagCategoryMood           = 5
)

// ReadingHistoryEntry represents a single reading entry for time-series stats.
type ReadingHistoryEntry struct {
	FinishedAt    string // date the read was finished
	Pages         int    // number of pages read/listened
	EditionFormat string // "physical", "ebook", "audiobook", etc.
}
