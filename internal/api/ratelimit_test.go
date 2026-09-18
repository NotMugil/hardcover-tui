package api

import (
	"testing"
	"time"
)

func TestParseHeaderBuckets(t *testing.T) {
	policy := `"Free";q=60;w=60;burst=10, "daily";q=5000;w=86400`
	status := `"Free";r=8;t=42, "daily";r=4231;t=51234`

	state := NewRateLimitState()
	state.UpdateFromHeaders(policy, status)

	if state.Plan != "Free" {
		t.Errorf("expected Plan Free, got %s", state.Plan)
	}
	if state.BurstLimit != 10 {
		t.Errorf("expected BurstLimit 10, got %d", state.BurstLimit)
	}
	if state.PerMinQuota != 60 {
		t.Errorf("expected PerMinQuota 60, got %d", state.PerMinQuota)
	}
	if state.DailyLimit != 5000 {
		t.Errorf("expected DailyLimit 5000, got %d", state.DailyLimit)
	}
	if state.PerMinRemain != 8 {
		t.Errorf("expected PerMinRemain 8, got %d", state.PerMinRemain)
	}
	if state.PerMinReset != 42*time.Second {
		t.Errorf("expected PerMinReset 42s, got %v", state.PerMinReset)
	}
	if state.DailyRemain != 4231 {
		t.Errorf("expected DailyRemain 4231, got %d", state.DailyRemain)
	}
	if state.DailyReset != 51234*time.Second {
		t.Errorf("expected DailyReset 51234s, got %v", state.DailyReset)
	}
}

func TestSupporterPlanUpdate(t *testing.T) {
	policy := `"Supporter";q=60;w=60;burst=15, "daily";q=50000;w=86400`
	status := `"Supporter";r=14;t=12, "daily";r=49900;t=1234`

	state := NewRateLimitState()
	state.UpdateFromHeaders(policy, status)

	if state.Plan != "Supporter" {
		t.Errorf("expected Plan Supporter, got %s", state.Plan)
	}
	if state.BurstLimit != 15 {
		t.Errorf("expected BurstLimit 15, got %d", state.BurstLimit)
	}
	if state.DailyLimit != 50000 {
		t.Errorf("expected DailyLimit 50000, got %d", state.DailyLimit)
	}
}

func TestParseRetryAfter(t *testing.T) {
	d := ParseRetryAfter("5", 2*time.Second)
	if d != 5*time.Second {
		t.Errorf("expected 5s, got %v", d)
	}

	d2 := ParseRetryAfter("", 2*time.Second)
	if d2 != 2*time.Second {
		t.Errorf("expected fallback 2s, got %v", d2)
	}
}
