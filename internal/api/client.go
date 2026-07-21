package api

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/NotMugil/hardcover-tui/internal/api/gen"
	"golang.org/x/time/rate"
)

const (
	graphqlEndpoint = "https://api.hardcover.app/v1/graphql"
	userAgent       = "github.com/NotMugil/hardcover-tui/1.1.1"
	requestsPerMin  = 60
	requestTimeout  = 30 * time.Second
)

// Client wraps the gqlgenc generated GraphQL client with rate limiting, auth, and query caching.
type Client struct {
	Gen     *gen.Client
	Cache   *QueryCache
	limiter *rate.Limiter
	token   string
	mu      sync.RWMutex
}

// authTransport injects auth headers and rate limiting into every request.
type authTransport struct {
	wrapped   http.RoundTripper
	tokenFunc func() string
	limiter   *rate.Limiter
}

func (t *authTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if err := t.limiter.Wait(req.Context()); err != nil {
		return nil, fmt.Errorf("rate limit: %w", err)
	}
	token := t.tokenFunc()
	req.Header.Set("Authorization", token)
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	return t.wrapped.RoundTrip(req)
}

// NewClient creates a new API client with the given auth token.
// The token should include the "Bearer " prefix.
func NewClient(token string) *Client {
	limiter := rate.NewLimiter(rate.Every(time.Minute/requestsPerMin), 1)
	c := &Client{
		token:   token,
		limiter: limiter,
		Cache:   NewQueryCache(),
	}

	httpClient := &http.Client{
		Timeout: requestTimeout,
		Transport: &authTransport{
			tokenFunc: func() string {
				c.mu.RLock()
				defer c.mu.RUnlock()
				return c.token
			},
			limiter: limiter,
			wrapped: http.DefaultTransport,
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

