package wordlist

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// WordlistManager handles wordlist operations
type WordlistManager struct {
	Words      []string
	FilePath   string
	Extensions []string
}

// New returns a new WordlistManager initialized with the specified file path and empty word and extension lists.
func New(filePath string) *WordlistManager {
	return &WordlistManager{
		Words:      make([]string, 0),
		FilePath:   filePath,
		Extensions: make([]string, 0),
	}
}

// LoadFromFile loads words from a file
func (wm *WordlistManager) LoadFromFile() error {
	file, err := os.Open(wm.FilePath)
	if err != nil {
		return fmt.Errorf("failed to open wordlist file: %v", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" && !strings.HasPrefix(line, "#") {
			wm.Words = append(wm.Words, line)
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading wordlist file: %v", err)
	}

	return nil
}

// LoadFromString loads words from a string (comma or newline separated)
func (wm *WordlistManager) LoadFromString(wordlist string) {
	lines := strings.Split(wordlist, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			// Check if line contains comma-separated values
			if strings.Contains(line, ",") {
				words := strings.Split(line, ",")
				for _, word := range words {
					word = strings.TrimSpace(word)
					if word != "" {
						wm.Words = append(wm.Words, word)
					}
				}
			} else {
				wm.Words = append(wm.Words, line)
			}
		}
	}
}

// AddWord adds a single word to the wordlist
func (wm *WordlistManager) AddWord(word string) {
	word = strings.TrimSpace(word)
	if word != "" {
		wm.Words = append(wm.Words, word)
	}
}

// AddWords adds multiple words to the wordlist
func (wm *WordlistManager) AddWords(words []string) {
	for _, word := range words {
		wm.AddWord(word)
	}
}

// SetExtensions sets file extensions to append to each word
func (wm *WordlistManager) SetExtensions(extensions []string) {
	wm.Extensions = extensions
}

// AddExtension adds a file extension
func (wm *WordlistManager) AddExtension(extension string) {
	if !strings.HasPrefix(extension, ".") {
		extension = "." + extension
	}
	wm.Extensions = append(wm.Extensions, extension)
}

// GetWords returns all words, optionally with extensions
func (wm *WordlistManager) GetWords() []string {
	if len(wm.Extensions) == 0 {
		return wm.Words
	}

	var result []string
	for _, word := range wm.Words {
		// Add the word without extension
		result = append(result, word)
		
		// Add the word with each extension
		for _, ext := range wm.Extensions {
			result = append(result, word+ext)
		}
	}

	return result
}

// GetWordsWithExtensions returns words only with extensions applied
func (wm *WordlistManager) GetWordsWithExtensions() []string {
	if len(wm.Extensions) == 0 {
		return wm.Words
	}

	var result []string
	for _, word := range wm.Words {
		for _, ext := range wm.Extensions {
			result = append(result, word+ext)
		}
	}

	return result
}

// RemoveDuplicates removes duplicate words from the wordlist
func (wm *WordlistManager) RemoveDuplicates() {
	seen := make(map[string]bool)
	var unique []string

	for _, word := range wm.Words {
		if !seen[word] {
			seen[word] = true
			unique = append(unique, word)
		}
	}

	wm.Words = unique
}

// Sort sorts the wordlist alphabetically
func (wm *WordlistManager) Sort() {
	sort.Strings(wm.Words)
}

// Filter filters words based on a condition function
func (wm *WordlistManager) Filter(condition func(string) bool) {
	var filtered []string
	for _, word := range wm.Words {
		if condition(word) {
			filtered = append(filtered, word)
		}
	}
	wm.Words = filtered
}

// Count returns the number of words in the wordlist
func (wm *WordlistManager) Count() int {
	return len(wm.Words)
}

// CountWithExtensions returns the total number of words including extensions
func (wm *WordlistManager) CountWithExtensions() int {
	if len(wm.Extensions) == 0 {
		return len(wm.Words)
	}
	return len(wm.Words) * (1 + len(wm.Extensions))
}

// SaveToFile saves the wordlist to a file
func (wm *WordlistManager) SaveToFile(filePath string) error {
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create file: %v", err)
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	for _, word := range wm.Words {
		_, err := writer.WriteString(word + "\n")
		if err != nil {
			return fmt.Errorf("failed to write word: %v", err)
		}
	}

	return writer.Flush()
}

// GetCommonExtensions returns a slice of commonly used file extensions for documents, scripts, archives, and configuration files.
func GetCommonExtensions() []string {
	return []string{
		".php", ".html", ".htm", ".asp", ".aspx", ".jsp", ".js",
		".css", ".txt", ".xml", ".json", ".pdf", ".doc", ".docx",
		".xls", ".xlsx", ".zip", ".rar", ".tar", ".gz", ".sql",
		".bak", ".old", ".tmp", ".log", ".conf", ".config", ".ini",
	}
}

// GetCommonDirectories returns a predefined list of commonly used directory names for web applications and file systems.
func GetCommonDirectories() []string {
	return []string{
		"admin", "administrator", "login", "panel", "dashboard",
		"wp-admin", "wp-content", "wp-includes", "uploads", "images",
		"css", "js", "javascript", "assets", "static", "public",
		"private", "backup", "backups", "old", "new", "test",
		"testing", "dev", "development", "staging", "prod",
		"production", "api", "v1", "v2", "docs", "documentation",
		"help", "support", "contact", "about", "home", "index",
	}
}

// MergeWordlists combines words from multiple wordlist files, removes duplicates, sorts them, and saves the result to the specified output file.
// Returns an error if any input file cannot be loaded or if saving fails.
func MergeWordlists(filePaths []string, outputPath string) error {
	merged := New("")

	for _, filePath := range filePaths {
		wm := New(filePath)
		err := wm.LoadFromFile()
		if err != nil {
			return fmt.Errorf("failed to load %s: %v", filePath, err)
		}
		merged.AddWords(wm.Words)
	}

	merged.RemoveDuplicates()
	merged.Sort()

	return merged.SaveToFile(outputPath)
}

// ValidateWordlistFile verifies that the provided wordlist file path is valid, exists, is a file (not a directory), and is not empty.
// Returns an error if any of these conditions are not met.
func ValidateWordlistFile(filePath string) error {
	if filePath == "" {
		return fmt.Errorf("wordlist file path is empty")
	}

	if !filepath.IsAbs(filePath) {
		abs, err := filepath.Abs(filePath)
		if err != nil {
			return fmt.Errorf("failed to get absolute path: %v", err)
		}
		filePath = abs
	}

	info, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("wordlist file does not exist: %s", filePath)
		}
		return fmt.Errorf("failed to access wordlist file: %v", err)
	}

	if info.IsDir() {
		return fmt.Errorf("wordlist path is a directory, not a file: %s", filePath)
	}

	if info.Size() == 0 {
		return fmt.Errorf("wordlist file is empty: %s", filePath)
	}

	return nil
}