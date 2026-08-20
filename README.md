# GOCF - Go Cloudflare Turnstile & Managed Challenge Solver

A Go library for solving Cloudflare Turnstile challenges and Managed Challenges with real browser fingerprints.

## Features

- **Real Browser Fingerprinting** - Chrome 138 HTTP/2 settings, headers, TLS
- **Turnstile Solver** - Automatic Cloudflare Turnstile token extraction
- **Managed Challenge Support** - Handles Cloudflare's full-page managed challenges
- **Advanced Stealth** - Comprehensive anti-detection: webdriver masking, plugin spoofing, WebGL override, screen properties
- **Proxy Support** - HTTP/SOCKS5 proxy support
- **Parallel Execution** - Worker pool for batch solving
- **Chrome-like Headers** - Proper header ordering matching real Chrome
- **Rod-based** - Uses [go-rod/rod](https://github.com/go-rod/rod) for reliable browser automation

## Installation

```bash
go get github.com/iamthebestm85/gocf
```

### Prerequisites

- Go 1.22+
- Chrome/Chromium installed (for turnstile solving)
- **Headful mode recommended** - Cloudflare's managed challenges may block headless browsers. Use `xvfb-run` on servers without a display.
- **Residential proxies** - Cloudflare often blocks datacenter IPs. Residential proxies work better for bypassing managed challenges.

## Quick Start

### Solve a Managed Challenge (Recommended)

```go
package main

import (
    "context"
    "fmt"
    "time"
    "github.com/iamthebestm85/gocf"
)

func main() {
    ctx := context.Background()
    
    solver := gocf.New(
        gocf.WithTimeout(90*time.Second),
        gocf.WithHeadless(false), // headful mode for best results
    )
    
    result, err := solver.Solve(ctx, "https://captcha.dstatbot.win/NyUrHdOG")
    if err != nil {
        panic(err)
    }
    
    fmt.Println("Token:", result.Token)
    fmt.Println("CF Clearance:", result.CFClearance)
    fmt.Println("Cookies:", result.Cookies)
    fmt.Println("Challenge Solved:", result.ChallengeSolved)
    fmt.Println("Time:", result.Elapsed)
}
```

### Using the CLI

```bash
# Solve the target URL
xvfb-run go run ./cmd/solve/main.go

# With custom options
go run ./cmd/solve/main.go \
    -url "https://captcha.dstatbot.win/NyUrHdOG" \
    -timeout 90s \
    -output result.txt

# With proxy
go run ./cmd/solve/main.go -proxy "socks5://127.0.0.1:1080"
```

### Solve and Request

```go
result, resp, err := gocf.SolveAndRequest(ctx, "https://example.com",
    gocf.WithTimeout(90*time.Second),
)
if err != nil {
    panic(err)
}
defer resp.Body.Close()
fmt.Println("Status:", resp.StatusCode)
```

### With Proxy

```go
result, err := gocf.Solve(ctx, "https://example.com",
    gocf.WithProxy("http://user:pass@127.0.0.1:8080"),
)
```

### Batch Parallel Solving

```go
urls := []string{"https://example.com", "https://example.org"}
results := gocf.BatchSolve(ctx, urls, 5,
    gocf.WithTimeout(90*time.Second),
)
```

## API Reference

### Options

| Option | Description | Default |
|--------|-------------|---------|
| `WithProxy(url)` | HTTP/SOCKS5 proxy | none |
| `WithTimeout(d)` | Request timeout | 90s |
| `WithHeadless(bool)` | Headless browser | false (headful) |
| `WithUserAgent(ua)` | Custom User-Agent | Chrome 138 |
| `WithViewport(w, h)` | Browser viewport | 1920x1080 |
| `WithDisableImages(bool)` | Disable image loading | false |

### Methods

- `Solve(ctx, url)` - Solve turnstile/managed challenge on URL
- `SolveWithWidget(ctx, url, sitekey)` - Solve with specific sitekey
- `SolveAll(ctx, urls, workers)` - Batch parallel solving
- `NewClient(platform)` - Create HTTP client with fingerprint
- `NewClientNoRedirect(platform)` - Create client that doesn't follow redirects
- `GetWithToken(ctx, url, token, cookies)` - Request with solved token
- `SolveAndRequest(ctx, url, ...opts)` - Solve and make follow-up request

### Result Fields

| Field | Description |
|-------|-------------|
| `Token` | The turnstile response token |
| `CFClearance` | The cf_clearance cookie value (if obtained) |
| `Cookies` | All cookies as a string |
| `ChallengeSolved` | Whether the challenge was fully solved |
| `FinalURL` | The final URL after any redirects |
| `Elapsed` | Time taken to solve |
| `UserAgent` | Browser user agent used |

## Supported Platforms

- Windows (Chrome 138)
- macOS (Chrome 138)
- Linux (Chrome 138)
- Android (Chrome 138, Pixel 8)

## Running on a Server (No Display)

Use `xvfb-run` to run Chrome headful mode without a display:

```bash
apt-get install -y xvfb
xvfb-run go run ./cmd/solve/main.go
```

## Important Notes

- **Headful mode recommended** - Cloudflare's managed challenges may block headless browsers. The default is headful mode (`Headless: false`). Use `xvfb-run` on servers without a display.
- **Residential proxies** - Cloudflare often blocks datacenter IPs. Residential proxies work better for bypassing managed challenges.
- **Timeout** - Managed challenges can take up to 60-90 seconds. The default timeout is 90 seconds.

## License

MIT
