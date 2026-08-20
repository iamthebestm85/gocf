package main

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/iamthebestm85/gocf"
	"github.com/iamthebestm85/gocf/fingerprint"
	"github.com/iamthebestm85/gocf/solver"
)

func main() {
	ctx := context.Background()

	fmt.Println("=== GOCF Parallel Demo ===")
	fmt.Println()

	// 1. Using the pool directly
	fmt.Println("[1] Parallel solving with pool...")
	cfg := solver.DefaultConfig()
	cfg.Proxy = ""
	cfg.Timeout = 45 * time.Second

	pool := solver.NewPool(ctx, 5, cfg)
	pool.Start()

	urls := []string{
		"https://example.com",
		"https://example.org",
		"https://example.net",
		"https://httpbin.org/get",
		"https://httpbin.org/ip",
		"https://httpbin.org/user-agent",
		"https://httpbin.org/headers",
		"https://httpbin.org/cookies",
	}

	for i, url := range urls {
		pool.Submit(solver.SolveTask{URL: url, ID: i})
	}

	go func() {
		time.Sleep(5 * time.Second)
		pool.Close()
	}()

	for r := range pool.Results() {
		if r.Err != nil {
			fmt.Printf("  Error [%s]: %v\n", r.URL, r.Err)
		} else {
			fmt.Printf("  Success [%s]: token=%s... time=%v\n",
				r.URL, truncate(r.Token, 30), r.Elapsed)
		}
	}

	fmt.Println()

	// 2. Using BatchSolve
	fmt.Println("[2] Using BatchSolve...")
	results := gocf.BatchSolve(ctx, urls[:4], 3,
		gocf.WithTimeout(45*time.Second),
	)

	for i, r := range results {
		if r.Err != nil {
			fmt.Printf("  [%d] Error: %v\n", i, r.Err)
		} else {
			fmt.Printf("  [%d] Token: %s... (%v)\n", i, truncate(r.Token, 30), r.Elapsed)
		}
	}

	fmt.Println()

	// 3. Solve + Request in parallel
	fmt.Println("[3] Solve and request in parallel...")
	var wg sync.WaitGroup
	for i, url := range urls[:3] {
		wg.Add(1)
		go func(i int, url string) {
			defer wg.Done()
			result, resp, err := gocf.SolveAndRequest(ctx, url,
				gocf.WithTimeout(45*time.Second),
			)
			if err != nil {
				log.Printf("  [%d] Error: %v", i, err)
				return
			}
			fmt.Printf("  [%d] Token: %s... Status: %d\n",
				i, truncate(result.Token, 30), resp.StatusCode)
			resp.Body.Close()
		}(i, url)
	}
	wg.Wait()

	fmt.Println()

	// 4. Different platforms in parallel
	fmt.Println("[4] Different platforms in parallel...")
	platforms := []fingerprint.Platform{
		fingerprint.Windows,
		fingerprint.MacOS,
		fingerprint.Linux,
	}

	var wg2 sync.WaitGroup
	for _, platform := range platforms {
		wg2.Add(1)
		go func(p fingerprint.Platform) {
			defer wg2.Done()
			s := gocf.New(
				gocf.WithTimeout(30*time.Second),
			)
			c, err := s.NewClient(p)
			if err != nil {
				log.Printf("  [%s] Client error: %v", p, err)
				return
			}
			resp, err := c.Get("https://httpbin.org/user-agent")
			if err != nil {
				log.Printf("  [%s] Request error: %v", p, err)
				return
			}
			defer resp.Body.Close()
			fmt.Printf("  [%s] Status: %d\n", p, resp.StatusCode)
		}(platform)
	}
	wg2.Wait()

	fmt.Println()
	fmt.Println("Done!")
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "..."
}
