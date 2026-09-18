package api

import (
	"strconv"
	"strings"
	"sync"
	"time"
)

// RateLimitBucket holds the parsed status of a rate limit bucket.
type RateLimitBucket struct {
	Name      string
	Quota     int           // q
	Window    time.Duration // w
	Burst     int           // burst
	Remaining int           // r
	ResetIn   time.Duration // t
}

// RateLimitState tracks current rate limiting policy and status from API response headers.
type RateLimitState struct {
	mu           sync.RWMutex
	Plan         string
	BurstLimit   int
	PerMinQuota  int
	PerMinRemain int
	PerMinReset  time.Duration
	DailyLimit   int
	DailyRemain  int
	DailyReset   time.Duration
	LastUpdated  time.Time
}

// NewRateLimitState creates a default RateLimitState with standard Free tier defaults.
func NewRateLimitState() *RateLimitState {
	return &RateLimitState{
		Plan:         "Free",
		BurstLimit:   10,
		PerMinQuota:  60,
		PerMinRemain: 10,
		PerMinReset:  60 * time.Second,
		DailyLimit:   5000,
		DailyRemain:  5000,
		DailyReset:   24 * time.Hour,
		LastUpdated:  time.Now(),
	}
}

// UpdateFromHeaders parses IETF RateLimit-Policy and RateLimit headers.
func (s *RateLimitState) UpdateFromHeaders(policyHeader, statusHeader string) {
	if s == nil || (policyHeader == "" && statusHeader == "") {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Parse RateLimit-Policy: "Free";q=60;w=60;burst=10, "daily";q=5000;w=86400
	if policyHeader != "" {
		buckets := parseHeaderBuckets(policyHeader)
		for _, b := range buckets {
			name := strings.ToLower(b.Name)
			if name == "daily" {
				if b.Quota > 0 {
					s.DailyLimit = b.Quota
				}
			} else {
				// Plan bucket (e.g. "Free" or "Supporter")
				s.Plan = b.Name
				if b.Quota > 0 {
					s.PerMinQuota = b.Quota
				}
				if b.Burst > 0 {
					s.BurstLimit = b.Burst
				}
			}
		}
	}

	// Parse RateLimit: "Free";r=8;t=42, "daily";r=4231;t=51234
	if statusHeader != "" {
		buckets := parseHeaderBuckets(statusHeader)
		for _, b := range buckets {
			name := strings.ToLower(b.Name)
			if name == "daily" {
				s.DailyRemain = b.Remaining
				s.DailyReset = b.ResetIn
			} else {
				s.PerMinRemain = b.Remaining
				s.PerMinReset = b.ResetIn
			}
		}
	}

	s.LastUpdated = time.Now()
}

// GetBurstLimit returns the current burst limit.
func (s *RateLimitState) GetBurstLimit() int {
	if s == nil {
		return 10
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.BurstLimit > 0 {
		return s.BurstLimit
	}
	return 10
}

// GetPerMinQuota returns requests per minute quota.
func (s *RateLimitState) GetPerMinQuota() int {
	if s == nil {
		return 60
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.PerMinQuota > 0 {
		return s.PerMinQuota
	}
	return 60
}

func parseHeaderBuckets(header string) []RateLimitBucket {
	var buckets []RateLimitBucket
	parts := strings.Split(header, ",")

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		subParts := strings.Split(part, ";")
		if len(subParts) == 0 {
			continue
		}

		name := strings.Trim(strings.TrimSpace(subParts[0]), `"'`)
		bucket := RateLimitBucket{Name: name}

		for _, param := range subParts[1:] {
			param = strings.TrimSpace(param)
			kv := strings.SplitN(param, "=", 2)
			if len(kv) != 2 {
				continue
			}
			k := strings.TrimSpace(kv[0])
			v := strings.TrimSpace(kv[1])

			switch k {
			case "q":
				if n, err := strconv.Atoi(v); err == nil {
					bucket.Quota = n
				}
			case "w":
				if n, err := strconv.Atoi(v); err == nil {
					bucket.Window = time.Duration(n) * time.Second
				}
			case "burst":
				if n, err := strconv.Atoi(v); err == nil {
					bucket.Burst = n
				}
			case "r":
				if n, err := strconv.Atoi(v); err == nil {
					bucket.Remaining = n
				}
			case "t":
				if n, err := strconv.Atoi(v); err == nil {
					bucket.ResetIn = time.Duration(n) * time.Second
				}
			}
		}

		buckets = append(buckets, bucket)
	}

	return buckets
}

// ParseRetryAfter parses the Retry-After header as seconds or returns fallback.
func ParseRetryAfter(header string, fallback time.Duration) time.Duration {
	header = strings.TrimSpace(header)
	if header == "" {
		return fallback
	}

	if secs, err := strconv.Atoi(header); err == nil && secs > 0 {
		return time.Duration(secs) * time.Second
	}

	if t, err := time.Parse(time.RFC1123, header); err == nil {
		diff := time.Until(t)
		if diff > 0 {
			return diff
		}
	}

	return fallback
}
