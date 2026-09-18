package api

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"golang.org/x/time/rate"
)

func TestInsufficientScopeError(t *testing.T) {
	err := &InsufficientScopeError{
		Scope:       "write:library",
		Description: "missing write permission on user books",
	}

	expected := "insufficient permissions: missing write permission on user books (requires scope: write:library)"
	if err.Error() != expected {
		t.Errorf("expected '%s', got '%s'", expected, err.Error())
	}
}

func TestRateLimitExceededError(t *testing.T) {
	err := &RateLimitExceededError{
		RetryAfter: 5 * time.Second,
	}

	expected := "rate limit exceeded: retry after 5s"
	if err.Error() != expected {
		t.Errorf("expected '%s', got '%s'", expected, err.Error())
	}
}

func TestRoundTripErrors(t *testing.T) {
	// Test 401 Unauthorized
	server401 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error":"invalid_token","error_description":"token is expired"}`))
	}))
	defer server401.Close()

	rateState := NewRateLimitState()
	limiter := rate.NewLimiter(rate.Every(time.Minute/requestsPerMin), defaultBurst)

	httpClient := &http.Client{
		Transport: &authTransport{
			tokenFunc: func() string { return "Bearer test-token" },
			limiter:   limiter,
			rateState: rateState,
			wrapped:   http.DefaultTransport,
		},
	}

	req, _ := http.NewRequestWithContext(context.Background(), http.MethodPost, server401.URL, nil)
	_, err := httpClient.Do(req)
	if !errors.Is(err, ErrTokenRevoked) {
		t.Errorf("expected ErrTokenRevoked, got %v", err)
	}

	// Test 403 Insufficient Scope
	server403 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"error":"insufficient_scope","error_description":"token lacks scope","scope":"write:library"}`))
	}))
	defer server403.Close()

	req403, _ := http.NewRequestWithContext(context.Background(), http.MethodPost, server403.URL, nil)
	_, err403 := httpClient.Do(req403)
	var scopeErr *InsufficientScopeError
	if !errors.As(err403, &scopeErr) {
		t.Errorf("expected InsufficientScopeError, got %v", err403)
	} else if scopeErr.Scope != "write:library" {
		t.Errorf("expected scope write:library, got %s", scopeErr.Scope)
	}
}
