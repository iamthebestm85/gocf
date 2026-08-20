package gocf

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/iamthebestm85/gocf/client"
	"github.com/iamthebestm85/gocf/fingerprint"
	"github.com/iamthebestm85/gocf/solver"
)

type Solver struct {
	config  solver.Config
	pool    *solver.Pool
	clients sync.Map
}

type Option func(*solver.Config)

func WithProxy(proxy string) Option {
	return func(c *solver.Config) {
		c.Proxy = proxy
	}
}

func WithHeadless(headless bool) Option {
	return func(c *solver.Config) {
		c.Headless = headless
	}
}

func WithTimeout(timeout time.Duration) Option {
	return func(c *solver.Config) {
		c.Timeout = timeout
	}
}

func WithUserAgent(ua string) Option {
	return func(c *solver.Config) {
		c.UserAgent = ua
	}
}

func WithViewport(width, height int) Option {
	return func(c *solver.Config) {
		c.ViewportWidth = width
		c.ViewportHeight = height
	}
}

func WithDisableImages(disable bool) Option {
	return func(c *solver.Config) {
		c.DisableImages = disable
	}
}

func New(opts ...Option) *Solver {
	cfg := solver.DefaultConfig()
	for _, opt := range opts {
		opt(&cfg)
	}
	return &Solver{
		config: cfg,
	}
}

// Solve navigates to the URL and solves any Cloudflare Turnstile or Managed Challenge present.
// It returns the result including token, cookies, and cf_clearance if obtained.
func (s *Solver) Solve(ctx context.Context, url string) (*solver.Result, error) {
	return solver.Solve(ctx, url, s.config)
}

// SolveWithWidget solves a Turnstile challenge by injecting a widget with the given sitekey.
func (s *Solver) SolveWithWidget(ctx context.Context, url, sitekey string) (*solver.Result, error) {
	return solver.SolveWithWidget(ctx, url, sitekey, s.config)
}

// SolveAll solves multiple URLs in parallel using a worker pool.
func (s *Solver) SolveAll(ctx context.Context, urls []string, workers int) []*solver.Result {
	pool := solver.NewPool(ctx, workers, s.config)
	return pool.SolveAll(urls)
}

// NewClient creates an HTTP client with browser-like fingerprinting.
func (s *Solver) NewClient(platform fingerprint.Platform) (*client.Client, error) {
	cfg := client.ClientConfig{
		Proxy:    s.config.Proxy,
		Platform: platform,
		Timeout:  s.config.Timeout,
	}
	return client.NewClient(cfg)
}

// NewClientNoRedirect creates an HTTP client that does not follow redirects,
// useful for capturing intermediate responses (e.g., 302 with cf_clearance).
func (s *Solver) NewClientNoRedirect(platform fingerprint.Platform) (*client.Client, error) {
	cfg := client.ClientConfig{
		Proxy:           s.config.Proxy,
		Platform:        platform,
		Timeout:         s.config.Timeout,
		FollowRedirects: false,
	}
	return client.NewClient(cfg)
}

// GetWithToken sends a GET request with the cookies obtained from solving.
func (s *Solver) GetWithToken(ctx context.Context, url, token, cookies string) (*http.Response, error) {
	c, err := s.NewClient(fingerprint.Windows)
	if err != nil {
		return nil, err
	}
	return c.DoWithCookies(url, cookies)
}

// PostWithToken sends a POST request.
func (s *Solver) PostWithToken(ctx context.Context, url, contentType string, body io.Reader, cookies string) (*http.Response, error) {
	c, err := s.NewClient(fingerprint.Windows)
	if err != nil {
		return nil, err
	}
	return c.Post(url, contentType, body)
}

// SolveAndRequest solves the challenge and immediately makes a follow-up request
// with the obtained cookies. This is the most common usage pattern for accessing
// Cloudflare-protected sites.
func SolveAndRequest(ctx context.Context, targetURL string, opts ...Option) (*solver.Result, *http.Response, error) {
	s := New(opts...)

	result, err := s.Solve(ctx, targetURL)
	if err != nil {
		return result, nil, fmt.Errorf("turnstile solve failed: %w", err)
	}

	// Use the challenge-specific client for the follow-up request
	c, err := s.NewClient(fingerprint.Windows)
	if err != nil {
		return result, nil, fmt.Errorf("client creation failed: %w", err)
	}

	// If the challenge was solved, use challenge-appropriate headers
	var resp *http.Response
	if result.ChallengeSolved && result.Cookies != "" {
		resp, err = c.DoWithChallengeCookies(targetURL, result.Cookies, targetURL)
	} else {
		resp, err = c.DoWithCookies(targetURL, result.Cookies)
	}
	if err != nil {
		return result, nil, fmt.Errorf("request failed: %w", err)
	}

	return result, resp, nil
}

// BatchSolve solves multiple URLs in parallel.
func BatchSolve(ctx context.Context, urls []string, workers int, opts ...Option) []*solver.Result {
	s := New(opts...)
	return s.SolveAll(ctx, urls, workers)
}
