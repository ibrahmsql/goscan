package scanner

import (
	"bufio"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/isa-programmer/goscan/modules/config"
	"github.com/isa-programmer/goscan/modules/logger"
)

// Result represents a scan result
type Result struct {
	URL          string        `json:"url"`
	StatusCode   int           `json:"status_code"`
	Size         int64         `json:"size"`
	Timestamp    time.Time     `json:"timestamp"`
	Redirect     string        `json:"redirect,omitempty"`
	ResponseTime time.Duration `json:"response_time"`
}

// Scanner handles the directory scanning operations
type Scanner struct {
	config     *config.Config
	logger     *logger.Logger
	client     *http.Client
	wordlist   []string
	results    []Result
	mutex      sync.Mutex
	stats      *Statistics
}

// Statistics holds scanning statistics
type Statistics struct {
	Total     int
	Found     int
	Errors    int
	StartTime time.Time
	mutex     sync.Mutex
}

// New returns a new Scanner instance configured with the provided settings and logger.
// It validates the configuration, sets up an HTTP client with the specified timeout, SSL, and redirect behavior, and initializes scanning statistics.
func New(cfg *config.Config, log *logger.Logger) *Scanner {
	// Validate configuration
	if err := cfg.Validate(); err != nil {
		log.Fatal(fmt.Sprintf("Configuration error: %v", err))
	}

	// Create HTTP client
	client := &http.Client{
		Timeout: time.Duration(cfg.Timeout) * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: cfg.IgnoreSSL,
			},
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 10,
		},
	}

	// Configure redirects
	if !cfg.FollowRedirects {
		client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}
	}

	return &Scanner{
		config: cfg,
		logger: log,
		client: client,
		stats: &Statistics{
			StartTime: time.Now(),
		},
	}
}

// Run starts the scanning process
func (s *Scanner) Run() ([]Result, error) {
	// Load wordlist
	if err := s.loadWordlist(); err != nil {
		return nil, fmt.Errorf("failed to load wordlist: %v", err)
	}

	s.logger.Info(fmt.Sprintf("Loaded %d words from wordlist", len(s.wordlist)))

	// Generate URLs to scan
	urls := s.generateURLs()
	s.stats.Total = len(urls)

	s.logger.Info(fmt.Sprintf("Starting scan with %d URLs using %d threads", len(urls), s.config.Threads))

	// Create worker pool
	urlChan := make(chan string, len(urls))
	var wg sync.WaitGroup

	// Start workers
	for i := 0; i < s.config.Threads; i++ {
		wg.Add(1)
		go s.worker(&wg, urlChan)
	}

	// Send URLs to workers
	for _, url := range urls {
		urlChan <- url
	}
	close(urlChan)

	// Start progress reporter
	if s.config.Verbose {
		go s.progressReporter()
	}

	// Wait for all workers to complete
	wg.Wait()

	// Final statistics
	elapsed := time.Since(s.stats.StartTime)
	s.logger.Statistics(s.stats.Found, s.stats.Total, elapsed)

	return s.results, nil
}

// loadWordlist loads words from the wordlist file
func (s *Scanner) loadWordlist() error {
	file, err := os.Open(s.config.WordlistPath)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		word := strings.TrimSpace(scanner.Text())
		if word != "" && !strings.HasPrefix(word, "#") {
			s.wordlist = append(s.wordlist, word)
		}
	}

	return scanner.Err()
}

// generateURLs creates all URLs to be scanned
func (s *Scanner) generateURLs() []string {
	var urls []string
	baseURL := s.config.GetBaseURL()

	for _, word := range s.wordlist {
		for _, ext := range s.config.Extensions {
			url := fmt.Sprintf("%s/%s%s", baseURL, word, ext)
			urls = append(urls, url)
		}
	}

	return urls
}

// worker processes URLs from the channel
func (s *Scanner) worker(wg *sync.WaitGroup, urlChan <-chan string) {
	defer wg.Done()

	for url := range urlChan {
		s.scanURL(url)
	}
}

// scanURL scans a single URL
func (s *Scanner) scanURL(url string) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		s.logger.Debug(fmt.Sprintf("Failed to create request for %s: %v", url, err))
		s.incrementErrors()
		return
	}

	// Set headers
	req.Header.Set("User-Agent", s.config.UserAgent)
	for key, value := range s.config.Headers {
		req.Header.Set(key, value)
	}

	// Make request and measure response time
	start := time.Now()
	resp, err := s.client.Do(req)
	responseTime := time.Since(start)
	if err != nil {
		s.logger.Debug(fmt.Sprintf("Request failed for %s: %v", url, err))
		s.incrementErrors()
		return
	}
	defer resp.Body.Close()

	// Log request details
	s.logger.Request("GET", url, resp.StatusCode)

	// Check if status code is interesting
	if s.isInterestingStatus(resp.StatusCode) {
		// Read response body to get size
		body, _ := ioutil.ReadAll(resp.Body)
		size := int64(len(body))

		// Get redirect location if applicable
		redirect := ""
		if resp.StatusCode >= 300 && resp.StatusCode < 400 {
			redirect = resp.Header.Get("Location")
		}

		// Create result
		result := Result{
			URL:          url,
			StatusCode:   resp.StatusCode,
			Size:         size,
			Timestamp:    time.Now(),
			Redirect:     redirect,
			ResponseTime: responseTime,
		}

		// Add to results
		s.mutex.Lock()
		s.results = append(s.results, result)
		s.mutex.Unlock()

		s.incrementFound()
		s.logger.Success(fmt.Sprintf("[%d] %s (Size: %d)", resp.StatusCode, url, size))
	}
}

// isInterestingStatus checks if a status code is worth reporting
func (s *Scanner) isInterestingStatus(statusCode int) bool {
	for _, code := range s.config.StatusCodes {
		if statusCode == code {
			return true
		}
	}
	return false
}

// incrementFound safely increments the found counter
func (s *Scanner) incrementFound() {
	s.stats.mutex.Lock()
	s.stats.Found++
	s.stats.mutex.Unlock()
}

// incrementErrors safely increments the error counter
func (s *Scanner) incrementErrors() {
	s.stats.mutex.Lock()
	s.stats.Errors++
	s.stats.mutex.Unlock()
}

// progressReporter shows progress updates
func (s *Scanner) progressReporter() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		s.stats.mutex.Lock()
		processed := s.stats.Found + s.stats.Errors
		if processed >= s.stats.Total {
			s.stats.mutex.Unlock()
			return
		}
		progress := float64(processed) / float64(s.stats.Total) * 100
		elapsed := time.Since(s.stats.StartTime)
		rate := float64(processed) / elapsed.Seconds()
		s.stats.mutex.Unlock()

		s.logger.Progress(fmt.Sprintf("Progress: %.1f%% (%d/%d) | Rate: %.1f req/s | Found: %d", 
			progress, processed, s.stats.Total, rate, s.stats.Found))
	}
}

// SaveResults writes scan results to a file in either JSON or plain text format, determined by the file extension.
// If the filename ends with ".json", results are saved as JSON; otherwise, results are saved as plain text.
// Returns an error if writing to the file fails.
func SaveResults(results []Result, filename string) error {
	ext := strings.ToLower(filepath.Ext(filename))

	switch ext {
	case ".json":
		return saveAsJSON(results, filename)
	case ".txt":
		return saveAsText(results, filename)
	default:
		return saveAsText(results, filename)
	}
}

// saveAsJSON writes the provided scan results to a file in indented JSON format.
// Returns an error if marshalling or file writing fails.
func saveAsJSON(results []Result, filename string) error {
	data, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		return err
	}
	return ioutil.WriteFile(filename, data, 0644)
}

// saveAsText writes scan results to a file in plain text format, listing each result's status code and URL, and including redirect locations if present.
// Returns an error if the file cannot be created or written.
func saveAsText(results []Result, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	for _, result := range results {
		line := fmt.Sprintf("[%d] %s\n", result.StatusCode, result.URL)
		if result.Redirect != "" {
			line = fmt.Sprintf("[%d] %s -> %s\n", result.StatusCode, result.URL, result.Redirect)
		}
		_, err := file.WriteString(line)
		if err != nil {
			return err
		}
	}

	return nil
}