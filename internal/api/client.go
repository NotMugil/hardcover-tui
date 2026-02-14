package api

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	graphql "github.com/hasura/go-graphql-client"
	"golang.org/x/time/rate"
)

const (
	graphqlEndpoint = "https://api.hardcover.app/v1/graphql"
	userAgent       = "hardcover-tui/1.0"
	requestsPerMin  = 60
	requestTimeout  = 30 * time.Second
)

// Client wraps the GraphQL client with rate limiting and auth.
type Client struct {
	gql     *graphql.Client
	limiter *rate.Limiter
	token   string
	mu      sync.RWMutex
}

// authTransport injects auth headers into every request.
type authTransport struct {
	wrapped   http.RoundTripper
	tokenFunc func() string
}

func (t *authTransport) RoundTrip(req *http.Request) (*http.Response, error) {
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
	c := &Client{
		token:   token,
		limiter: rate.NewLimiter(rate.Every(time.Minute/requestsPerMin), 1),
	}

	httpClient := &http.Client{
		Timeout: requestTimeout,
		Transport: &authTransport{
			tokenFunc: func() string {
				c.mu.RLock()
				defer c.mu.RUnlock()
				return c.token
			},
			wrapped: http.DefaultTransport,
		},
	}

	c.gql = graphql.NewClient(graphqlEndpoint, httpClient)
	return c
}

// SetToken updates the auth token.
func (c *Client) SetToken(token string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.token = token
}

// Query executes a GraphQL query with rate limiting.
func (c *Client) Query(ctx context.Context, q interface{}, variables map[string]interface{}) error {
	if err := c.limiter.Wait(ctx); err != nil {
		return fmt.Errorf("rate limit: %w", err)
	}
	return c.gql.Query(ctx, q, variables)
}

// Mutate executes a GraphQL mutation with rate limiting.
func (c *Client) Mutate(ctx context.Context, m interface{}, variables map[string]interface{}) error {
	if err := c.limiter.Wait(ctx); err != nil {
		return fmt.Errorf("rate limit: %w", err)
	}
	return c.gql.Mutate(ctx, m, variables)
}

// ExecRaw executes a raw GraphQL query string with rate limiting.
func (c *Client) ExecRaw(ctx context.Context, query string, variables map[string]any) ([]byte, error) {
	if err := c.limiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("rate limit: %w", err)
	}
	return c.gql.ExecRaw(ctx, query, variables)
}
