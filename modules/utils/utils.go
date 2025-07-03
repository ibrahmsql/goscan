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

// ValidateURL checks whether the provided URL string is non-empty, properly formatted, uses the http or https scheme, and includes a host.
// Returns an error if the URL is invalid.
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

// NormalizeURL returns the input URL string with a trailing slash, adding one if it is missing.
func NormalizeURL(targetURL string) string {
	if !strings.HasSuffix(targetURL, "/") {
		return targetURL + "/"
	}
	return targetURL
}

// ExtractDomain returns the host or domain part of the given URL string.
// Returns an error if the URL cannot be parsed.
func ExtractDomain(targetURL string) (string, error) {
	parsedURL, err := url.Parse(targetURL)
	if err != nil {
		return "", err
	}
	return parsedURL.Host, nil
}

// FileExists returns true if a file exists at the specified path, or false otherwise.
func FileExists(filePath string) bool {
	_, err := os.Stat(filePath)
	return !os.IsNotExist(err)
}

// CreateDirIfNotExists creates the specified directory and all necessary parent directories if they do not already exist.
// Returns an error if the directory cannot be created.
func CreateDirIfNotExists(dirPath string) error {
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		return os.MkdirAll(dirPath, 0755)
	}
	return nil
}

// GetFileSize returns the size of the specified file in bytes.
// It returns an error if the file does not exist or cannot be accessed.
func GetFileSize(filePath string) (int64, error) {
	info, err := os.Stat(filePath)
	if err != nil {
		return 0, err
	}
	return info.Size(), nil
}

// FormatDuration returns a human-readable string representation of a time.Duration using appropriate units (milliseconds, seconds, minutes, or hours).
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

// FormatBytes converts a byte count into a human-readable string with appropriate units (B, KB, MB, GB, etc.).
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

// GenerateRandomString returns a cryptographically secure random alphanumeric string of the specified length.
func GenerateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		b[i] = charset[n.Int64()]
	}
	return string(b)
}

// IsValidStatusCode returns true if the given integer is a valid HTTP status code (100–599).
func IsValidStatusCode(code int) bool {
	return code >= 100 && code <= 599
}

// GetStatusCodeDescription returns a short description for common HTTP status codes, or "Unknown" if the code is not recognized.
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

// ParseRange converts a string representing numbers and numeric ranges (e.g., "1-3,5,7-9") into a slice of integers.
// Returns an error if the input contains invalid formats or ranges.
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

// RemoveDuplicateInts returns a new slice containing the unique integers from the input slice, preserving their original order.
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

// RemoveDuplicateStrings returns a new slice with duplicate strings removed, preserving the order of the first occurrence of each string.
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

// IsValidRegex returns true if the provided string is a valid regular expression pattern.
func IsValidRegex(pattern string) bool {
	_, err := regexp.Compile(pattern)
	return err == nil
}

// GetSystemInfo returns a map containing basic system information such as operating system, architecture, Go version, number of CPUs, and number of goroutines.
func GetSystemInfo() map[string]string {
	return map[string]string{
		"os":           runtime.GOOS,
		"arch":         runtime.GOARCH,
		"go_version":   runtime.Version(),
		"num_cpu":      strconv.Itoa(runtime.NumCPU()),
		"num_goroutine": strconv.Itoa(runtime.NumGoroutine()),
	}
}

// GetWorkingDirectory returns the current working directory path.
// It returns an error if the working directory cannot be determined.
func GetWorkingDirectory() (string, error) {
	return os.Getwd()
}

// GetAbsolutePath returns the absolute path for the given file or directory path.
// If the input is already absolute, it is returned unchanged. Returns an error if the path cannot be resolved.
func GetAbsolutePath(path string) (string, error) {
	return filepath.Abs(path)
}

// TruncateString shortens a string to a specified maximum length, appending "..." if truncation occurs.
func TruncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// PadString returns the input string padded on the right with the specified character until it reaches the desired length. If the string is already at least the specified length, it is returned unchanged.
func PadString(s string, length int, padChar rune) string {
	if len(s) >= length {
		return s
	}
	padding := strings.Repeat(string(padChar), length-len(s))
	return s + padding
}

// ColorizeStatusCode returns the HTTP status code as a string wrapped in terminal color codes based on its range. 2xx codes are green, 3xx yellow, 4xx red, 5xx purple, and others white.
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

// ColorizeText wraps the given text with terminal color codes for the specified color and resets formatting at the end.
func ColorizeText(text, color string) string {
	return fmt.Sprintf("%s%s%s", color, text, ColorReset)
}

// ProgressBar returns a textual progress bar representing the completion percentage based on the current and total values, with a configurable width.
func ProgressBar(current, total int, width int) string {
	if total == 0 {
		return ""
	}

	percent := float64(current) / float64(total)
	filled := int(percent * float64(width))

	bar := strings.Repeat("█", filled) + strings.Repeat("░", width-filled)
	return fmt.Sprintf("[%s] %.1f%% (%d/%d)", bar, percent*100, current, total)
}

// GetEnvOrDefault returns the value of the specified environment variable, or the provided default value if the variable is not set or is empty.
func GetEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// IsPortValid returns true if the given port number is within the valid TCP/UDP port range (1–65535).
func IsPortValid(port int) bool {
	return port > 0 && port <= 65535
}

// SanitizeFilename returns a safe filename by replacing invalid characters with underscores, trimming leading and trailing spaces or dots, and defaulting to "unnamed" if the result is empty.
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