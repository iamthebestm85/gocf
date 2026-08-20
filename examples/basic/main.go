package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/iamthebestm85/gocf"
)

func main() {
	ctx := context.Background()

	fmt.Println("=== GOCF Basic Demo ===")
	fmt.Println()

	targetURL := "https://captcha.dstatbot.win/NyUrHdOG"

	// 1. Solve a Cloudflare managed challenge page
	fmt.Println("[1] Solving managed challenge page...")
	solver := gocf.New(
		gocf.WithTimeout(90*time.Second),
		gocf.WithHeadless(false),
	)

	result, err := solver.Solve(ctx, targetURL)
	if err != nil {
		log.Printf("Error: %v", err)
	} else {
		fmt.Printf("  Token: %s...\n", truncate(result.Token, 50))
		fmt.Printf("  Cookies: %s\n", truncate(result.Cookies, 80))
		fmt.Printf("  CF Clearance: %s\n", truncate(result.CFClearance, 50))
		fmt.Printf("  Challenge Solved: %v\n", result.ChallengeSolved)
		fmt.Printf("  Final URL: %s\n", result.FinalURL)
		fmt.Printf("  Time: %v\n", result.Elapsed)
	}

	fmt.Println()

	// 2. Solve and make request in one go
	fmt.Println("[2] Solve and request...")
	result2, resp, err := gocf.SolveAndRequest(ctx, targetURL,
		gocf.WithTimeout(90*time.Second),
		gocf.WithHeadless(false),
	)
	if err != nil {
		log.Printf("Error: %v", err)
	} else {
		fmt.Printf("  Token: %s...\n", truncate(result2.Token, 50))
		fmt.Printf("  Response Status: %d\n", resp.StatusCode)
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		fmt.Printf("  Body length: %d bytes\n", len(body))
		if len(body) > 0 {
			fmt.Printf("  Body preview: %s\n", truncate(string(body), 200))
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
