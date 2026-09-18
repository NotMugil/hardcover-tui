package auth

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

const (
	DefaultClientID       = "06b50590-3815-4378-97bd-c7e40d86707e"
	deviceEndpoint        = "https://hardcover.app/oauth2/device"
	tokenEndpoint         = "https://hardcover.app/oauth2/token"
	fallbackTokenEndpoint = "https://hardcover.app/oauth/token"
	defaultUserAgent      = "HardcoverTUI/1.2.0 (github.com/NotMugil/hardcover-tui)"
)

// DeviceAuthResponse is the payload returned by the device authorization endpoint.
type DeviceAuthResponse struct {
	DeviceCode              string `json:"device_code"`
	UserCode                string `json:"user_code"`
	VerificationURI         string `json:"verification_uri"`
	VerificationURIComplete string `json:"verification_uri_complete"`
	ExpiresIn               int    `json:"expires_in"`
	Interval                int    `json:"interval"`
	Error                   string `json:"error,omitempty"`
	ErrorDescription        string `json:"error_description,omitempty"`
}

// TokenResponse is the payload returned by the token endpoint.
type TokenResponse struct {
	AccessToken      string `json:"access_token"`
	TokenType        string `json:"token_type"`
	ExpiresIn        int    `json:"expires_in"`
	RefreshToken     string `json:"refresh_token"`
	Scope            string `json:"scope"`
	Error            string `json:"error,omitempty"`
	ErrorDescription string `json:"error_description,omitempty"`
}

// GetClientID returns the configured Hardcover client ID.
func GetClientID() string {
	if envID := os.Getenv("HARDCOVER_CLIENT_ID"); envID != "" {
		return strings.TrimSpace(envID)
	}
	if envID := os.Getenv("CLIENT_ID"); envID != "" {
		return strings.TrimSpace(envID)
	}

	// Attempt reading from .env file if present in current directory
	if id := readClientIDFromEnvFile(".env"); id != "" {
		return id
	}

	return DefaultClientID
}

func readClientIDFromEnvFile(filename string) string {
	file, err := os.Open(filename)
	if err != nil {
		return ""
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "#") || !strings.Contains(line, "=") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])
		val = strings.Trim(val, `"'`)
		if key == "CLIENT_ID" || key == "HARDCOVER_CLIENT_ID" {
			return val
		}
	}
	return ""
}

// RequestDeviceCode initiates the Device Authorization flow.
func RequestDeviceCode(ctx context.Context, clientID string) (*DeviceAuthResponse, error) {
	data := url.Values{}
	data.Set("client_id", clientID)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, deviceEndpoint, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("create device auth request: %w", err)
	}

	req.Header.Set("User-Agent", defaultUserAgent)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request device authorization: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read device authorization response: %w", err)
	}

	var authResp DeviceAuthResponse
	if err := json.Unmarshal(body, &authResp); err != nil {
		return nil, fmt.Errorf("parse device authorization response (status %d): %w", resp.StatusCode, err)
	}

	if authResp.Error != "" {
		desc := authResp.ErrorDescription
		if desc == "" {
			desc = authResp.Error
		}
		return nil, fmt.Errorf("device authorization error: %s", desc)
	}

	if authResp.DeviceCode == "" || authResp.UserCode == "" {
		return nil, errors.New("invalid response from authorization server: missing device or user code")
	}

	if authResp.Interval <= 0 {
		authResp.Interval = 5
	}
	if authResp.ExpiresIn <= 0 {
		authResp.ExpiresIn = 900 // 15 minutes default
	}
	if authResp.VerificationURI == "" {
		authResp.VerificationURI = "https://hardcover.app/link"
	}
	if authResp.VerificationURIComplete == "" {
		cleanCode := strings.ReplaceAll(authResp.UserCode, "-", "")
		authResp.VerificationURIComplete = fmt.Sprintf("https://hardcover.app/link?c=%s", url.QueryEscape(cleanCode))
	}

	return &authResp, nil
}

// PollToken polls the token endpoint until the user authorizes or the request expires.
func PollToken(ctx context.Context, clientID, deviceCode string, interval time.Duration, expiresAt time.Time, onTick func()) (string, error) {
	if interval < 1*time.Second {
		interval = 5 * time.Second
	}

	client := &http.Client{Timeout: 15 * time.Second}

	for {
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		default:
		}

		if time.Now().After(expiresAt) {
			return "", errors.New("device authorization request has expired. Please run 'hardcover-tui auth login' again")
		}

		if onTick != nil {
			onTick()
		}

		token, err, shouldRetry := requestTokenOnce(ctx, client, clientID, deviceCode)
		if err == nil && token != "" {
			return token, nil
		}

		if !shouldRetry {
			return "", err
		}

		// Check if error requested slow down
		if errors.Is(err, ErrSlowDown) {
			interval += 5 * time.Second
		}

		select {
		case <-ctx.Done():
			return "", ctx.Err()
		case <-time.After(interval):
		}
	}
}

var (
	ErrSlowDown = errors.New("slow_down")
	ErrPending  = errors.New("authorization_pending")
	ErrDenied   = errors.New("access_denied")
	ErrExpired  = errors.New("expired_token")
)

func requestTokenOnce(ctx context.Context, client *http.Client, clientID, deviceCode string) (token string, err error, shouldRetry bool) {
	data := url.Values{}
	data.Set("client_id", clientID)
	data.Set("grant_type", "urn:ietf:params:oauth:grant-type:device_code")
	data.Set("device_code", deviceCode)

	endpoints := []string{tokenEndpoint, fallbackTokenEndpoint}
	var lastErr error

	for _, endpoint := range endpoints {
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(data.Encode()))
		if err != nil {
			return "", err, false
		}

		req.Header.Set("User-Agent", defaultUserAgent)
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.Header.Set("Accept", "application/json")

		resp, err := client.Do(req)
		if err != nil {
			lastErr = err
			continue
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			lastErr = err
			continue
		}

		var tokenResp TokenResponse
		if err := json.Unmarshal(body, &tokenResp); err != nil {
			lastErr = fmt.Errorf("parse token response: %w", err)
			continue
		}

		if resp.StatusCode == http.StatusOK && tokenResp.AccessToken != "" {
			return tokenResp.AccessToken, nil, false
		}

		switch tokenResp.Error {
		case "authorization_pending":
			return "", ErrPending, true
		case "slow_down":
			return "", ErrSlowDown, true
		case "access_denied":
			return "", fmt.Errorf("authorization denied by user: %s", tokenResp.ErrorDescription), false
		case "expired_token":
			return "", errors.New("authorization code expired"), false
		default:
			if resp.StatusCode == http.StatusNotFound {
				// Try fallback endpoint
				continue
			}
			desc := tokenResp.ErrorDescription
			if desc == "" {
				desc = tokenResp.Error
			}
			if desc == "" {
				desc = fmt.Sprintf("HTTP %d", resp.StatusCode)
			}
			return "", fmt.Errorf("token request failed: %s", desc), false
		}
	}

	if lastErr != nil {
		return "", lastErr, true
	}
	return "", errors.New("token request failed"), false
}

// OpenBrowser opens the given URL in the user's default browser.
func OpenBrowser(targetURL string) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", targetURL)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", targetURL)
	default: // linux, bsd
		cmd = exec.Command("xdg-open", targetURL)
	}

	return cmd.Start()
}
