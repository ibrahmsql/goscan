package app

import (
	"fmt"
	"strings"
	"time"

	"github.com/isa-programmer/goscan/internal/config"
	"github.com/isa-programmer/goscan/internal/scanner"
	"github.com/isa-programmer/goscan/internal/wordlist"
	"github.com/isa-programmer/goscan/pkg/output"
	"github.com/isa-programmer/goscan/pkg/stats"
)

// App represents the main application
type App struct {
	config  *config.Config
	scanner *scanner.Scanner
	stats   *stats.Stats
	output  *output.Manager
}

// New creates a new application instance
func New(cfg *config.Config) *App {
	return &App{
		config: cfg,
		stats:  stats.New(),
		output: output.New(cfg),
	}
}

// Run executes the main application logic
func (a *App) Run() error {
	// Initialize scanner
	var err error
	a.scanner, err = scanner.New(a.config)
	if err != nil {
		return fmt.Errorf("failed to initialize scanner: %w", err)
	}

	// Load wordlist
	wordlistManager := wordlist.New(a.config)
	words, err := wordlistManager.Load()
	if err != nil {
		return fmt.Errorf("failed to load wordlist: %w", err)
	}

	// Generate URLs to test
	urls := a.generateURLs(words)

	// Show scan information
	if !a.config.Quiet {
		a.showScanInfo(len(words), len(urls))
	}

	// Initialize stats
	a.stats.SetTotal(int64(len(urls)))
	a.stats.Start()

	// Start progress reporting
	var done chan bool
	if !a.config.Quiet && !a.config.NoProgress {
		done = make(chan bool)
		go a.showProgress(done)
	}

	// Perform scanning
	results, err := a.scanner.Scan(urls, a.stats)
	if err != nil {
		return fmt.Errorf("scanning failed: %w", err)
	}

	// Stop progress reporting
	if !a.config.Quiet && !a.config.NoProgress {
		a.stats.Stop()
		close(done)
	}

	// Output results
	if err := a.output.WriteResults(results); err != nil {
		return fmt.Errorf("failed to write results: %w", err)
	}

	// Show summary
	if !a.config.Quiet {
		a.showSummary(len(urls), len(results))
	}

	return nil
}

// generateURLs generates URLs to test from wordlist
func (a *App) generateURLs(words []string) []string {
	var urls []string

	for _, word := range words {
		// Apply word transformations
		transformedWords := a.transformWord(word)

		for _, transformedWord := range transformedWords {
			// Add base word
			urls = append(urls, a.config.Target+transformedWord)

			// Add with extensions
			for _, ext := range a.config.Extensions {
				if !strings.HasPrefix(ext, ".") {
					ext = "." + ext
				}
				urls = append(urls, a.config.Target+transformedWord+ext)
			}

			// Add with slash for directories
			if a.config.AppendSlash {
				urls = append(urls, a.config.Target+transformedWord+"/")
			}
		}
	}

	return urls
}

// transformWord applies various transformations to a word
func (a *App) transformWord(word string) []string {
	words := []string{word}

	if a.config.Lowercase {
		words = append(words, strings.ToLower(word))
	}

	if a.config.Uppercase {
		words = append(words, strings.ToUpper(word))
	}

	if a.config.Capitalize {
		words = append(words, strings.Title(strings.ToLower(word)))
	}

	// Remove duplicates
	seen := make(map[string]bool)
	var unique []string
	for _, w := range words {
		if !seen[w] {
			seen[w] = true
			unique = append(unique, w)
		}
	}

	return unique
}

// showScanInfo displays scan information
func (a *App) showScanInfo(wordCount, urlCount int) {
	fmt.Printf("Target: %s\n", a.config.Target)
	fmt.Printf("Wordlist: %s (%d words)\n", a.config.Wordlist, wordCount)
	fmt.Printf("URLs to test: %d\n", urlCount)
	fmt.Printf("Threads: %d\n", a.config.Threads)
	fmt.Printf("Timeout: %v\n", a.config.Timeout)
	if a.config.Method != "GET" {
		fmt.Printf("Method: %s\n", a.config.Method)
	}
	if len(a.config.Extensions) > 0 {
		fmt.Printf("Extensions: %v\n", a.config.Extensions)
	}
	if len(a.config.StatusCodes) > 0 {
		fmt.Printf("Status codes: %v\n", a.config.StatusCodes)
	}
	fmt.Println("=")
}

// showProgress displays real-time progress
func (a *App) showProgress(done <-chan bool) {
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			a.stats.PrintProgress()
		case <-done:
			return
		}
	}
}

// showSummary displays scan summary
func (a *App) showSummary(totalURLs, foundURLs int) {
	elapsed := a.stats.GetElapsed()
	fmt.Printf("\n=== SCAN COMPLETE ===\n")
	fmt.Printf("Total URLs: %d\n", totalURLs)
	fmt.Printf("Found: %d\n", foundURLs)
	fmt.Printf("Errors: %d\n", a.stats.GetErrors())
	fmt.Printf("Time: %v\n", elapsed.Round(time.Second))
	if elapsed.Seconds() > 0 {
		fmt.Printf("Requests/sec: %.2f\n", float64(totalURLs)/elapsed.Seconds())
	}
}