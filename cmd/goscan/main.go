package main

import (
	"fmt"
	"os"

	"github.com/isa-programmer/goscan/internal/app"
	"github.com/isa-programmer/goscan/internal/config"
	"github.com/isa-programmer/goscan/pkg/banner"
)

func main() {
	// Show banner
	banner.Show()

	// Parse configuration
	cfg, err := config.Parse()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Run the application
	app := app.New(cfg)
	if err := app.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}