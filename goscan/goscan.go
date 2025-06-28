// Name: goscan
// Description: A cross-platform directory scanner written in Golang
// Author: isa-programmer
// Repository: https://github.com/isa-programmer/goscan
// LICENSE: MIT

package main

import (
	"bufio"
	"crypto/tls"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type Config struct {
	Wordlist     string
	Target       string
	Threads      int
	Timeout      time.Duration
	Method       string
	UserAgent    string
	Headers      map[string]string
	Proxy        string
	Extensions   []string
	StatusCodes  []int
	HideLength   []int
	ShowLength   bool
	ShowSize     bool
	Verbose      bool
	Quiet        bool
	FollowRedirect bool
	Insecure     bool
	Delay        time.Duration
}

type Result struct {
	URL        string
	StatusCode int
	Size       int64
	Redirect   string
	Time       time.Duration
}

type Stats struct {
	Total     int64
	Tested    int64
	Found     int64
	Errors    int64
	StartTime time.Time
}

func getWordsFromFile(path string) ([]string, error) {
	var words []string
	file, err := os.Open(path)
	if err != nil {
		return words, err
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		word := strings.TrimSpace(scanner.Text())
		if word != "" && !strings.HasPrefix(word, "#") {
			words = append(words, word)
		}
	}
	return words, nil
}

func createHTTPClient(config *Config) *http.Client {
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: config.Insecure,
		},
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 100,
		IdleConnTimeout:     90 * time.Second,
	}

	if config.Proxy != "" {
		proxyURL, err := url.Parse(config.Proxy)
		if err == nil {
			transport.Proxy = http.ProxyURL(proxyURL)
		}
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   config.Timeout,
	}

	if !config.FollowRedirect {
		client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}
	}

	return client
}

func makeRequest(client *http.Client, targetURL string, config *Config) (*Result, error) {
	start := time.Now()
	
	req, err := http.NewRequest(config.Method, targetURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", config.UserAgent)
	for key, value := range config.Headers {
		req.Header.Set(key, value)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	size := int64(len(body))
	duration := time.Since(start)

	result := &Result{
		URL:        targetURL,
		StatusCode: resp.StatusCode,
		Size:       size,
		Time:       duration,
	}

	if resp.StatusCode >= 300 && resp.StatusCode < 400 {
		if location := resp.Header.Get("Location"); location != "" {
			result.Redirect = location
		}
	}

	return result, nil
}

func shouldShow(result *Result, config *Config) bool {
	// Check status codes
	if len(config.StatusCodes) > 0 {
		found := false
		for _, code := range config.StatusCodes {
			if result.StatusCode == code {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	} else {
		// Default: show 200-399 status codes
		if result.StatusCode < 200 || result.StatusCode >= 400 {
			return false
		}
	}

	// Check hidden lengths
	for _, length := range config.HideLength {
		if int(result.Size) == length {
			return false
		}
	}

	return true
}

func formatResult(result *Result, config *Config) string {
	output := fmt.Sprintf("[%d] %s", result.StatusCode, result.URL)
	
	if config.ShowSize {
		output += fmt.Sprintf(" [Size: %d]", result.Size)
	}
	
	if config.ShowLength {
		output += fmt.Sprintf(" [Length: %d]", result.Size)
	}
	
	if result.Redirect != "" {
		output += fmt.Sprintf(" -> %s", result.Redirect)
	}
	
	if config.Verbose {
		output += fmt.Sprintf(" [Time: %v]", result.Time.Round(time.Millisecond))
	}
	
	return output
}

func printProgress(stats *Stats, total int64) {
	if total == 0 {
		return
	}
	
	tested := atomic.LoadInt64(&stats.Tested)
	found := atomic.LoadInt64(&stats.Found)
	errors := atomic.LoadInt64(&stats.Errors)
	elapsed := time.Since(stats.StartTime)
	progress := float64(tested) / float64(total) * 100
	rps := float64(tested) / elapsed.Seconds()
	
	fmt.Printf("\rProgress: %.1f%% (%d/%d) | Found: %d | Errors: %d | RPS: %.1f", 
		progress, tested, total, found, errors, rps)
}

func parseStatusCodes(input string) []int {
	if input == "" {
		return nil
	}
	
	var codes []int
	parts := strings.Split(input, ",")
	for _, part := range parts {
		if code, err := strconv.Atoi(strings.TrimSpace(part)); err == nil {
			codes = append(codes, code)
		}
	}
	return codes
}

func parseIntSlice(input string) []int {
	if input == "" {
		return nil
	}
	
	var nums []int
	parts := strings.Split(input, ",")
	for _, part := range parts {
		if num, err := strconv.Atoi(strings.TrimSpace(part)); err == nil {
			nums = append(nums, num)
		}
	}
	return nums
}

func parseHeaders(input string) map[string]string {
	headers := make(map[string]string)
	if input == "" {
		return headers
	}
	
	parts := strings.Split(input, ",")
	for _, part := range parts {
		if kv := strings.SplitN(strings.TrimSpace(part), ":", 2); len(kv) == 2 {
			headers[strings.TrimSpace(kv[0])] = strings.TrimSpace(kv[1])
		}
	}
	return headers
}

func main() {
	config := &Config{}
	
	// Command line flags
	flag.StringVar(&config.Wordlist, "w", "", "Wordlist file path (required)")
	flag.StringVar(&config.Target, "u", "", "Target URL (required)")
	flag.IntVar(&config.Threads, "t", 50, "Number of concurrent threads")
	timeoutFlag := flag.Int("timeout", 10, "HTTP timeout in seconds")
	flag.StringVar(&config.Method, "m", "GET", "HTTP method")
	flag.StringVar(&config.UserAgent, "a", "goscan/2.0 (Advanced Directory Scanner)", "User-Agent string")
	headersFlag := flag.String("H", "", "Custom headers (format: 'Header1:Value1,Header2:Value2')")
	flag.StringVar(&config.Proxy, "p", "", "Proxy URL (http://proxy:port)")
	extensionsFlag := flag.String("x", "", "File extensions (comma separated)")
	statusCodesFlag := flag.String("s", "", "Status codes to show (comma separated)")
	hideLengthFlag := flag.String("hide-length", "", "Hide responses with these lengths (comma separated)")
	flag.BoolVar(&config.ShowLength, "l", false, "Show response length")
	flag.BoolVar(&config.ShowSize, "size", true, "Show response size")
	flag.BoolVar(&config.Verbose, "v", false, "Verbose output")
	flag.BoolVar(&config.Quiet, "q", false, "Quiet mode (no progress)")
	flag.BoolVar(&config.FollowRedirect, "r", false, "Follow redirects")
	flag.BoolVar(&config.Insecure, "k", false, "Skip SSL certificate verification")
	delayFlag := flag.Int("delay", 0, "Delay between requests in milliseconds")
	
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "GoScan v2.0 - Advanced Directory Scanner\n\n")
		fmt.Fprintf(os.Stderr, "Usage: %s [options]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Required:\n")
		fmt.Fprintf(os.Stderr, "  -w string    Wordlist file path\n")
		fmt.Fprintf(os.Stderr, "  -u string    Target URL\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  %s -w wordlist.txt -u https://example.com\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -w dirs.txt -u https://example.com -t 100 -x php,html\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -w files.txt -u https://example.com -s 200,301,403 -v\n", os.Args[0])
	}
	
	flag.Parse()
	
	// Validate required arguments
	if config.Wordlist == "" || config.Target == "" {
		flag.Usage()
		os.Exit(1)
	}
	
	// Parse additional flags
	config.Timeout = time.Duration(*timeoutFlag) * time.Second
	config.Headers = parseHeaders(*headersFlag)
	config.StatusCodes = parseStatusCodes(*statusCodesFlag)
	config.HideLength = parseIntSlice(*hideLengthFlag)
	config.Delay = time.Duration(*delayFlag) * time.Millisecond
	
	if *extensionsFlag != "" {
		config.Extensions = strings.Split(*extensionsFlag, ",")
		for i := range config.Extensions {
			config.Extensions[i] = strings.TrimSpace(config.Extensions[i])
			if !strings.HasPrefix(config.Extensions[i], ".") {
				config.Extensions[i] = "." + config.Extensions[i]
			}
		}
	}
	
	// Ensure target URL has proper format
	if !strings.HasPrefix(config.Target, "http://") && !strings.HasPrefix(config.Target, "https://") {
		config.Target = "http://" + config.Target
	}
	if !strings.HasSuffix(config.Target, "/") {
		config.Target += "/"
	}
	
	// Load wordlist
	words, err := getWordsFromFile(config.Wordlist)
	if err != nil {
		fmt.Printf("Error loading wordlist: %v\n", err)
		os.Exit(1)
	}
	
	// Generate URLs to test
	var urls []string
	for _, word := range words {
		// Add base word
		urls = append(urls, config.Target+word)
		
		// Add with extensions
		for _, ext := range config.Extensions {
			urls = append(urls, config.Target+word+ext)
		}
	}
	
	if !config.Quiet {
		fmt.Printf("GoScan v2.0 - Advanced Directory Scanner\n")
		fmt.Printf("Target: %s\n", config.Target)
		fmt.Printf("Wordlist: %s (%d words)\n", config.Wordlist, len(words))
		fmt.Printf("URLs to test: %d\n", len(urls))
		fmt.Printf("Threads: %d\n", config.Threads)
		fmt.Printf("Timeout: %v\n", config.Timeout)
		fmt.Println("=")
	}
	
	// Initialize stats
	stats := &Stats{
		Total:     int64(len(urls)),
		StartTime: time.Now(),
	}
	
	// Create HTTP client
	client := createHTTPClient(config)
	
	// Create semaphore for limiting concurrent requests
	sem := make(chan struct{}, config.Threads)
	var wg sync.WaitGroup
	var resultsMutex sync.Mutex
	var results []Result
	
	// Progress ticker
	var progressTicker *time.Ticker
	if !config.Quiet {
		progressTicker = time.NewTicker(1 * time.Second)
		go func() {
			for range progressTicker.C {
				printProgress(stats, int64(len(urls)))
			}
		}()
	}
	
	// Process URLs
	for _, targetURL := range urls {
		wg.Add(1)
		go func(url string) {
			defer wg.Done()
			
			// Acquire semaphore
			sem <- struct{}{}
			defer func() { <-sem }()
			
			// Add delay if configured
			if config.Delay > 0 {
				time.Sleep(config.Delay)
			}
			
			// Make request
			result, err := makeRequest(client, url, config)
			atomic.AddInt64(&stats.Tested, 1)
			
			if err != nil {
				atomic.AddInt64(&stats.Errors, 1)
				if config.Verbose {
					fmt.Printf("\nError: %s - %v\n", url, err)
				}
				return
			}
			
			// Check if we should show this result
			if shouldShow(result, config) {
				atomic.AddInt64(&stats.Found, 1)
				
				if !config.Quiet {
					fmt.Printf("\n%s\n", formatResult(result, config))
				}
				
				resultsMutex.Lock()
				results = append(results, *result)
				resultsMutex.Unlock()
			}
		}(targetURL)
	}
	
	// Wait for all requests to complete
	wg.Wait()
	
	// Stop progress ticker
	if progressTicker != nil {
		progressTicker.Stop()
	}
	
	// Final progress update
	if !config.Quiet {
		printProgress(stats, int64(len(urls)))
		fmt.Println()
	}
	
	// Sort results by status code
	sort.Slice(results, func(i, j int) bool {
		return results[i].StatusCode < results[j].StatusCode
	})
	
	// Print summary
	elapsed := time.Since(stats.StartTime)
	fmt.Printf("\n=== SCAN COMPLETE ===\n")
	fmt.Printf("Total URLs: %d\n", len(urls))
	fmt.Printf("Found: %d\n", atomic.LoadInt64(&stats.Found))
	fmt.Printf("Errors: %d\n", atomic.LoadInt64(&stats.Errors))
	fmt.Printf("Time: %v\n", elapsed.Round(time.Second))
	fmt.Printf("Requests/sec: %.2f\n", float64(len(urls))/elapsed.Seconds())
}
