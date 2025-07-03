package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/isa-programmer/goscan/modules/apiscanner"
	"github.com/isa-programmer/goscan/modules/config"
	"github.com/isa-programmer/goscan/modules/logger"
)

// max returns the maximum of two integers
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func printBanner() {
	banner := `
	 █████╗ ██████╗ ██╗███████╗ ██████╗ █████╗ ███╗   ██║
	██╔══██╗██╔══██╗██║██╔════╝██╔════╝██╔══██╗████╗  ██║
	███████║██████╔╝██║███████╗██║     ███████║██╔██╗ ██║
	██╔══██║██╔═══╝ ██║╚════██║██║     ██╔══██║██║╚██╗██║
	██║  ██║██║     ██║███████║╚██████╗██║  ██║██║ ╚████║
	╚═╝  ╚═╝╚═╝     ╚═╝╚══════╝ ╚═════╝╚═╝  ╚═╝╚═╝  ╚═══╝`

	fmt.Printf("\x1b[38;5;3m %s \x1b[0m \n", banner)
	fmt.Println("\t ⚡️ blazing fast API endpoint scanner ⚡️ v1.0.0")
	fmt.Println("\t Made by isa-programmer & ibrahimsql")
}

func main() {
	// Define flags with default values and descriptions
	var (
		outputFile = flag.String("output", "", "Output file for results (JSON format)")
		threads    = flag.Int("threads", 10, "Number of concurrent threads")
		timeout    = flag.Int("timeout", 10, "Request timeout in seconds")
		quiet      = flag.Bool("quiet", false, "Suppress verbose output")
		help       = flag.Bool("help", false, "Show help message")
	)

	// Custom usage function
	flag.Usage = func() {
		printBanner()
		fmt.Fprintf(os.Stderr, "\nUsage: %s [OPTIONS] <wordlist> <target_url>\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Arguments:\n")
		fmt.Fprintf(os.Stderr, "  wordlist    Path to wordlist file containing API endpoints\n")
		fmt.Fprintf(os.Stderr, "  target_url  Target URL to scan (e.g., https://api.example.com)\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  %s wordlists/api-endpoints.txt https://api.example.com\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -output results.json -threads 20 wordlists/api-endpoints.txt https://api.example.com\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -timeout 15 -quiet wordlists/api-endpoints.txt https://api.example.com\n", os.Args[0])
	}

	// Parse flags
	flag.Parse()

	// Show help if requested
	if *help {
		flag.Usage()
		return
	}

	// Validate required positional arguments
	args := flag.Args()
	if len(args) < 2 {
		fmt.Fprintf(os.Stderr, "Error: Missing required arguments\n\n")
		flag.Usage()
		os.Exit(1)
	}

	wordlistPath := args[0]
	targetURL := args[1]

	// Validate arguments
	if _, err := os.Stat(wordlistPath); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "Error: Wordlist file '%s' does not exist\n", wordlistPath)
		os.Exit(1)
	}

	if !strings.HasPrefix(targetURL, "http://") && !strings.HasPrefix(targetURL, "https://") {
		fmt.Fprintf(os.Stderr, "Error: Target URL must start with http:// or https://\n")
		os.Exit(1)
	}

	if *threads <= 0 {
		fmt.Fprintf(os.Stderr, "Error: Threads must be a positive number\n")
		os.Exit(1)
	}

	if *timeout <= 0 {
		fmt.Fprintf(os.Stderr, "Error: Timeout must be a positive number\n")
		os.Exit(1)
	}

	verbose := !*quiet

	// Initialize configuration
	cfg := config.New()
	cfg.WordlistPath = wordlistPath
	cfg.TargetURL = targetURL
	cfg.Threads = *threads
	cfg.Verbose = verbose
	cfg.Timeout = *timeout
	cfg.StatusCodes = []int{200, 201, 202, 204, 301, 302, 400, 401, 403, 404, 405, 500, 501, 502, 503}

	// Initialize logger
	log := logger.New(verbose)
	printBanner()
	log.Info("Starting API scan...")
	log.Info(fmt.Sprintf("Target: %s", cfg.TargetURL))
	log.Info(fmt.Sprintf("Wordlist: %s", cfg.WordlistPath))
	log.Info(fmt.Sprintf("Threads: %d", cfg.Threads))

	// Initialize and run API scanner
	apiScanner := apiscanner.New(cfg, log)
	results, err := apiScanner.ScanAPI()
	if err != nil {
		log.Error(fmt.Sprintf("API scanner error: %v", err))
		os.Exit(1)
	}

	// Display results
	var found int = 0
	var methodCounts = make(map[string]int)
	var statusCounts = make(map[int]int)

	for _, result := range results {
		var color string
		switch {
		case result.StatusCode >= 200 && result.StatusCode < 300:
			color = "\x1b[38;5;2m" // Green
		case result.StatusCode >= 300 && result.StatusCode < 400:
			color = "\x1b[38;5;3m" // Yellow
		case result.StatusCode >= 400 && result.StatusCode < 500:
			color = "\x1b[38;5;1m" // Red
		default:
			color = "\x1b[38;5;5m" // Purple
		}

		// Extract path from URL
		path := strings.Replace(result.URL, cfg.TargetURL, "", 1)
		space := strings.Repeat(" ", max(0, 40-len(path)))
		methodSpace := strings.Repeat(" ", max(0, 8-len(result.Method)))

		fmt.Printf("%s[+]\x1b[0m %s%s -> %s%s [%d] %s\n", 
			color, result.Method, methodSpace, path, space, result.StatusCode, result.ContentType)

		found++
		methodCounts[result.Method]++
		statusCounts[result.StatusCode]++
	}

	// Print statistics
	fmt.Println("\n\x1b[38;5;6m=== API Scan Results ===\x1b[0m")
	fmt.Printf("\x1b[38;5;2mTotal Found:\x1b[0m %d\n", found)

	fmt.Println("\n\x1b[38;5;4mBy HTTP Method:\x1b[0m")
	for method, count := range methodCounts {
		fmt.Printf("  %s: %d\n", method, count)
	}

	fmt.Println("\n\x1b[38;5;4mBy Status Code:\x1b[0m")
	for status, count := range statusCounts {
		fmt.Printf("  %d: %d\n", status, count)
	}

	// Export to JSON if requested
	if *outputFile != "" {
		if err := apiScanner.ExportResults(*outputFile); err != nil {
			log.Error(fmt.Sprintf("Failed to export results: %v", err))
		} else {
			log.Success(fmt.Sprintf("Results exported to %s", *outputFile))
		}
	}

	// Also save a summary JSON
	summary := map[string]interface{}{
		"target":        cfg.TargetURL,
		"total_found":   found,
		"method_counts": methodCounts,
		"status_counts": statusCounts,
		"results":       results,
	}

	// Marshal summary to JSON with proper error handling
	summaryData, marshalErr := json.MarshalIndent(summary, "", "  ")
	if marshalErr != nil {
		log.Error(fmt.Sprintf("Failed to marshal summary to JSON: %v", marshalErr))
	} else {
		// Write JSON data to file with proper error handling
		if writeErr := os.WriteFile("apiscan_summary.json", summaryData, 0644); writeErr != nil {
			log.Error(fmt.Sprintf("Failed to write summary file: %v", writeErr))
		} else {
			log.Info("Summary saved to apiscan_summary.json")
		}
	}
}