package config

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config holds all configuration options
type Config struct {
	// Target configuration
	Target    string
	Wordlist  string
	OutputDir string
	OutputFile string

	// HTTP configuration
	Method         string
	UserAgent      string
	Headers        map[string]string
	Cookies        string
	Timeout        time.Duration
	FollowRedirect bool
	MaxRedirects   int
	Insecure       bool

	// Performance configuration
	Threads    int
	Delay      time.Duration
	Retries    int
	RateLimit  int
	BurstLimit int

	// Proxy configuration
	Proxy     string
	ProxyAuth string

	// Filtering configuration
	Extensions     []string
	StatusCodes    []int
	HideStatusCodes []int
	HideLength     []int
	HideWords      []int
	HideLines      []int
	ShowLength     bool
	ShowWords      bool
	ShowLines      bool
	ShowSize       bool
	ShowTime       bool

	// Content filtering
	IncludeRegex   string
	ExcludeRegex   string
	MatchString    string
	ExcludeString  string

	// Output configuration
	Verbose       bool
	Quiet         bool
	NoProgress    bool
	NoColor       bool
	JSON          bool
	CSV           bool
	XML           bool

	// Advanced features
	Recursive      bool
	MaxDepth       int
	Wildcard       bool
	AppendSlash    bool
	Lowercase      bool
	Uppercase      bool
	Capitalize     bool
	AddExtensions  bool
	RemoveExtensions bool

	// Authentication
	BasicAuth     string
	BearerToken   string
	NTLMAuth      string
	DigestAuth    string

	// SSL/TLS configuration
	ClientCert    string
	ClientKey     string
	CACert        string
	TLSVersion    string

	// Wordlist configuration
	WordlistMode  string // single, multiple, stdin
	Wordlists     []string
	WordlistDir   string

	// Discovery modes
	DirMode       bool
	FileMode      bool
	VhostMode     bool
	SubdomainMode bool
	APIMode       bool

	// Backup and common files
	BackupFiles   bool
	CommonFiles   bool
	ConfigFiles   bool
	LogFiles      bool
	TempFiles     bool

	// Version and help
	Version       bool
	Help          bool
}

// Parse parses command line arguments and returns configuration
func Parse() (*Config, error) {
	cfg := &Config{}

	// Target configuration
	flag.StringVar(&cfg.Target, "u", "", "Target URL (required)")
	flag.StringVar(&cfg.Target, "url", "", "Target URL (required)")
	flag.StringVar(&cfg.Wordlist, "w", "", "Wordlist file path (required)")
	flag.StringVar(&cfg.Wordlist, "wordlist", "", "Wordlist file path (required)")
	flag.StringVar(&cfg.OutputDir, "o", "", "Output directory")
	flag.StringVar(&cfg.OutputFile, "output", "", "Output file path")

	// HTTP configuration
	flag.StringVar(&cfg.Method, "m", "GET", "HTTP method (GET, POST, PUT, DELETE, HEAD, OPTIONS)")
	flag.StringVar(&cfg.Method, "method", "GET", "HTTP method")
	flag.StringVar(&cfg.UserAgent, "a", "goscan/2.1.0 (Advanced Directory Scanner)", "User-Agent string")
	flag.StringVar(&cfg.UserAgent, "user-agent", "goscan/2.1.0 (Advanced Directory Scanner)", "User-Agent string")
	headersFlag := flag.String("H", "", "Custom headers (format: 'Header1:Value1,Header2:Value2')")
	headersFlag2 := flag.String("headers", "", "Custom headers")
	flag.StringVar(&cfg.Cookies, "c", "", "Cookies (format: 'name1=value1; name2=value2')")
	flag.StringVar(&cfg.Cookies, "cookies", "", "Cookies")
	timeoutFlag := flag.Int("timeout", 10, "HTTP timeout in seconds")
	flag.BoolVar(&cfg.FollowRedirect, "r", false, "Follow redirects")
	flag.BoolVar(&cfg.FollowRedirect, "follow-redirect", false, "Follow redirects")
	flag.IntVar(&cfg.MaxRedirects, "max-redirects", 10, "Maximum number of redirects to follow")
	flag.BoolVar(&cfg.Insecure, "k", false, "Skip SSL certificate verification")
	flag.BoolVar(&cfg.Insecure, "insecure", false, "Skip SSL certificate verification")

	// Performance configuration
	flag.IntVar(&cfg.Threads, "t", 50, "Number of concurrent threads")
	flag.IntVar(&cfg.Threads, "threads", 50, "Number of concurrent threads")
	delayFlag := flag.Int("delay", 0, "Delay between requests in milliseconds")
	flag.IntVar(&cfg.Retries, "retries", 3, "Number of retries for failed requests")
	flag.IntVar(&cfg.RateLimit, "rate-limit", 0, "Rate limit (requests per second)")
	flag.IntVar(&cfg.BurstLimit, "burst-limit", 10, "Burst limit for rate limiting")

	// Proxy configuration
	flag.StringVar(&cfg.Proxy, "p", "", "Proxy URL (http://proxy:port or socks5://proxy:port)")
	flag.StringVar(&cfg.Proxy, "proxy", "", "Proxy URL")
	flag.StringVar(&cfg.ProxyAuth, "proxy-auth", "", "Proxy authentication (username:password)")

	// Filtering configuration
	extensionsFlag := flag.String("x", "", "File extensions (comma separated)")
	extensionsFlag2 := flag.String("extensions", "", "File extensions")
	statusCodesFlag := flag.String("s", "", "Status codes to show (comma separated)")
	statusCodesFlag2 := flag.String("status-codes", "", "Status codes to show")
	hideStatusFlag := flag.String("hide-status", "", "Status codes to hide (comma separated)")
	hideLengthFlag := flag.String("hide-length", "", "Hide responses with these lengths (comma separated)")
	hideWordsFlag := flag.String("hide-words", "", "Hide responses with these word counts")
	hideLinesFlag := flag.String("hide-lines", "", "Hide responses with these line counts")
	flag.BoolVar(&cfg.ShowLength, "l", false, "Show response length")
	flag.BoolVar(&cfg.ShowLength, "length", false, "Show response length")
	flag.BoolVar(&cfg.ShowWords, "words", false, "Show response word count")
	flag.BoolVar(&cfg.ShowLines, "lines", false, "Show response line count")
	flag.BoolVar(&cfg.ShowSize, "size", true, "Show response size")
	flag.BoolVar(&cfg.ShowTime, "time", false, "Show response time")

	// Content filtering
	flag.StringVar(&cfg.IncludeRegex, "include-regex", "", "Include responses matching regex")
	flag.StringVar(&cfg.ExcludeRegex, "exclude-regex", "", "Exclude responses matching regex")
	flag.StringVar(&cfg.MatchString, "match-string", "", "Include responses containing string")
	flag.StringVar(&cfg.ExcludeString, "exclude-string", "", "Exclude responses containing string")

	// Output configuration
	flag.BoolVar(&cfg.Verbose, "v", false, "Verbose output")
	flag.BoolVar(&cfg.Verbose, "verbose", false, "Verbose output")
	flag.BoolVar(&cfg.Quiet, "q", false, "Quiet mode (no progress)")
	flag.BoolVar(&cfg.Quiet, "quiet", false, "Quiet mode")
	flag.BoolVar(&cfg.NoProgress, "no-progress", false, "Disable progress bar")
	flag.BoolVar(&cfg.NoColor, "no-color", false, "Disable colored output")
	flag.BoolVar(&cfg.JSON, "json", false, "Output in JSON format")
	flag.BoolVar(&cfg.CSV, "csv", false, "Output in CSV format")
	flag.BoolVar(&cfg.XML, "xml", false, "Output in XML format")

	// Advanced features
	flag.BoolVar(&cfg.Recursive, "recursive", false, "Recursive directory scanning")
	flag.IntVar(&cfg.MaxDepth, "max-depth", 3, "Maximum recursion depth")
	flag.BoolVar(&cfg.Wildcard, "wildcard", false, "Use wildcard for subdirectory discovery")
	flag.BoolVar(&cfg.AppendSlash, "append-slash", false, "Append slash to directories")
	flag.BoolVar(&cfg.Lowercase, "lowercase", false, "Convert wordlist to lowercase")
	flag.BoolVar(&cfg.Uppercase, "uppercase", false, "Convert wordlist to uppercase")
	flag.BoolVar(&cfg.Capitalize, "capitalize", false, "Capitalize first letter of words")
	flag.BoolVar(&cfg.AddExtensions, "add-extensions", false, "Add extensions to words without extensions")
	flag.BoolVar(&cfg.RemoveExtensions, "remove-extensions", false, "Remove extensions from words")

	// Authentication
	flag.StringVar(&cfg.BasicAuth, "basic-auth", "", "Basic authentication (username:password)")
	flag.StringVar(&cfg.BearerToken, "bearer-token", "", "Bearer token for authentication")
	flag.StringVar(&cfg.NTLMAuth, "ntlm-auth", "", "NTLM authentication (domain\\username:password)")
	flag.StringVar(&cfg.DigestAuth, "digest-auth", "", "Digest authentication (username:password)")

	// SSL/TLS configuration
	flag.StringVar(&cfg.ClientCert, "client-cert", "", "Client certificate file")
	flag.StringVar(&cfg.ClientKey, "client-key", "", "Client private key file")
	flag.StringVar(&cfg.CACert, "ca-cert", "", "CA certificate file")
	flag.StringVar(&cfg.TLSVersion, "tls-version", "", "TLS version (1.0, 1.1, 1.2, 1.3)")

	// Wordlist configuration
	flag.StringVar(&cfg.WordlistMode, "wordlist-mode", "single", "Wordlist mode (single, multiple, stdin)")
	wordlistsFlag := flag.String("wordlists", "", "Multiple wordlist files (comma separated)")
	flag.StringVar(&cfg.WordlistDir, "wordlist-dir", "", "Directory containing wordlist files")

	// Discovery modes
	flag.BoolVar(&cfg.DirMode, "dir", true, "Directory discovery mode")
	flag.BoolVar(&cfg.FileMode, "file", false, "File discovery mode")
	flag.BoolVar(&cfg.VhostMode, "vhost", false, "Virtual host discovery mode")
	flag.BoolVar(&cfg.SubdomainMode, "subdomain", false, "Subdomain discovery mode")
	flag.BoolVar(&cfg.APIMode, "api", false, "API endpoint discovery mode")

	// Backup and common files
	flag.BoolVar(&cfg.BackupFiles, "backup-files", false, "Include common backup file extensions")
	flag.BoolVar(&cfg.CommonFiles, "common-files", false, "Include common file names")
	flag.BoolVar(&cfg.ConfigFiles, "config-files", false, "Include common config file names")
	flag.BoolVar(&cfg.LogFiles, "log-files", false, "Include common log file names")
	flag.BoolVar(&cfg.TempFiles, "temp-files", false, "Include common temporary file names")

	// Version and help
	flag.BoolVar(&cfg.Version, "version", false, "Show version information")
	flag.BoolVar(&cfg.Help, "h", false, "Show help")
	flag.BoolVar(&cfg.Help, "help", false, "Show help")

	// Custom usage function
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "GoScan v2.1.0 - Advanced Directory & File Scanner\n\n")
		fmt.Fprintf(os.Stderr, "Usage: %s [options]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Required:\n")
		fmt.Fprintf(os.Stderr, "  -u, --url string         Target URL\n")
		fmt.Fprintf(os.Stderr, "  -w, --wordlist string    Wordlist file path\n\n")
		fmt.Fprintf(os.Stderr, "HTTP Options:\n")
		fmt.Fprintf(os.Stderr, "  -m, --method string      HTTP method (default \"GET\")\n")
		fmt.Fprintf(os.Stderr, "  -a, --user-agent string  User-Agent string\n")
		fmt.Fprintf(os.Stderr, "  -H, --headers string     Custom headers\n")
		fmt.Fprintf(os.Stderr, "  -c, --cookies string     Cookies\n")
		fmt.Fprintf(os.Stderr, "  --timeout int            HTTP timeout in seconds (default 10)\n")
		fmt.Fprintf(os.Stderr, "  -r, --follow-redirect    Follow redirects\n")
		fmt.Fprintf(os.Stderr, "  -k, --insecure           Skip SSL certificate verification\n\n")
		fmt.Fprintf(os.Stderr, "Performance Options:\n")
		fmt.Fprintf(os.Stderr, "  -t, --threads int        Number of concurrent threads (default 50)\n")
		fmt.Fprintf(os.Stderr, "  --delay int              Delay between requests in milliseconds\n")
		fmt.Fprintf(os.Stderr, "  --retries int            Number of retries (default 3)\n")
		fmt.Fprintf(os.Stderr, "  --rate-limit int         Rate limit (requests per second)\n\n")
		fmt.Fprintf(os.Stderr, "Filtering Options:\n")
		fmt.Fprintf(os.Stderr, "  -x, --extensions string  File extensions (comma separated)\n")
		fmt.Fprintf(os.Stderr, "  -s, --status-codes string Status codes to show\n")
		fmt.Fprintf(os.Stderr, "  --hide-status string     Status codes to hide\n")
		fmt.Fprintf(os.Stderr, "  --hide-length string     Hide responses with these lengths\n")
		fmt.Fprintf(os.Stderr, "  -l, --length             Show response length\n")
		fmt.Fprintf(os.Stderr, "  --size                   Show response size (default true)\n\n")
		fmt.Fprintf(os.Stderr, "Output Options:\n")
		fmt.Fprintf(os.Stderr, "  -v, --verbose            Verbose output\n")
		fmt.Fprintf(os.Stderr, "  -q, --quiet              Quiet mode\n")
		fmt.Fprintf(os.Stderr, "  -o string                Output directory\n")
		fmt.Fprintf(os.Stderr, "  --output string          Output file path\n")
		fmt.Fprintf(os.Stderr, "  --json                   Output in JSON format\n")
		fmt.Fprintf(os.Stderr, "  --csv                    Output in CSV format\n\n")
		fmt.Fprintf(os.Stderr, "Discovery Modes:\n")
		fmt.Fprintf(os.Stderr, "  --dir                    Directory discovery (default true)\n")
		fmt.Fprintf(os.Stderr, "  --file                   File discovery\n")
		fmt.Fprintf(os.Stderr, "  --vhost                  Virtual host discovery\n")
		fmt.Fprintf(os.Stderr, "  --subdomain              Subdomain discovery\n")
		fmt.Fprintf(os.Stderr, "  --api                    API endpoint discovery\n\n")
		fmt.Fprintf(os.Stderr, "Examples:\n")
		fmt.Fprintf(os.Stderr, "  %s -u https://example.com -w wordlist.txt\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -u https://example.com -w dirs.txt -t 100 -x php,html\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -u https://example.com -w files.txt -s 200,301,403 -v\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -u https://example.com -w api.txt --api --json -o results/\n", os.Args[0])
	}

	flag.Parse()

	// Handle version flag
	if cfg.Version {
		fmt.Println("GoScan v2.1.0")
		os.Exit(0)
	}

	// Handle help flag
	if cfg.Help {
		flag.Usage()
		os.Exit(0)
	}

	// Validate required arguments
	if cfg.Target == "" || cfg.Wordlist == "" {
		flag.Usage()
		return nil, fmt.Errorf("target URL and wordlist are required")
	}

	// Parse additional flags
	cfg.Timeout = time.Duration(*timeoutFlag) * time.Second
	cfg.Delay = time.Duration(*delayFlag) * time.Millisecond

	// Parse headers
	headerStr := *headersFlag
	if headerStr == "" {
		headerStr = *headersFlag2
	}
	cfg.Headers = parseHeaders(headerStr)

	// Parse extensions
	extStr := *extensionsFlag
	if extStr == "" {
		extStr = *extensionsFlag2
	}
	cfg.Extensions = parseStringSlice(extStr)

	// Parse status codes
	statusStr := *statusCodesFlag
	if statusStr == "" {
		statusStr = *statusCodesFlag2
	}
	cfg.StatusCodes = parseIntSlice(statusStr)
	cfg.HideStatusCodes = parseIntSlice(*hideStatusFlag)
	cfg.HideLength = parseIntSlice(*hideLengthFlag)
	cfg.HideWords = parseIntSlice(*hideWordsFlag)
	cfg.HideLines = parseIntSlice(*hideLinesFlag)

	// Parse wordlists
	cfg.Wordlists = parseStringSlice(*wordlistsFlag)

	// Ensure target URL has proper format
	if !strings.HasPrefix(cfg.Target, "http://") && !strings.HasPrefix(cfg.Target, "https://") {
		cfg.Target = "http://" + cfg.Target
	}
	if !strings.HasSuffix(cfg.Target, "/") {
		cfg.Target += "/"
	}

	return cfg, nil
}

// parseHeaders parses header string into map
func parseHeaders(input string) map[string]string {
	headers := make(map[string]string)
	if input == "" {
		return headers
	}

	parts := strings.Split(input, ",")
	for _, part := range parts {
		if kv := strings.SplitN(strings.TrimSpace(part), ":", 2); len(kv) == 2 {
			headers[strings.TrimSpace(kv[0])] = strings.TrimSpace(kv[1])
		}
	}
	return headers
}

// parseIntSlice parses comma-separated integers
func parseIntSlice(input string) []int {
	if input == "" {
		return nil
	}

	var nums []int
	parts := strings.Split(input, ",")
	for _, part := range parts {
		if num, err := strconv.Atoi(strings.TrimSpace(part)); err == nil {
			nums = append(nums, num)
		}
	}
	return nums
}

// parseStringSlice parses comma-separated strings
func parseStringSlice(input string) []string {
	if input == "" {
		return nil
	}

	var strs []string
	parts := strings.Split(input, ",")
	for _, part := range parts {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			strs = append(strs, trimmed)
		}
	}
	return strs
}