package utils

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"
)

// Colors for terminal output
const (
	ColorReset  = "\x1b[0m"
	ColorRed    = "\x1b[38;5;1m"
	ColorGreen  = "\x1b[38;5;2m"
	ColorYellow = "\x1b[38;5;3m"
	ColorBlue   = "\x1b[38;5;4m"
	ColorPurple = "\x1b[38;5;5m"
	ColorCyan   = "\x1b[38;5;6m"
	ColorWhite  = "\x1b[38;5;7m"
)

// ValidateURL validates if a URL is properly formatted
func ValidateURL(targetURL string) error {
	if targetURL == "" {
		return fmt.Errorf("URL cannot be empty")
	}

	parsedURL, err := url.Parse(targetURL)
	if err != nil {
		return fmt.Errorf("invalid URL format: %v", err)
	}

	if parsedURL.Scheme == "" {
		return fmt.Errorf("URL must include scheme (http:// or https://)")
	}

	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return fmt.Errorf("URL scheme must be http or https")
	}

	if parsedURL.Host == "" {
		return fmt.Errorf("URL must include a host")
	}

	return nil
}

// NormalizeURL ensures URL ends with a slash
func NormalizeURL(targetURL string) string {
	if !strings.HasSuffix(targetURL, "/") {
		return targetURL + "/"
	}
	return targetURL
}

// ExtractDomain extracts domain from URL
func ExtractDomain(targetURL string) (string, error) {
	parsedURL, err := url.Parse(targetURL)
	if err != nil {
		return "", err
	}
	return parsedURL.Host, nil
}

// FileExists checks if a file exists
func FileExists(filePath string) bool {
	_, err := os.Stat(filePath)
	return !os.IsNotExist(err)
}

// CreateDirIfNotExists creates a directory if it doesn't exist
func CreateDirIfNotExists(dirPath string) error {
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		return os.MkdirAll(dirPath, 0755)
	}
	return nil
}

// GetFileSize returns the size of a file in bytes
func GetFileSize(filePath string) (int64, error) {
	info, err := os.Stat(filePath)
	if err != nil {
		return 0, err
	}
	return info.Size(), nil
}

// FormatDuration formats a duration into a human-readable string
func FormatDuration(d time.Duration) string {
	if d < time.Second {
		return fmt.Sprintf("%.0fms", float64(d.Nanoseconds())/1e6)
	}
	if d < time.Minute {
		return fmt.Sprintf("%.1fs", d.Seconds())
	}
	if d < time.Hour {
		return fmt.Sprintf("%.1fm", d.Minutes())
	}
	return fmt.Sprintf("%.1fh", d.Hours())
}

// FormatBytes formats bytes into human readable format
func FormatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// GenerateRandomString generates a random string of specified length
func GenerateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		b[i] = charset[n.Int64()]
	}
	return string(b)
}

// IsValidStatusCode checks if a status code is valid
func IsValidStatusCode(code int) bool {
	return code >= 100 && code <= 599
}

// GetStatusCodeDescription returns a description for HTTP status codes
func GetStatusCodeDescription(code int) string {
	switch code {
	case 200:
		return "OK"
	case 201:
		return "Created"
	case 204:
		return "No Content"
	case 301:
		return "Moved Permanently"
	case 302:
		return "Found"
	case 304:
		return "Not Modified"
	case 400:
		return "Bad Request"
	case 401:
		return "Unauthorized"
	case 403:
		return "Forbidden"
	case 404:
		return "Not Found"
	case 405:
		return "Method Not Allowed"
	case 500:
		return "Internal Server Error"
	case 502:
		return "Bad Gateway"
	case 503:
		return "Service Unavailable"
	default:
		return "Unknown"
	}
}

// ParseRange parses a range string like "1-100" or "80,443,8080"
func ParseRange(rangeStr string) ([]int, error) {
	if rangeStr == "" {
		return []int{}, nil
	}

	var result []int
	parts := strings.Split(rangeStr, ",")

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		// Handle ranges like "1-100"
		if strings.Contains(part, "-") {
			rangeParts := strings.Split(part, "-")
			if len(rangeParts) != 2 {
				return nil, fmt.Errorf("invalid range format: %s", part)
			}

			start, err := strconv.Atoi(strings.TrimSpace(rangeParts[0]))
			if err != nil {
				return nil, fmt.Errorf("invalid start number in range: %s", part)
			}

			end, err := strconv.Atoi(strings.TrimSpace(rangeParts[1]))
			if err != nil {
				return nil, fmt.Errorf("invalid end number in range: %s", part)
			}

			if start > end {
				return nil, fmt.Errorf("start number cannot be greater than end number: %s", part)
			}

			for i := start; i <= end; i++ {
				result = append(result, i)
			}
		} else {
			// Handle single numbers
			num, err := strconv.Atoi(part)
			if err != nil {
				return nil, fmt.Errorf("invalid number: %s", part)
			}
			result = append(result, num)
		}
	}

	return result, nil
}

// RemoveDuplicateInts removes duplicate integers from a slice
func RemoveDuplicateInts(slice []int) []int {
	seen := make(map[int]bool)
	var result []int

	for _, item := range slice {
		if !seen[item] {
			seen[item] = true
			result = append(result, item)
		}
	}

	return result
}

// RemoveDuplicateStrings removes duplicate strings from a slice
func RemoveDuplicateStrings(slice []string) []string {
	seen := make(map[string]bool)
	var result []string

	for _, item := range slice {
		if !seen[item] {
			seen[item] = true
			result = append(result, item)
		}
	}

	return result
}

// IsValidRegex checks if a string is a valid regular expression
func IsValidRegex(pattern string) bool {
	_, err := regexp.Compile(pattern)
	return err == nil
}

// GetSystemInfo returns basic system information
func GetSystemInfo() map[string]string {
	return map[string]string{
		"os":           runtime.GOOS,
		"arch":         runtime.GOARCH,
		"go_version":   runtime.Version(),
		"num_cpu":      strconv.Itoa(runtime.NumCPU()),
		"num_goroutine": strconv.Itoa(runtime.NumGoroutine()),
	}
}

// GetWorkingDirectory returns the current working directory
func GetWorkingDirectory() (string, error) {
	return os.Getwd()
}

// GetAbsolutePath returns the absolute path of a file
func GetAbsolutePath(path string) (string, error) {
	return filepath.Abs(path)
}

// TruncateString truncates a string to a maximum length
func TruncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// PadString pads a string to a specific length
func PadString(s string, length int, padChar rune) string {
	if len(s) >= length {
		return s
	}
	padding := strings.Repeat(string(padChar), length-len(s))
	return s + padding
}

// ColorizeStatusCode returns a colorized status code string
func ColorizeStatusCode(code int) string {
	var color string
	switch {
	case code >= 200 && code < 300:
		color = ColorGreen
	case code >= 300 && code < 400:
		color = ColorYellow
	case code >= 400 && code < 500:
		color = ColorRed
	case code >= 500:
		color = ColorPurple
	default:
		color = ColorWhite
	}
	return fmt.Sprintf("%s%d%s", color, code, ColorReset)
}

// ColorizeText returns colorized text
func ColorizeText(text, color string) string {
	return fmt.Sprintf("%s%s%s", color, text, ColorReset)
}

// ProgressBar creates a simple progress bar
func ProgressBar(current, total int, width int) string {
	if total == 0 {
		return ""
	}

	percent := float64(current) / float64(total)
	filled := int(percent * float64(width))

	bar := strings.Repeat("█", filled) + strings.Repeat("░", width-filled)
	return fmt.Sprintf("[%s] %.1f%% (%d/%d)", bar, percent*100, current, total)
}

// GetEnvOrDefault gets an environment variable or returns a default value
func GetEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// IsPortValid checks if a port number is valid
func IsPortValid(port int) bool {
	return port > 0 && port <= 65535
}

// SanitizeFilename removes invalid characters from a filename
func SanitizeFilename(filename string) string {
	// Remove or replace invalid characters
	invalidChars := regexp.MustCompile(`[<>:"/\\|?*]`)
	sanitized := invalidChars.ReplaceAllString(filename, "_")
	
	// Remove leading/trailing spaces and dots
	sanitized = strings.Trim(sanitized, " .")
	
	// Ensure it's not empty
	if sanitized == "" {
		sanitized = "unnamed"
	}
	
	return sanitized
}