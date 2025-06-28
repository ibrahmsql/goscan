package scanner

import (
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/isa-programmer/goscan/internal/config"
	"github.com/isa-programmer/goscan/pkg/stats"
)

// Result represents a scan result
type Result struct {
	URL        string    `json:"url"`
	StatusCode int       `json:"status_code"`
	Size       int64     `json:"size"`
	Words      int       `json:"words"`
	Lines      int       `json:"lines"`
	Redirect   string    `json:"redirect,omitempty"`
	Time       time.Duration `json:"time"`
	Headers    map[string]string `json:"headers,omitempty"`
	Body       string    `json:"body,omitempty"`
	Timestamp  time.Time `json:"timestamp"`
}

// Scanner handles HTTP requests and scanning logic
type Scanner struct {
	config     *config.Config
	client     *http.Client
	includeRegex *regexp.Regexp
	excludeRegex *regexp.Regexp
}

// New creates a new scanner instance
func New(cfg *config.Config) (*Scanner, error) {
	s := &Scanner{
		config: cfg,
		client: createHTTPClient(cfg),
	}

	// Compile regex patterns
	if cfg.IncludeRegex != "" {
		var err error
		s.includeRegex, err = regexp.Compile(cfg.IncludeRegex)
		if err != nil {
			return nil, fmt.Errorf("invalid include regex: %w", err)
		}
	}

	if cfg.ExcludeRegex != "" {
		var err error
		s.excludeRegex, err = regexp.Compile(cfg.ExcludeRegex)
		if err != nil {
			return nil, fmt.Errorf("invalid exclude regex: %w", err)
		}
	}

	return s, nil
}

// Scan performs the actual scanning
func (s *Scanner) Scan(urls []string, stats *stats.Stats) ([]*Result, error) {
	// Create semaphore for limiting concurrent requests
	sem := make(chan struct{}, s.config.Threads)
	var wg sync.WaitGroup
	var resultsMutex sync.Mutex
	var results []*Result

	// Rate limiter
	var rateLimiter <-chan time.Time
	if s.config.RateLimit > 0 {
		rateLimiter = time.Tick(time.Second / time.Duration(s.config.RateLimit))
	}

	// Process URLs
	for _, targetURL := range urls {
		wg.Add(1)
		go func(url string) {
			defer wg.Done()

			// Rate limiting
			if rateLimiter != nil {
				<-rateLimiter
			}

			// Acquire semaphore
			sem <- struct{}{}
			defer func() { <-sem }()

			// Add delay if configured
			if s.config.Delay > 0 {
				time.Sleep(s.config.Delay)
			}

			// Make request with retries
			var result *Result
			var err error
			for attempt := 0; attempt <= s.config.Retries; attempt++ {
				result, err = s.makeRequest(url)
				if err == nil {
					break
				}
				if attempt < s.config.Retries {
					time.Sleep(time.Duration(attempt+1) * time.Second)
				}
			}

			stats.IncrementTested()

			if err != nil {
				stats.IncrementErrors()
				if s.config.Verbose {
					fmt.Printf("\nError: %s - %v\n", url, err)
				}
				return
			}

			// Check if we should show this result
			if s.shouldShow(result) {
				stats.IncrementFound()

				if !s.config.Quiet {
					fmt.Printf("\n%s\n", s.formatResult(result))
				}

				resultsMutex.Lock()
				results = append(results, result)
				resultsMutex.Unlock()
			}
		}(targetURL)
	}

	// Wait for all requests to complete
	wg.Wait()

	return results, nil
}

// makeRequest performs a single HTTP request
func (s *Scanner) makeRequest(targetURL string) (*Result, error) {
	start := time.Now()

	req, err := http.NewRequest(s.config.Method, targetURL, nil)
	if err != nil {
		return nil, err
	}

	// Set headers
	req.Header.Set("User-Agent", s.config.UserAgent)
	for key, value := range s.config.Headers {
		req.Header.Set(key, value)
	}

	// Set cookies
	if s.config.Cookies != "" {
		req.Header.Set("Cookie", s.config.Cookies)
	}

	// Set authentication
	if s.config.BasicAuth != "" {
		parts := strings.SplitN(s.config.BasicAuth, ":", 2)
		if len(parts) == 2 {
			req.SetBasicAuth(parts[0], parts[1])
		}
	}

	if s.config.BearerToken != "" {
		req.Header.Set("Authorization", "Bearer "+s.config.BearerToken)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	size := int64(len(body))
	duration := time.Since(start)

	// Count words and lines
	bodyStr := string(body)
	words := len(strings.Fields(bodyStr))
	lines := len(strings.Split(bodyStr, "\n"))

	result := &Result{
		URL:        targetURL,
		StatusCode: resp.StatusCode,
		Size:       size,
		Words:      words,
		Lines:      lines,
		Time:       duration,
		Timestamp:  time.Now(),
	}

	// Store body if verbose
	if s.config.Verbose {
		result.Body = bodyStr
		result.Headers = make(map[string]string)
		for key, values := range resp.Header {
			result.Headers[key] = strings.Join(values, ", ")
		}
	}

	// Handle redirects
	if resp.StatusCode >= 300 && resp.StatusCode < 400 {
		if location := resp.Header.Get("Location"); location != "" {
			result.Redirect = location
		}
	}

	return result, nil
}

// shouldShow determines if a result should be displayed
func (s *Scanner) shouldShow(result *Result) bool {
	// Check status codes to show
	if len(s.config.StatusCodes) > 0 {
		found := false
		for _, code := range s.config.StatusCodes {
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

	// Check status codes to hide
	for _, code := range s.config.HideStatusCodes {
		if result.StatusCode == code {
			return false
		}
	}

	// Check hidden lengths
	for _, length := range s.config.HideLength {
		if int(result.Size) == length {
			return false
		}
	}

	// Check hidden words
	for _, words := range s.config.HideWords {
		if result.Words == words {
			return false
		}
	}

	// Check hidden lines
	for _, lines := range s.config.HideLines {
		if result.Lines == lines {
			return false
		}
	}

	// Check regex patterns
	if s.includeRegex != nil {
		if !s.includeRegex.MatchString(result.Body) {
			return false
		}
	}

	if s.excludeRegex != nil {
		if s.excludeRegex.MatchString(result.Body) {
			return false
		}
	}

	// Check string matching
	if s.config.MatchString != "" {
		if !strings.Contains(result.Body, s.config.MatchString) {
			return false
		}
	}

	if s.config.ExcludeString != "" {
		if strings.Contains(result.Body, s.config.ExcludeString) {
			return false
		}
	}

	return true
}

// formatResult formats a result for display
func (s *Scanner) formatResult(result *Result) string {
	output := fmt.Sprintf("[%d] %s", result.StatusCode, result.URL)

	if s.config.ShowSize {
		output += fmt.Sprintf(" [Size: %d]", result.Size)
	}

	if s.config.ShowLength {
		output += fmt.Sprintf(" [Length: %d]", result.Size)
	}

	if s.config.ShowWords {
		output += fmt.Sprintf(" [Words: %d]", result.Words)
	}

	if s.config.ShowLines {
		output += fmt.Sprintf(" [Lines: %d]", result.Lines)
	}

	if result.Redirect != "" {
		output += fmt.Sprintf(" -> %s", result.Redirect)
	}

	if s.config.ShowTime || s.config.Verbose {
		output += fmt.Sprintf(" [Time: %v]", result.Time.Round(time.Millisecond))
	}

	return output
}

// createHTTPClient creates and configures HTTP client
func createHTTPClient(cfg *config.Config) *http.Client {
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: cfg.Insecure,
		},
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 100,
		IdleConnTimeout:     90 * time.Second,
	}

	// Configure proxy
	if cfg.Proxy != "" {
		proxyURL, err := url.Parse(cfg.Proxy)
		if err == nil {
			transport.Proxy = http.ProxyURL(proxyURL)
		}
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   cfg.Timeout,
	}

	// Configure redirect handling
	if !cfg.FollowRedirect {
		client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}
	} else if cfg.MaxRedirects > 0 {
		client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			if len(via) >= cfg.MaxRedirects {
				return fmt.Errorf("stopped after %d redirects", cfg.MaxRedirects)
			}
			return nil
		}
	}

	return client
}