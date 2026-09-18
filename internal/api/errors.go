package api

import (
	"errors"
	"fmt"
	"time"
)

var (
	// ErrTokenRevoked is returned when the API token is invalid, expired, or revoked (HTTP 401).
	ErrTokenRevoked = errors.New("API token is invalid, expired, or revoked")

	// ErrDailyLimitExceeded is returned when the user's daily API request quota is reached.
	ErrDailyLimitExceeded = errors.New("daily API limit reached; resets at midnight UTC")
)

// InsufficientScopeError indicates that the API token is missing a required scope (HTTP 403).
type InsufficientScopeError struct {
	Scope       string
	Description string
}

func (e *InsufficientScopeError) Error() string {
	if e.Description != "" {
		return fmt.Sprintf("insufficient permissions: %s (requires scope: %s)", e.Description, e.Scope)
	}
	return fmt.Sprintf("insufficient permissions: missing required scope '%s'", e.Scope)
}

// RateLimitExceededError indicates that request rate limiting was hit (HTTP 429).
type RateLimitExceededError struct {
	RetryAfter time.Duration
	Message    string
}

func (e *RateLimitExceededError) Error() string {
	if e.RetryAfter > 0 {
		return fmt.Sprintf("rate limit exceeded: retry after %v", e.RetryAfter)
	}
	if e.Message != "" {
		return fmt.Sprintf("rate limit exceeded: %s", e.Message)
	}
	return "rate limit exceeded; please wait a moment"
}
