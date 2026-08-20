package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"

	"github.com/iamthebestm85/gocf"
)

func main() {
	url := flag.String("url", "https://captcha.dstatbot.win/NyUrHdOG", "Target URL to solve")
	headless := flag.Bool("headless", false, "Run browser in headless mode")
	timeout := flag.Duration("timeout", 90*time.Second, "Solve timeout")
	proxy := flag.String("proxy", "", "HTTP/SOCKS5 proxy URL")
	output := flag.String("output", "", "Output file for response body")
	verbose := flag.Bool("verbose", true, "Verbose output")
	flag.Parse()

	ctx := context.Background()

	if *verbose {
		fmt.Printf("Target: %s\n", *url)
		fmt.Printf("Headless: %v\n", *headless)
		fmt.Printf("Timeout: %v\n", *timeout)
		if *proxy != "" {
			fmt.Printf("Proxy: %s\n", *proxy)
		}
		fmt.Println()
	}

	opts := []gocf.Option{
		gocf.WithTimeout(*timeout),
		gocf.WithHeadless(*headless),
	}
	if *proxy != "" {
		opts = append(opts, gocf.WithProxy(*proxy))
	}

	fmt.Println("[*] Launching browser and solving challenge...")

	solver := gocf.New(opts...)
	result, err := solver.Solve(ctx, *url)
	if err != nil {
		log.Fatalf("[!] Solve failed: %v", err)
	}

	fmt.Println("[+] Challenge solve completed!")
	fmt.Println()
	fmt.Printf("  Challenge Solved:   %v\n", result.ChallengeSolved)
	fmt.Printf("  Token Length:       %d\n", len(result.Token))
	fmt.Printf("  Token:              %s...\n", truncate(result.Token, 60))
	fmt.Printf("  CF Clearance:       %s\n", truncate(result.CFClearance, 50))
	fmt.Printf("  Cookies Length:     %d\n", len(result.Cookies))
	fmt.Printf("  Elapsed:            %v\n", result.Elapsed)
	fmt.Printf("  User Agent:         %s\n", result.UserAgent)
	if result.FinalURL != "" {
		fmt.Printf("  Final URL:          %s\n", result.FinalURL)
	}

	// Now make a follow-up request with the solved cookies
	fmt.Println()
	fmt.Println("[*] Making follow-up request with solved cookies...")

	result2, resp, err := gocf.SolveAndRequest(ctx, *url, opts...)
	if err != nil {
		log.Fatalf("[!] Follow-up request failed: %v", err)
	}
	defer resp.Body.Close()

	fmt.Printf("[+] Response Status: %d\n", resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("[!] Failed to read response body: %v", err)
	}

	fmt.Printf("[+] Response Body Length: %d bytes\n", len(body))

	// Print response headers
	fmt.Println()
	fmt.Println("[*] Response Headers:")
	for k, v := range resp.Header {
		fmt.Printf("  %s: %s\n", k, strings.Join(v, ", "))
	}

	// Print body preview
	if len(body) > 0 {
		bodyStr := string(body)
		fmt.Println()
		fmt.Printf("[*] Body Preview (first 500 chars):\n%s\n", truncate(bodyStr, 500))
	}

	// Save to file if requested
	if *output != "" {
		f, err := os.Create(*output)
		if err != nil {
			log.Printf("[!] Failed to create output file: %v", err)
		} else {
			defer f.Close()
			f.WriteString(fmt.Sprintf("URL: %s\n", *url))
			f.WriteString(fmt.Sprintf("Solved: %v\n", result2.ChallengeSolved))
			f.WriteString(fmt.Sprintf("Token: %s\n", result2.Token))
			f.WriteString(fmt.Sprintf("CF Clearance: %s\n", result2.CFClearance))
			f.WriteString(fmt.Sprintf("Cookies: %s\n", result2.Cookies))
			f.WriteString(fmt.Sprintf("Status: %d\n", resp.StatusCode))
			f.WriteString(fmt.Sprintf("Body: %s\n", string(body)))
			fmt.Printf("\n[+] Output saved to: %s\n", *output)
		}
	}

	fmt.Println()
	fmt.Println("[+] Done!")
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "..."
}
