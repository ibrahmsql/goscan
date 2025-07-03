package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/isa-programmer/goscan/modules/config"
	"github.com/isa-programmer/goscan/modules/logger"
	"github.com/isa-programmer/goscan/modules/scanner"
)

// printBanner prints the colored ASCII art banner for the goscan tool.
// If printHelp is true, it also displays usage instructions, version information, authorship, and example commands.
func printBanner(printHelp bool) {
	banner := `
	 ██████╗  ██████╗ ███████╗ ██████╗ █████╗ ███╗   ██║
	██╔════╝ ██╔═══██╗██╔════╝██╔════╝██╔══██╗████╗  ██║
	██║  ███╗██║   ██║███████╗██║     ███████║██╔██╗ ██║
	██║   ██║██║   ██║╚════██║██║     ██╔══██║██║╚██╗██║
	╚██████╔╝╚██████╔╝███████║╚██████╗██║  ██║██║ ╚████║
	 ╚═════╝  ╚═════╝ ╚══════╝ ╚═════╝╚═╝  ╚═╝╚═╝  ╚═══╝`

	fmt.Printf("\x1b[38;5;1m %s \x1b[0m \n", banner)
	if printHelp {
		fmt.Println("\t ⚡️ blazing fast directory scanner ⚡️ v1.0.1")
		fmt.Println("\t Made by isa-programmer & ibrahimsql")
		fmt.Println("\t Usage:")
		fmt.Println("\t\t goscan wordlist/wordlist.txt https://example.com/")
		fmt.Println("\t\t goscan wordlist/wordlist.txt https://example.com/ --no-warning # If you want ignore errors")
	}
}

// main is the entry point for the goscan command-line tool.
// It parses command-line arguments, initializes configuration and logging, runs the directory scan, and prints color-coded results and a summary to the console.
func main() {
	var warning bool = true
	if len(os.Args) < 3 {
		printBanner(true)
		return
	}

	if len(os.Args) >= 4 && os.Args[3] == "--no-warning" {
		warning = false
	}

	wordlistPath := os.Args[1]
	targetURL := os.Args[2]

	// Initialize configuration
	cfg := config.New()
	cfg.WordlistPath = wordlistPath
	cfg.TargetURL = targetURL
	cfg.Threads = 10
	cfg.Verbose = true
	cfg.Timeout = 10

	// Initialize logger
	log := logger.New(true)
	printBanner(false)
	log.Info("Starting goscan...")
	log.Info(fmt.Sprintf("Target: %s", cfg.TargetURL))
	log.Info(fmt.Sprintf("Wordlist: %s", cfg.WordlistPath))

	// Initialize and run scanner
	scanner := scanner.New(cfg, log)
	results, err := scanner.Run()
	if err != nil && warning {
		log.Error(fmt.Sprintf("Scanner error: %v", err))
	}

	// Display results with colors like original
	var success int = 0
	var failed int = 0
	for _, result := range results {
		if result.StatusCode != 0 {
			var color string
			if result.StatusCode >= 400 {
				color = "\x1b[38;5;1m"
			} else {
				color = "\x1b[38;5;2m"
			}
			// Extract path from URL
			path := strings.Replace(result.URL, cfg.TargetURL, "", 1)
			space := strings.Repeat(" ", 35-len(path))
			fmt.Printf("%s[+]\x1b[0m %s -> %s [%d] \n", color, path, space, result.StatusCode)
			success++
		} else {
			failed++
		}
	}

	fmt.Println("\x1b[38;5;2mSuccess:\x1b[0m", success)
	fmt.Println("\x1b[38;5;1mFailed:\x1b[0m", failed)
}