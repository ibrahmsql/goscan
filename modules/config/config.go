package config

import (
	"fmt"
	"net/url"
	"os"
	"strings"
)

// Config holds all configuration options for goscan
type Config struct {
	WordlistPath string
	TargetURL    string
	Threads      int
	Verbose      bool
	Timeout      int
	OutputFile   string
	UserAgent    string
	Headers      map[string]string
	ProxyURL     string
	FollowRedirects bool
	MaxRedirects    int
	StatusCodes     []int
	Extensions      []string
	IgnoreSSL       bool
}

// New creates a new Config with default values
func New() *Config {
	return &Config{
		Threads:         10,
		Verbose:         false,
		Timeout:         10,
		UserAgent:       "goscan/1.0",
		Headers:         make(map[string]string),
		FollowRedirects: false,
		MaxRedirects:    3,
		StatusCodes:     []int{200, 301, 302, 403, 401},
		Extensions:      []string{"", ".php", ".html", ".txt", ".js", ".css"},
		IgnoreSSL:       false,
	}
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	// Validate URL
	if c.TargetURL == "" {
		return fmt.Errorf("target URL is required")
	}
	
	// Ensure URL has proper format
	if !strings.HasPrefix(c.TargetURL, "http://") && !strings.HasPrefix(c.TargetURL, "https://") {
		c.TargetURL = "http://" + c.TargetURL
	}
	
	// Parse and validate URL
	_, err := url.Parse(c.TargetURL)
	if err != nil {
		return fmt.Errorf("invalid target URL: %v", err)
	}
	
	// Validate WordlistPath exists and is readable
	if c.WordlistPath != "" {
		if _, err := os.Stat(c.WordlistPath); os.IsNotExist(err) {
			return fmt.Errorf("wordlist file does not exist: %s", c.WordlistPath)
		}
		if file, err := os.Open(c.WordlistPath); err != nil {
			return fmt.Errorf("wordlist file is not readable: %s", c.WordlistPath)
		} else {
			file.Close()
		}
	}
	
	// Validate ProxyURL format if provided
	if c.ProxyURL != "" {
		if _, err := url.Parse(c.ProxyURL); err != nil {
			return fmt.Errorf("invalid proxy URL format: %v", err)
		}
	}
	
	// Validate Extensions start with dot
	for i, ext := range c.Extensions {
		if ext != "" && !strings.HasPrefix(ext, ".") {
			return fmt.Errorf("extension at index %d must start with a dot: %s", i, ext)
		}
	}
	
	// Validate threads
	if c.Threads <= 0 {
		c.Threads = 10
	}
	if c.Threads > 100 {
		c.Threads = 100
	}
	
	// Validate timeout
	if c.Timeout <= 0 {
		c.Timeout = 10
	}
	
	return nil
}

// GetBaseURL returns the base URL without trailing slash
func (c *Config) GetBaseURL() string {
	return strings.TrimSuffix(c.TargetURL, "/")
}

// AddHeader adds a custom header
func (c *Config) AddHeader(key, value string) {
	c.Headers[key] = value
}

// SetStatusCodes sets the status codes to consider as valid
func (c *Config) SetStatusCodes(codes []int) {
	c.StatusCodes = codes
}

// SetExtensions sets the file extensions to append to wordlist entries
func (c *Config) SetExtensions(extensions []string) {
	c.Extensions = extensions
}