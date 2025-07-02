package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/isa-programmer/goscan/modules/apiscanner"
	"github.com/isa-programmer/goscan/modules/config"
	"github.com/isa-programmer/goscan/modules/logger"
)

func printBanner(printHelp bool) {
	banner := `
	 █████╗ ██████╗ ██╗███████╗ ██████╗ █████╗ ███╗   ██╗
	██╔══██╗██╔══██╗██║██╔════╝██╔════╝██╔══██╗████╗  ██║
	███████║██████╔╝██║███████╗██║     ███████║██╔██╗ ██║
	██╔══██║██╔═══╝ ██║╚════██║██║     ██╔══██║██║╚██╗██║
	██║  ██║██║     ██║███████║╚██████╗██║  ██║██║ ╚████║
	╚═╝  ╚═╝╚═╝     ╚═╝╚══════╝ ╚═════╝╚═╝  ╚═╝╚═╝  ╚═══╝`

	fmt.Printf("\x1b[38;5;3m %s \x1b[0m \n", banner)
	if printHelp {
		fmt.Println("\t ⚡️ blazing fast API endpoint scanner ⚡️ v1.0.0")
		fmt.Println("\t Made by isa-programmer & ibrahimsql")
		fmt.Println("\t Usage:")
		fmt.Println("\t\t apiscan wordlists/api-endpoints.txt https://api.example.com")
		fmt.Println("\t\t apiscan wordlists/api-endpoints.txt https://api.example.com --output results.json")
		fmt.Println("\t\t apiscan wordlists/api-endpoints.txt https://api.example.com --threads 20 --timeout 15")
	}
}

func main() {
	if len(os.Args) < 3 {
		printBanner(true)
		return
	}

	wordlistPath := os.Args[1]
	targetURL := os.Args[2]

	// Parse additional arguments
	var outputFile string
	threads := 10
	timeout := 10
	verbose := true

	for i := 3; i < len(os.Args); i++ {
		switch os.Args[i] {
		case "--output", "-o":
			if i+1 < len(os.Args) {
				outputFile = os.Args[i+1]
				i++
			}
		case "--threads", "-t":
			if i+1 < len(os.Args) {
				fmt.Sscanf(os.Args[i+1], "%d", &threads)
				i++
			}
		case "--timeout":
			if i+1 < len(os.Args) {
				fmt.Sscanf(os.Args[i+1], "%d", &timeout)
				i++
			}
		case "--quiet", "-q":
			verbose = false
		}
	}

	// Initialize configuration
	cfg := config.New()
	cfg.WordlistPath = wordlistPath
	cfg.TargetURL = targetURL
	cfg.Threads = threads
	cfg.Verbose = verbose
	cfg.Timeout = timeout
	cfg.StatusCodes = []int{200, 201, 202, 204, 301, 302, 400, 401, 403, 404, 405, 500, 501, 502, 503}

	// Initialize logger
	log := logger.New(verbose)
	printBanner(false)
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
		space := strings.Repeat(" ", 40-len(path))
		methodSpace := strings.Repeat(" ", 8-len(result.Method))

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
	if outputFile != "" {
		if err := apiScanner.ExportResults(outputFile); err != nil {
			log.Error(fmt.Sprintf("Failed to export results: %v", err))
		} else {
			log.Success(fmt.Sprintf("Results exported to %s", outputFile))
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

	if summaryData, err := json.MarshalIndent(summary, "", "  "); err == nil {
		if err := os.WriteFile("apiscan_summary.json", summaryData, 0644); err == nil {
			log.Info("Summary saved to apiscan_summary.json")
		}
	}
}