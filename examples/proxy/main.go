package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/iamthebestm85/gocf"
	"github.com/iamthebestm85/gocf/fingerprint"
)

func main() {
	ctx := context.Background()

	fmt.Println("=== GOCF Proxy Demo ===")
	fmt.Println()

	// Proxy formats supported:
	//   http://user:pass@host:port
	//   socks5://user:pass@host:port
	//   http://host:port
	//   socks5://host:port

	proxy := "http://user:pass@127.0.0.1:8080"

	fmt.Println("[1] Solving with proxy...")
	solver := gocf.New(
		gocf.WithProxy(proxy),
		gocf.WithTimeout(45*time.Second),
		gocf.WithHeadless(true),
	)

	result, err := solver.Solve(ctx, "https://example.com")
	if err != nil {
		log.Printf("Error: %v", err)
	} else {
		fmt.Printf("  Token: %s...\n", truncate(result.Token, 50))
		fmt.Printf("  Cookies: %s\n", truncate(result.Cookies, 80))
		fmt.Printf("  Time: %v\n", result.Elapsed)
	}

	fmt.Println()

	// Use proxy with HTTP client
	fmt.Println("[2] Making request with proxy...")
	c, err := solver.NewClient(fingerprint.Windows)
	if err != nil {
		log.Printf("Client error: %v", err)
	} else {
		resp, err := c.DoWithCookies("https://httpbin.org/ip", result.Cookies)
		if err != nil {
			log.Printf("Request error: %v", err)
		} else {
			defer resp.Body.Close()
			body, _ := io.ReadAll(resp.Body)
			fmt.Printf("  Response: %s\n", string(body))
		}
	}

	fmt.Println()

	// Batch with proxy
	fmt.Println("[3] Batch solving with proxy...")
	urls := []string{
		"https://example.com",
		"https://httpbin.org/get",
		"https://httpbin.org/ip",
	}

	results := gocf.BatchSolve(ctx, urls, 2,
		gocf.WithProxy(proxy),
		gocf.WithTimeout(60*time.Second),
	)

	for i, r := range results {
		if r.Err != nil {
			fmt.Printf("  [%d] Error: %v\n", i, r.Err)
		} else {
			fmt.Printf("  [%d] Token: %s... (%v)\n", i, truncate(r.Token, 30), r.Elapsed)
		}
	}

	fmt.Println()
	fmt.Println("Done!")
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "..."
}
