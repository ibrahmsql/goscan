package filter

import (
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/isa-programmer/goscan/modules/scanner"
)

// FilterConfig holds filtering configuration
type FilterConfig struct {
	StatusCodes    []int
	ExcludeStatus  []int
	MinSize        int64
	MaxSize        int64
	MinTime        time.Duration
	MaxTime        time.Duration
	ContentTypes   []string
	ExcludeContent []string
	RegexPatterns  []*regexp.Regexp
	ExcludeRegex   []*regexp.Regexp
	Words          []string
	ExcludeWords   []string
}

// ResultFilter handles filtering of scan results
type ResultFilter struct {
	Config *FilterConfig
}

// New returns a ResultFilter initialized with default filtering criteria for HTTP scan results, including common status codes to include and exclude, and no limits on size or time.
func New() *ResultFilter {
	return &ResultFilter{
		Config: &FilterConfig{
			StatusCodes:    []int{200, 201, 202, 204, 301, 302, 307, 308, 401, 403, 500},
			ExcludeStatus:  []int{404, 400},
			MinSize:        0,
			MaxSize:        0, // 0 means no limit
			MinTime:        0,
			MaxTime:        0, // 0 means no limit
			ContentTypes:   []string{},
			ExcludeContent: []string{},
			RegexPatterns:  []*regexp.Regexp{},
			ExcludeRegex:   []*regexp.Regexp{},
			Words:          []string{},
			ExcludeWords:   []string{},
		},
	}
}

// SetStatusCodes sets the status codes to include
func (rf *ResultFilter) SetStatusCodes(codes []int) {
	rf.Config.StatusCodes = codes
}

// AddStatusCode adds a status code to include
func (rf *ResultFilter) AddStatusCode(code int) {
	rf.Config.StatusCodes = append(rf.Config.StatusCodes, code)
}

// SetExcludeStatus sets the status codes to exclude
func (rf *ResultFilter) SetExcludeStatus(codes []int) {
	rf.Config.ExcludeStatus = codes
}

// AddExcludeStatus adds a status code to exclude
func (rf *ResultFilter) AddExcludeStatus(code int) {
	rf.Config.ExcludeStatus = append(rf.Config.ExcludeStatus, code)
}

// SetSizeRange sets the response size range filter
func (rf *ResultFilter) SetSizeRange(minSize, maxSize int64) {
	rf.Config.MinSize = minSize
	rf.Config.MaxSize = maxSize
}

// SetTimeRange sets the response time range filter
func (rf *ResultFilter) SetTimeRange(minTime, maxTime time.Duration) {
	rf.Config.MinTime = minTime
	rf.Config.MaxTime = maxTime
}

// AddContentType adds a content type to include
func (rf *ResultFilter) AddContentType(contentType string) {
	rf.Config.ContentTypes = append(rf.Config.ContentTypes, contentType)
}

// AddExcludeContentType adds a content type to exclude
func (rf *ResultFilter) AddExcludeContentType(contentType string) {
	rf.Config.ExcludeContent = append(rf.Config.ExcludeContent, contentType)
}

// AddRegexPattern adds a regex pattern to match response content
func (rf *ResultFilter) AddRegexPattern(pattern string) error {
	regex, err := regexp.Compile(pattern)
	if err != nil {
		return err
	}
	rf.Config.RegexPatterns = append(rf.Config.RegexPatterns, regex)
	return nil
}

// AddExcludeRegexPattern adds a regex pattern to exclude from response content
func (rf *ResultFilter) AddExcludeRegexPattern(pattern string) error {
	regex, err := regexp.Compile(pattern)
	if err != nil {
		return err
	}
	rf.Config.ExcludeRegex = append(rf.Config.ExcludeRegex, regex)
	return nil
}

// AddWord adds a word to search for in response content
func (rf *ResultFilter) AddWord(word string) {
	rf.Config.Words = append(rf.Config.Words, word)
}

// AddExcludeWord adds a word to exclude from response content
func (rf *ResultFilter) AddExcludeWord(word string) {
	rf.Config.ExcludeWords = append(rf.Config.ExcludeWords, word)
}

// ShouldInclude determines if a result should be included based on filters
func (rf *ResultFilter) ShouldInclude(result scanner.Result, response *http.Response, body string) bool {
	// Check status code inclusion
	if len(rf.Config.StatusCodes) > 0 {
		included := false
		for _, code := range rf.Config.StatusCodes {
			if result.StatusCode == code {
				included = true
				break
			}
		}
		if !included {
			return false
		}
	}

	// Check status code exclusion
	for _, code := range rf.Config.ExcludeStatus {
		if result.StatusCode == code {
			return false
		}
	}

	// Check response size
	if response != nil {
		contentLength := response.ContentLength
		if contentLength == -1 && body != "" {
			contentLength = int64(len(body))
		}

		if rf.Config.MinSize > 0 && contentLength < rf.Config.MinSize {
			return false
		}
		if rf.Config.MaxSize > 0 && contentLength > rf.Config.MaxSize {
			return false
		}
	}

	// Check response time
	if result.ResponseTime > 0 {
		if rf.Config.MinTime > 0 && result.ResponseTime < rf.Config.MinTime {
			return false
		}
		if rf.Config.MaxTime > 0 && result.ResponseTime > rf.Config.MaxTime {
			return false
		}
	}

	// Check content type
	if response != nil && len(rf.Config.ContentTypes) > 0 {
		contentType := response.Header.Get("Content-Type")
		included := false
		for _, ct := range rf.Config.ContentTypes {
			if strings.Contains(contentType, ct) {
				included = true
				break
			}
		}
		if !included {
			return false
		}
	}

	// Check excluded content types
	if response != nil && len(rf.Config.ExcludeContent) > 0 {
		contentType := response.Header.Get("Content-Type")
		for _, ct := range rf.Config.ExcludeContent {
			if strings.Contains(contentType, ct) {
				return false
			}
		}
	}

	// Check regex patterns (must match)
	if len(rf.Config.RegexPatterns) > 0 && body != "" {
		matched := false
		for _, regex := range rf.Config.RegexPatterns {
			if regex.MatchString(body) {
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}

	// Check excluded regex patterns (must not match)
	if len(rf.Config.ExcludeRegex) > 0 && body != "" {
		for _, regex := range rf.Config.ExcludeRegex {
			if regex.MatchString(body) {
				return false
			}
		}
	}

	// Check words (must contain)
	if len(rf.Config.Words) > 0 && body != "" {
		matched := false
		for _, word := range rf.Config.Words {
			if strings.Contains(strings.ToLower(body), strings.ToLower(word)) {
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}

	// Check excluded words (must not contain)
	if len(rf.Config.ExcludeWords) > 0 && body != "" {
		for _, word := range rf.Config.ExcludeWords {
			if strings.Contains(strings.ToLower(body), strings.ToLower(word)) {
				return false
			}
		}
	}

	return true
}

// FilterResults filters a slice of results based on the configured filters
func (rf *ResultFilter) FilterResults(results []scanner.Result) []scanner.Result {
	var filtered []scanner.Result
	for _, result := range results {
		if rf.ShouldInclude(result, nil, "") {
			filtered = append(filtered, result)
		}
	}
	return filtered
}

// ParseStatusCodes converts a comma-separated string of HTTP status codes and ranges (e.g., "200,201,300-305") into a slice of integers.
// Returns an error if any code or range is invalid.
func ParseStatusCodes(codes string) ([]int, error) {
	if codes == "" {
		return []int{}, nil
	}

	var result []int
	parts := strings.Split(codes, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		// Handle ranges like "200-299"
		if strings.Contains(part, "-") {
			rangeParts := strings.Split(part, "-")
			if len(rangeParts) == 2 {
				start, err1 := strconv.Atoi(strings.TrimSpace(rangeParts[0]))
				end, err2 := strconv.Atoi(strings.TrimSpace(rangeParts[1]))
				if err1 == nil && err2 == nil && start <= end {
					for i := start; i <= end; i++ {
						result = append(result, i)
					}
					continue
				}
			}
		}

		// Handle single status code
		code, err := strconv.Atoi(part)
		if err != nil {
			return nil, err
		}
		result = append(result, code)
	}

	return result, nil
}

// GetInterestingStatusCodes returns a slice of HTTP status codes that are typically considered noteworthy, including successful responses, redirects, authentication errors, and server errors.
func GetInterestingStatusCodes() []int {
	return []int{
		200, 201, 202, 204, // Success
		301, 302, 307, 308, // Redirects
		401, 403, // Authentication/Authorization
		500, 502, 503, // Server errors
	}
}

// GetCommonExcludeStatusCodes returns a slice of HTTP status codes that are commonly excluded from scan results, such as 404 (Not Found), 400 (Bad Request), and similar error codes.
func GetCommonExcludeStatusCodes() []int {
	return []int{
		404, 400, 405, 406, 408, 410, 414, 429,
	}
}