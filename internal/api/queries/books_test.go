package queries

import (
	"testing"
)

func TestParseSearchResults_FlatHits(t *testing.T) {
	raw := []byte(`{
		"search": {
			"results": {
				"hits": [
					{
						"document": {
							"id": "101",
							"title": "The Burning God",
							"slug": "the-burning-god",
							"author_names": ["R. F. Kuang"],
							"pages": "640",
							"rating": 4.18,
							"users_count": 12000,
							"release_year": 2020,
							"description": "The conclusion to the Poppy War trilogy.",
							"has_audiobook": true,
							"has_ebook": true
						}
					}
				]
			}
		}
	}`)

	books, err := ParseSearchResults(raw)
	if err != nil {
		t.Fatalf("ParseSearchResults failed: %v", err)
	}

	if len(books) != 1 {
		t.Fatalf("expected 1 book, got %d", len(books))
	}

	b := books[0]
	if b.ID != 101 {
		t.Errorf("expected ID 101, got %d", b.ID)
	}
	if b.Title != "The Burning God" {
		t.Errorf("expected Title 'The Burning God', got %q", b.Title)
	}
	if b.Pages == nil || *b.Pages != 640 {
		t.Errorf("expected Pages 640, got %v", b.Pages)
	}
	if len(b.Contributions) != 1 || b.Contributions[0].Author.Name != "R. F. Kuang" {
		t.Errorf("expected author R. F. Kuang, got %v", b.Contributions)
	}
}

func TestParseSearchResults_StringifiedJSON(t *testing.T) {
	raw := []byte(`{
		"search": {
			"results": "{\"hits\":[{\"document\":{\"id\":202,\"title\":\"Dune\",\"slug\":\"dune\",\"author_names\":[\"Frank Herbert\"],\"pages\":412,\"rating\":4.24,\"users_count\":250000}}]}"
		}
	}`)

	books, err := ParseSearchResults(raw)
	if err != nil {
		t.Fatalf("ParseSearchResults failed: %v", err)
	}

	if len(books) != 1 {
		t.Fatalf("expected 1 book, got %d", len(books))
	}

	b := books[0]
	if b.ID != 202 {
		t.Errorf("expected ID 202, got %d", b.ID)
	}
	if b.Title != "Dune" {
		t.Errorf("expected Title 'Dune', got %q", b.Title)
	}
}

func TestParseSearchResults_EmptyOrNull(t *testing.T) {
	testCases := [][]byte{
		[]byte(`{"search":{"results":null}}`),
		[]byte(`{"search":{"results":{"hits":[]}}}`),
		[]byte(`{"search":{"results":""}}`),
	}

	for i, tc := range testCases {
		books, err := ParseSearchResults(tc)
		if err != nil {
			t.Errorf("case %d: unexpected error: %v", i, err)
		}
		if len(books) != 0 {
			t.Errorf("case %d: expected 0 books, got %d", i, len(books))
		}
	}
}
