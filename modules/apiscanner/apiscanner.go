package apiscanner

import (
	"bufio"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/isa-programmer/goscan/modules/config"
	"github.com/isa-programmer/goscan/modules/logger"
)

// APIResult represents an API scan result
type APIResult struct {
	URL          string            `json:"url"`
	Method       string            `json:"method"`
	StatusCode   int               `json:"status_code"`
	Size         int64             `json:"size"`
	Timestamp    time.Time         `json:"timestamp"`
	ResponseTime time.Duration     `json:"response_time"`
	Headers      map[string]string `json:"headers,omitempty"`
	ContentType  string            `json:"content_type,omitempty"`
}

// APIScanner handles API endpoint scanning
type APIScanner struct {
	config     *config.Config
	logger     *logger.Logger
	client     *http.Client
	endpoints  []string
	results    []APIResult
	mutex      sync.Mutex
	methods    []string
}

// New creates a new APIScanner instance
func New(cfg *config.Config, log *logger.Logger) *APIScanner {
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

	return &APIScanner{
		config:  cfg,
		logger:  log,
		client:  client,
		methods: []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS", "HEAD"},
	}
}

// LoadEndpoints loads API endpoints from wordlist
func (a *APIScanner) LoadEndpoints() error {
	file, err := os.Open(a.config.WordlistPath)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		endpoint := strings.TrimSpace(scanner.Text())
		if endpoint != "" && !strings.HasPrefix(endpoint, "#") {
			// Ensure endpoint starts with /
			if !strings.HasPrefix(endpoint, "/") {
				endpoint = "/" + endpoint
			}
			a.endpoints = append(a.endpoints, endpoint)
		}
	}

	return scanner.Err()
}

// ScanAPI scans API endpoints with different HTTP methods
func (a *APIScanner) ScanAPI() ([]APIResult, error) {
	if err := a.LoadEndpoints(); err != nil {
		return nil, fmt.Errorf("failed to load endpoints: %v", err)
	}

	a.logger.Info(fmt.Sprintf("Loaded %d API endpoints", len(a.endpoints)))

	// Generate URLs to scan
	urls := a.generateAPIURLs()
	total := len(urls)

	a.logger.Info(fmt.Sprintf("Starting API scan with %d requests using %d threads", total, a.config.Threads))

	// Create worker pool
	urlChan := make(chan APIRequest, total)
	var wg sync.WaitGroup

	// Start workers
	for i := 0; i < a.config.Threads; i++ {
		wg.Add(1)
		go a.worker(&wg, urlChan)
	}

	// Send requests to workers
	for _, req := range urls {
		urlChan <- req
	}
	close(urlChan)

	// Wait for all workers to complete
	wg.Wait()

	a.logger.Info(fmt.Sprintf("API scan completed. Found %d responses", len(a.results)))

	return a.results, nil
}

// APIRequest represents an API request to be made
type APIRequest struct {
	URL    string
	Method string
}

// generateAPIURLs creates all API URLs to be scanned
func (a *APIScanner) generateAPIURLs() []APIRequest {
	var requests []APIRequest
	baseURL := a.config.GetBaseURL()

	for _, endpoint := range a.endpoints {
		for _, method := range a.methods {
			url := baseURL + endpoint
			requests = append(requests, APIRequest{
				URL:    url,
				Method: method,
			})
		}
	}

	return requests
}

// worker processes API requests from the channel
func (a *APIScanner) worker(wg *sync.WaitGroup, reqChan <-chan APIRequest) {
	defer wg.Done()

	for apiReq := range reqChan {
		a.scanAPIEndpoint(apiReq)
	}
}

// scanAPIEndpoint scans a single API endpoint
func (a *APIScanner) scanAPIEndpoint(apiReq APIRequest) {
	req, err := http.NewRequest(apiReq.Method, apiReq.URL, nil)
	if err != nil {
		a.logger.Debug(fmt.Sprintf("Failed to create request for %s %s: %v", apiReq.Method, apiReq.URL, err))
		return
	}

	// Set headers
	req.Header.Set("User-Agent", a.config.UserAgent)
	req.Header.Set("Accept", "application/json, application/xml, text/plain, */*")
	for key, value := range a.config.Headers {
		req.Header.Set(key, value)
	}

	// Make request and measure response time
	start := time.Now()
	resp, err := a.client.Do(req)
	responseTime := time.Since(start)
	if err != nil {
		a.logger.Debug(fmt.Sprintf("Request failed for %s %s: %v", apiReq.Method, apiReq.URL, err))
		return
	}
	defer resp.Body.Close()

	// Check if status code is in the list of valid codes
	validStatus := false
	for _, code := range a.config.StatusCodes {
		if resp.StatusCode == code {
			validStatus = true
			break
		}
	}

	// Only record interesting responses
	if validStatus || resp.StatusCode < 500 {
		result := APIResult{
			URL:          apiReq.URL,
			Method:       apiReq.Method,
			StatusCode:   resp.StatusCode,
			Size:         resp.ContentLength,
			Timestamp:    time.Now(),
			ResponseTime: responseTime,
			ContentType:  resp.Header.Get("Content-Type"),
			Headers:      make(map[string]string),
		}

		// Capture interesting headers
		interestingHeaders := []string{"Server", "X-Powered-By", "X-Framework", "X-API-Version", "Access-Control-Allow-Origin"}
		for _, header := range interestingHeaders {
			if value := resp.Header.Get(header); value != "" {
				result.Headers[header] = value
			}
		}

		a.mutex.Lock()
		a.results = append(a.results, result)
		a.mutex.Unlock()

		a.logger.Request(apiReq.Method, apiReq.URL, resp.StatusCode)
	}
}

// ExportResults exports results to JSON file
func (a *APIScanner) ExportResults(filename string) error {
	data, err := json.MarshalIndent(a.results, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filename, data, 0644)
}

// GetResults returns the scan results
func (a *APIScanner) GetResults() []APIResult {
	return a.results
}