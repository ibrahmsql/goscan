package request

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// RequestConfig holds HTTP request configuration
type RequestConfig struct {
	UserAgent      string
	Timeout        time.Duration
	FollowRedirect bool
	MaxRedirects   int
	Headers        map[string]string
	ProxyURL       string
	IgnoreSSL      bool
	Cookies        []*http.Cookie
}

// RequestManager handles HTTP requests with custom configuration
type RequestManager struct {
	Client *http.Client
	Config *RequestConfig
}

// New creates a new RequestManager with default configuration
func New() *RequestManager {
	config := &RequestConfig{
		UserAgent:      "goscan/1.0.1 (https://github.com/isa-programmer/goscan)",
		Timeout:        10 * time.Second,
		FollowRedirect: false,
		MaxRedirects:   5,
		Headers:        make(map[string]string),
		IgnoreSSL:      false,
		Cookies:        make([]*http.Cookie, 0),
	}

	return &RequestManager{
		Config: config,
		Client: createHTTPClient(config),
	}
}

// createHTTPClient creates an HTTP client with the given configuration
func createHTTPClient(config *RequestConfig) *http.Client {
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: config.IgnoreSSL,
		},
	}

	// Set proxy if configured
	if config.ProxyURL != "" {
		proxyURL, err := url.Parse(config.ProxyURL)
		if err == nil {
			transport.Proxy = http.ProxyURL(proxyURL)
		}
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   config.Timeout,
	}

	// Configure redirect policy
	if !config.FollowRedirect {
		client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}
	} else {
		client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			if len(via) >= config.MaxRedirects {
				return fmt.Errorf("stopped after %d redirects", config.MaxRedirects)
			}
			return nil
		}
	}

	return client
}

// SetUserAgent sets the User-Agent header
func (rm *RequestManager) SetUserAgent(userAgent string) {
	rm.Config.UserAgent = userAgent
}

// SetTimeout sets the request timeout
func (rm *RequestManager) SetTimeout(timeout time.Duration) {
	rm.Config.Timeout = timeout
	rm.Client.Timeout = timeout
}

// SetProxy sets the proxy URL
func (rm *RequestManager) SetProxy(proxyURL string) error {
	rm.Config.ProxyURL = proxyURL
	rm.Client = createHTTPClient(rm.Config)
	return nil
}

// AddHeader adds a custom header
func (rm *RequestManager) AddHeader(key, value string) {
	rm.Config.Headers[key] = value
}

// AddCookie adds a cookie
func (rm *RequestManager) AddCookie(cookie *http.Cookie) {
	rm.Config.Cookies = append(rm.Config.Cookies, cookie)
}

// SetIgnoreSSL sets whether to ignore SSL certificate errors
func (rm *RequestManager) SetIgnoreSSL(ignore bool) {
	rm.Config.IgnoreSSL = ignore
	rm.Client = createHTTPClient(rm.Config)
}

// SetFollowRedirect sets whether to follow redirects
func (rm *RequestManager) SetFollowRedirect(follow bool) {
	rm.Config.FollowRedirect = follow
	rm.Client = createHTTPClient(rm.Config)
}

// MakeRequest makes an HTTP request with the configured settings
func (rm *RequestManager) MakeRequest(method, targetURL string) (*http.Response, error) {
	req, err := http.NewRequest(method, targetURL, nil)
	if err != nil {
		return nil, err
	}

	// Set User-Agent
	req.Header.Set("User-Agent", rm.Config.UserAgent)

	// Add custom headers
	for key, value := range rm.Config.Headers {
		req.Header.Set(key, value)
	}

	// Add cookies
	for _, cookie := range rm.Config.Cookies {
		req.AddCookie(cookie)
	}

	return rm.Client.Do(req)
}

// Get makes a GET request
func (rm *RequestManager) Get(targetURL string) (*http.Response, error) {
	return rm.MakeRequest("GET", targetURL)
}

// Head makes a HEAD request
func (rm *RequestManager) Head(targetURL string) (*http.Response, error) {
	return rm.MakeRequest("HEAD", targetURL)
}

// Post makes a POST request
func (rm *RequestManager) Post(targetURL string) (*http.Response, error) {
	return rm.MakeRequest("POST", targetURL)
}

// Options makes an OPTIONS request
func (rm *RequestManager) Options(targetURL string) (*http.Response, error) {
	return rm.MakeRequest("OPTIONS", targetURL)
}

// GetCommonUserAgents returns a list of common user agents
func GetCommonUserAgents() []string {
	return []string{
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36",
		"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:89.0) Gecko/20100101 Firefox/89.0",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:89.0) Gecko/20100101 Firefox/89.0",
		"Mozilla/5.0 (X11; Linux x86_64; rv:89.0) Gecko/20100101 Firefox/89.0",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/14.1.1 Safari/605.1.15",
		"goscan/1.0.1 (https://github.com/isa-programmer/goscan)",
	}
}

// IsValidURL checks if a URL is valid
func IsValidURL(targetURL string) bool {
	_, err := url.Parse(targetURL)
	return err == nil && (strings.HasPrefix(targetURL, "http://") || strings.HasPrefix(targetURL, "https://"))
}

// NormalizeURL normalizes a URL by ensuring it ends with a slash
func NormalizeURL(targetURL string) string {
	if !strings.HasSuffix(targetURL, "/") {
		return targetURL + "/"
	}
	return targetURL
}