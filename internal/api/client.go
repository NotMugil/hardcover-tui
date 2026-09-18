package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/NotMugil/hardcover-tui/internal/api/gen"
	"golang.org/x/time/rate"
)

const (
	graphqlEndpoint = "https://api.hardcover.app/v1/graphql"
	userAgent       = "github.com/NotMugil/hardcover-tui/1.2.0"
	requestsPerMin  = 60
	defaultBurst    = 10
	requestTimeout  = 30 * time.Second
)

// Client wraps the gqlgenc generated GraphQL client with rate limiting, auth, and query caching.
type Client struct {
	Gen       *gen.Client
	Cache     *QueryCache
	RateState *RateLimitState
	limiter   *rate.Limiter
	token     string
	mu        sync.RWMutex
}

// authTransport injects auth headers, rate limiting, retry backoff, and header tracking into every request.
type authTransport struct {
	wrapped   http.RoundTripper
	tokenFunc func() string
	limiter   *rate.Limiter
	rateState *RateLimitState
}

func (t *authTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	var getBody func() (io.ReadCloser, error)
	if req.GetBody != nil {
		getBody = req.GetBody
	} else if req.Body != nil {
		bodyBytes, err := io.ReadAll(req.Body)
		if err != nil {
			return nil, err
		}
		_ = req.Body.Close()
		req.Body = io.NopCloser(bytes.NewReader(bodyBytes))
		getBody = func() (io.ReadCloser, error) {
			return io.NopCloser(bytes.NewReader(bodyBytes)), nil
		}
	}

	maxRetries := 2
	for attempt := 0; attempt <= maxRetries; attempt++ {
		if err := t.limiter.Wait(req.Context()); err != nil {
			return nil, fmt.Errorf("rate limit: %w", err)
		}

		if attempt > 0 && getBody != nil {
			body, err := getBody()
			if err != nil {
				return nil, err
			}
			req.Body = body
		}

		token := t.tokenFunc()
		req.Header.Set("Authorization", token)
		req.Header.Set("User-Agent", userAgent)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json")

		resp, err := t.wrapped.RoundTrip(req)
		if err != nil {
			return nil, err
		}

		// Parse IETF rate limit headers if present
		policyHeader := resp.Header.Get("RateLimit-Policy")
		statusHeader := resp.Header.Get("RateLimit")
		if policyHeader != "" || statusHeader != "" {
			t.rateState.UpdateFromHeaders(policyHeader, statusHeader)
			// Dynamically adjust local limiter settings based on server policy
			newBurst := t.rateState.GetBurstLimit()
			newQuota := t.rateState.GetPerMinQuota()
			if t.limiter.Burst() != newBurst {
				t.limiter.SetBurst(newBurst)
			}
			newLimit := rate.Every(time.Minute / time.Duration(newQuota))
			if t.limiter.Limit() != newLimit {
				t.limiter.SetLimit(newLimit)
			}
		}

		// Handle HTTP 401 Unauthorized (invalid, expired, or revoked token)
		if resp.StatusCode == http.StatusUnauthorized {
			_ = resp.Body.Close()
			return nil, ErrTokenRevoked
		}

		// Handle HTTP 429 Too Many Requests with Retry-After backoff
		if resp.StatusCode == http.StatusTooManyRequests {
			retryAfter := ParseRetryAfter(resp.Header.Get("Retry-After"), 2*time.Second)
			_ = resp.Body.Close()

			if t.rateState != nil && t.rateState.DailyRemain == 0 && t.rateState.DailyLimit > 0 {
				return nil, ErrDailyLimitExceeded
			}

			if attempt < maxRetries {
				select {
				case <-req.Context().Done():
					return nil, req.Context().Err()
				case <-time.After(retryAfter):
					continue
				}
			}

			return nil, &RateLimitExceededError{
				RetryAfter: retryAfter,
				Message:    "Too Many Requests: rate limit ceiling reached, please wait before retrying",
			}
		}

		// Handle HTTP 403 with detailed error descriptions
		if resp.StatusCode == http.StatusForbidden {
			bodyBytes, _ := io.ReadAll(resp.Body)
			_ = resp.Body.Close()
			resp.Body = io.NopCloser(bytes.NewReader(bodyBytes))

			var errResp struct {
				Error            string   `json:"error"`
				ErrorDescription string   `json:"error_description"`
				Scope            string   `json:"scope"`
				Errors           []string `json:"errors"`
			}
			if json.Unmarshal(bodyBytes, &errResp) == nil {
				if errResp.Error == "insufficient_scope" {
					return nil, &InsufficientScopeError{
						Scope:       errResp.Scope,
						Description: errResp.ErrorDescription,
					}
				}
				if errResp.Error == "top_level_limit_exceeded" || (len(errResp.Errors) > 0 && errResp.Errors[0] == "top_level_limit_exceeded") {
					return nil, errors.New("API query error (403): Exceeded 5 top-level operations in single request")
				}
				if errResp.Error == "unsupported_operation" {
					return nil, errors.New("API operation error (403): Operation unsupported for API tokens")
				}
			}
		}

		return resp, nil
	}

	return nil, &RateLimitExceededError{
		Message: "rate limit retries exhausted",
	}
}

// NewClient creates a new API client with the given auth token.
// The token should include the "Bearer " prefix.
func NewClient(token string) *Client {
	rateState := NewRateLimitState()
	limiter := rate.NewLimiter(rate.Every(time.Minute/requestsPerMin), defaultBurst)
	c := &Client{
		token:     token,
		limiter:   limiter,
		RateState: rateState,
		Cache:     NewQueryCache(),
	}

	httpClient := &http.Client{
		Timeout: requestTimeout,
		Transport: &authTransport{
			tokenFunc: func() string {
				c.mu.RLock()
				defer c.mu.RUnlock()
				return c.token
			},
			limiter:   limiter,
			rateState: rateState,
			wrapped:   http.DefaultTransport,
		},
	}

	c.Gen = gen.NewClient(httpClient, graphqlEndpoint, nil)
	return c
}

// SetToken updates the auth token.
func (c *Client) SetToken(token string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.token = token
}
