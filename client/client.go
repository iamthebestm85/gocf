package client

import (
	"compress/gzip"
	"compress/zlib"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/iamthebestm85/gocf/fingerprint"
)

type Client struct {
	httpClient *http.Client
	profile    fingerprint.BrowserProfile
	h2Settings fingerprint.H2Settings
	cookieJar  sync.Map
}

type ClientConfig struct {
	Proxy        string
	Platform     fingerprint.Platform
	Timeout      time.Duration
	MaxIdleConns int
	// FollowRedirects controls whether the client follows HTTP redirects.
	// When false, the client returns the redirect response as-is (useful for
	// capturing cf_clearance cookies from intermediate responses).
	FollowRedirects bool
}

func DefaultClientConfig() ClientConfig {
	return ClientConfig{
		Platform:         fingerprint.Windows,
		Timeout:          30 * time.Second,
		MaxIdleConns:     100,
		FollowRedirects:  true,
	}
}

func NewClient(cfg ClientConfig) (*Client, error) {
	profile := fingerprint.GetProfile(cfg.Platform)
	h2Settings := fingerprint.ChromeH2Settings()

	tlsConfig := &tls.Config{
		MinVersion: tls.VersionTLS12,
		MaxVersion: tls.VersionTLS13,
		CipherSuites: []uint16{
			tls.TLS_AES_128_GCM_SHA256,
			tls.TLS_AES_256_GCM_SHA384,
			tls.TLS_CHACHA20_POLY1305_SHA256,
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
			tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
		},
		CurvePreferences: []tls.CurveID{
			tls.X25519,
			tls.CurveP256,
			tls.CurveP384,
		},
	}

	transport := &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   cfg.Timeout,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		TLSClientConfig:     tlsConfig,
		MaxIdleConns:        cfg.MaxIdleConns,
		MaxIdleConnsPerHost: cfg.MaxIdleConns,
		IdleConnTimeout:     90 * time.Second,
		TLSHandshakeTimeout: cfg.Timeout,
		DisableKeepAlives:   false,
		DisableCompression:  true,
		ForceAttemptHTTP2:   true,
	}

	if cfg.Proxy != "" {
		proxyURL, err := url.Parse(cfg.Proxy)
		if err != nil {
			return nil, fmt.Errorf("invalid proxy URL: %w", err)
		}
		transport.Proxy = http.ProxyURL(proxyURL)
	}

	httpClient := &http.Client{
		Transport: &h2Transport{
			inner:      transport,
			h2Settings: h2Settings,
			profile:    profile,
		},
		Timeout: cfg.Timeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if !cfg.FollowRedirects {
				return http.ErrUseLastResponse
			}
			if len(via) >= 10 {
				return fmt.Errorf("too many redirects")
			}
			return nil
		},
	}

	return &Client{
		httpClient: httpClient,
		profile:    profile,
		h2Settings: h2Settings,
	}, nil
}

type h2Transport struct {
	inner      *http.Transport
	h2Settings fingerprint.H2Settings
	profile    fingerprint.BrowserProfile
}

func (t *h2Transport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("User-Agent", t.profile.UserAgent)

	orderedHeaders := fingerprint.HeaderOrder
	newHeader := http.Header{}
	for _, key := range orderedHeaders {
		if val := req.Header.Get(key); val != "" {
			newHeader.Set(key, val)
		}
	}
	for key, values := range req.Header {
		lowerKey := strings.ToLower(key)
		found := false
		for _, ordered := range orderedHeaders {
			if strings.ToLower(ordered) == lowerKey {
				found = true
				break
			}
		}
		if !found {
			for _, v := range values {
				newHeader.Add(key, v)
			}
		}
	}
	req.Header = newHeader

	resp, err := t.inner.RoundTrip(req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (c *Client) Get(url string) (*http.Response, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	return c.Do(req)
}

func (c *Client) Post(url, contentType string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", contentType)
	return c.Do(req)
}

func (c *Client) Do(req *http.Request) (*http.Response, error) {
	for key, val := range fingerprint.BuildHeaders(c.profile) {
		if req.Header.Get(key) == "" {
			req.Header.Set(key, val[0])
		}
	}

	origin := req.URL.Scheme + "://" + req.URL.Host
	req.Header.Set("Referer", origin)

	return c.httpClient.Do(req)
}

// DoWithChallengeCookies sends a request with cookies obtained from a solved challenge.
// It sets up headers appropriate for a post-challenge request (same-origin sec-fetch-* headers).
func (c *Client) DoWithChallengeCookies(urlStr string, cookies string, referer string) (*http.Response, error) {
	req, err := http.NewRequest("GET", urlStr, nil)
	if err != nil {
		return nil, err
	}

	if cookies != "" {
		req.Header.Set("Cookie", cookies)
	}

	// Use challenge-specific headers for post-solve requests
	if referer == "" {
		referer = urlStr
	}
	challengeHeaders := fingerprint.BuildChallengeHeaders(c.profile, referer)
	for key, val := range challengeHeaders {
		if req.Header.Get(key) == "" {
			req.Header.Set(key, val[0])
		}
	}

	return c.httpClient.Do(req)
}

func (c *Client) DoWithCookies(urlStr string, cookies string) (*http.Response, error) {
	req, err := http.NewRequest("GET", urlStr, nil)
	if err != nil {
		return nil, err
	}

	if cookies != "" {
		req.Header.Set("Cookie", cookies)
	}

	return c.Do(req)
}

// ExtractCFClearance extracts the cf_clearance cookie value from a cookie string.
func ExtractCFClearance(cookies string) string {
	parts := strings.Split(cookies, ";")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if strings.HasPrefix(part, "cf_clearance=") {
			return strings.TrimPrefix(part, "cf_clearance=")
		}
	}
	return ""
}

// ExtractCookie extracts a specific cookie value by name from a cookie string.
func ExtractCookie(cookies, name string) string {
	parts := strings.Split(cookies, ";")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if strings.HasPrefix(part, name+"=") {
			return strings.TrimPrefix(part, name+"=")
		}
	}
	return ""
}

func (c *Client) SetCookies(domain, cookies string) {
	c.cookieJar.Store(domain, cookies)
}

func (c *Client) GetCookies(domain string) string {
	if val, ok := c.cookieJar.Load(domain); ok {
		return val.(string)
	}
	return ""
}

func DecompressBody(resp *http.Response) ([]byte, error) {
	var reader io.Reader = resp.Body

	switch resp.Header.Get("Content-Encoding") {
	case "gzip":
		gr, err := gzip.NewReader(resp.Body)
		if err != nil {
			return nil, err
		}
		defer gr.Close()
		reader = gr
	case "deflate":
		zr, err := zlib.NewReader(resp.Body)
		if err != nil {
			return nil, err
		}
		defer zr.Close()
		reader = zr
	case "br":
		return io.ReadAll(resp.Body)
	}

	return io.ReadAll(reader)
}
