package gocf

import (
	"context"
	"testing"
	"time"

	"github.com/iamthebestm85/gocf/fingerprint"
)

func TestNew(t *testing.T) {
	s := New()
	if s == nil {
		t.Fatal("New() returned nil")
	}
}

func TestNewWithProxy(t *testing.T) {
	s := New(WithProxy("http://127.0.0.1:8080"))
	if s == nil {
		t.Fatal("New() returned nil")
	}
}

func TestNewWithTimeout(t *testing.T) {
	s := New(WithTimeout(10 * time.Second))
	if s == nil {
		t.Fatal("New() returned nil")
	}
}

func TestNewClient(t *testing.T) {
	s := New()
	c, err := s.NewClient(fingerprint.Windows)
	if err != nil {
		t.Fatalf("NewClient() error: %v", err)
	}
	if c == nil {
		t.Fatal("NewClient() returned nil")
	}
}

func TestNewClientNoRedirect(t *testing.T) {
	s := New()
	c, err := s.NewClientNoRedirect(fingerprint.Windows)
	if err != nil {
		t.Fatalf("NewClientNoRedirect() error: %v", err)
	}
	if c == nil {
		t.Fatal("NewClientNoRedirect() returned nil")
	}
}

func TestNewClientWithProxy(t *testing.T) {
	s := New(WithProxy("http://127.0.0.1:8080"))
	_, err := s.NewClient(fingerprint.Windows)
	if err == nil {
		t.Log("Client created (proxy may not be reachable)")
	}
}

func TestFingerprint(t *testing.T) {
	profile := fingerprint.GetProfile(fingerprint.Windows)
	if profile.UserAgent == "" {
		t.Fatal("UserAgent is empty")
	}
	if profile.SecCHUA == "" {
		t.Fatal("SecCHUA is empty")
	}
	// Verify Chrome 138
	if !contains(profile.UserAgent, "Chrome/138") {
		t.Fatalf("Expected Chrome/138 in UserAgent, got: %s", profile.UserAgent)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestH2Settings(t *testing.T) {
	settings := fingerprint.ChromeH2Settings()
	if settings.HeaderTableSize != 65536 {
		t.Fatalf("HeaderTableSize = %d, want 65536", settings.HeaderTableSize)
	}
	if settings.EnablePush != 0 {
		t.Fatalf("EnablePush = %d, want 0", settings.EnablePush)
	}
	if settings.InitialWindowSize != 6291456 {
		t.Fatalf("InitialWindowSize = %d, want 6291456", settings.InitialWindowSize)
	}
	if settings.MaxHeaderListSize != 262144 {
		t.Fatalf("MaxHeaderListSize = %d, want 262144", settings.MaxHeaderListSize)
	}
	if settings.WindowUpdateSize != 15663105 {
		t.Fatalf("WindowUpdateSize = %d, want 15663105", settings.WindowUpdateSize)
	}
	if settings.MaxFrameSize != 16384 {
		t.Fatalf("MaxFrameSize = %d, want 16384", settings.MaxFrameSize)
	}
	if settings.PriorityWeight != 256 {
		t.Fatalf("PriorityWeight = %d, want 256", settings.PriorityWeight)
	}
}

func TestHeaders(t *testing.T) {
	profile := fingerprint.GetProfile(fingerprint.Windows)
	headers := fingerprint.BuildHeaders(profile)

	if headers.Get("user-agent") == "" {
		t.Fatal("user-agent header is empty")
	}
	if headers.Get("sec-ch-ua") == "" {
		t.Fatal("sec-ch-ua header is empty")
	}
	if headers.Get("accept") == "" {
		t.Fatal("accept header is empty")
	}
	if headers.Get("priority") == "" {
		t.Fatal("priority header is empty")
	}
}

func TestChallengeHeaders(t *testing.T) {
	profile := fingerprint.GetProfile(fingerprint.Windows)
	headers := fingerprint.BuildChallengeHeaders(profile, "https://example.com")

	if headers.Get("sec-fetch-site") != "same-origin" {
		t.Fatalf("sec-fetch-site = %s, want same-origin", headers.Get("sec-fetch-site"))
	}
	if headers.Get("sec-fetch-mode") != "navigate" {
		t.Fatalf("sec-fetch-mode = %s, want navigate", headers.Get("sec-fetch-mode"))
	}
	if headers.Get("referer") != "https://example.com" {
		t.Fatalf("referer = %s, want https://example.com", headers.Get("referer"))
	}
}

func TestAkamaiFingerprint(t *testing.T) {
	settings := fingerprint.ChromeH2Settings()
	fp := fingerprint.FormatAkamaiFingerprint(settings)
	if fp == "" {
		t.Fatal("Akamai fingerprint is empty")
	}
	t.Logf("Akamai fingerprint: %s", fp)
}

func TestSolveTimeout(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	s := New(WithTimeout(5 * time.Second))
	_, err := s.Solve(ctx, "https://example.com")
	if err != nil {
		t.Logf("Solve error (expected for non-turnstile page): %v", err)
	}
}

func TestBatchSolveTimeout(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	urls := []string{
		"https://example.com",
		"https://example.org",
	}

	results := BatchSolve(ctx, urls, 2, WithTimeout(10*time.Second))
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
}
