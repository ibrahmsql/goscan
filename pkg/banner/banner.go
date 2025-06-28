package banner

import (
	"fmt"
	"runtime"
)

const (
	version = "v2.1.0"
	author  = "isa-programmer" && "ibrahimsql"
)

// Show displays the application banner
func Show() {
	fmt.Printf(`
   ██████╗  ██████╗ ███████╗ ██████╗ █████╗ ███╗   ██╗
  ██╔════╝ ██╔═══██╗██╔════╝██╔════╝██╔══██╗████╗  ██║
  ██║  ███╗██║   ██║███████╗██║     ███████║██╔██╗ ██║
  ██║   ██║██║   ██║╚════██║██║     ██╔══██║██║╚██╗██║
  ╚██████╔╝╚██████╔╝███████║╚██████╗██║  ██║██║ ╚████║
   ╚═════╝  ╚═════╝ ╚══════╝ ╚═════╝╚═╝  ╚═╝╚═╝  ╚═══╝

`)
	fmt.Printf("  🚀 Advanced Directory & File Scanner %s\n", version)
	fmt.Printf("  💻 Built with Go %s | Author: %s\n", runtime.Version(), author)
	fmt.Printf("  🎯 Better than Gobuster - Faster, Smarter, More Features\n")
	fmt.Printf("  🔗 https://github.com/isa-programmer/goscan\n\n")
}

// GetVersion returns the current version
func GetVersion() string {
	return version
}

// GetAuthor returns the author name
func GetAuthor() string {
	return author
}
