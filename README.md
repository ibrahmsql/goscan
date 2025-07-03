# goscan

A blazing fast directory and API endpoint scanner written in Go. goscan provides two powerful tools for penetration testers, security researchers, and developers who need to quickly discover hidden directories, files, and API endpoints on web servers.

## 🚀 Tools

### goscan - Directory & File Scanner
 is a powerful and efficient tool for scanning directories across multiple platforms. Built with Go, it offers fast performance and easy deployment as a single binary. Use it to enumerate files, analyze directory structures, or integrate it into your automation workflows.

### Apiscan - API Endpoint Scanner  
Apiscan is a specialized tool for discovering and testing API endpoints with multiple HTTP methods.

## ✨ Features

- **Lightning Fast**: Multi-threaded scanning with configurable concurrency
- **Dual Scanning Modes**: Directory scanning and API endpoint discovery
- **Comprehensive Wordlists**: 13 specialized wordlists for different scanning scenarios
- **Multiple HTTP Methods**: API scanner tests GET, POST, PUT, DELETE, PATCH, OPTIONS, HEAD
- **Multiple Output Formats**: Text, JSON, and CSV output support
- **Smart Filtering**: Configurable status code filtering and response analysis
- **Security Focused**: Built-in wordlists for security testing and vulnerability assessment
- **Cross-Platform**: Works on Linux, macOS, and Windows
- **Lightweight**: Single binaries with no external dependencies
- **Customizable**: Extensive configuration options for different use cases

## Installation

### From Source

```bash
git clone https://github.com/isa-programmer/goscan.git
cd goscan
make build
# or build individually:
# make goscan
# make apiscan
```

### Using Install Script

```bash
curl -sSL https://raw.githubusercontent.com/isa-programmer/goscan/main/install.sh | bash
```

## Usage

### Basic Usage

#### Directory Scanning (goscan)

```bash
# Basic directory scanning
./goscan wordlists/common-web-paths.txt https://example.com

# Scan with custom threads and timeout
./goscan wordlists/bigwordlist.txt https://example.com --threads 20 --timeout 15

# Scan with no warnings
./goscan wordlists/security-testing.txt https://example.com --no-warning
```

#### API Endpoint Scanning (apiscan)

```bash
# Basic API endpoint scanning
./apiscan wordlists/api-endpoints.txt https://api.example.com

# API scan with JSON output
./apiscan wordlists/api-endpoints.txt https://api.example.com --output results.json

# API scan with custom settings
./apiscan wordlists/api-endpoints.txt https://api.example.com --threads 20 --timeout 15 --quiet
```

### Advanced Usage

```bash
# Scan with custom headers
./goscan -u https://example.com -H "Authorization: Bearer token123"

# Use proxy
./goscan -u https://example.com -p http://127.0.0.1:8080

# Filter by status codes
./goscan -u https://example.com -mc 200,301,302

# Set custom User-Agent
./goscan -u https://example.com -a "Mozilla/5.0 Custom Scanner"

# Rate limiting (requests per second)
./goscan -u https://example.com -r 10
```

## Command Line Options

```
-u, --url string          Target URL to scan
-w, --wordlist string     Path to wordlist file (default: "wordlists/wordlist.txt")
-t, --threads int         Number of concurrent threads (default: 20)
-o, --output string       Output file path
-f, --format string       Output format: txt, json, csv (default: "txt")
-mc, --match-codes string Comma-separated list of status codes to match
-fc, --filter-codes string Comma-separated list of status codes to filter out
-H, --headers string      Custom headers (can be used multiple times)
-a, --user-agent string   Custom User-Agent string
-p, --proxy string        HTTP/HTTPS proxy URL
-r, --rate int           Requests per second rate limit
-k, --insecure           Skip SSL certificate verification
-v, --verbose            Enable verbose output
-h, --help               Show help message
```

## Wordlists

GoScan comes with 13 specialized wordlists for different scanning scenarios:

### Core Wordlists
- **`wordlist.txt`** - General purpose directory and file names
- **`bigwordlist.txt`** - Extended general purpose wordlist
- **`comprehensive-2025.txt`** - Modern comprehensive wordlist

### Technology-Specific Wordlists
- **`web-technologies.txt`** - Popular web frameworks and CMS paths
- **`cms-frameworks.txt`** - Detailed CMS and framework specific paths
- **`api-endpoints.txt`** - Modern API endpoints and services
- **`database-paths.txt`** - Database and data storage related paths

### Security-Focused Wordlists
- **`security-testing.txt`** - Penetration testing and security audit paths
- **`sensitive-files.txt`** - Configuration files, credentials, and security files
- **`backup-temp-files.txt`** - Backup files and temporary file patterns

### Infrastructure Wordlists
- **`server-infrastructure.txt`** - Server and infrastructure related paths
- **`common-web-paths.txt`** - Common web application paths and endpoints
- **`file-extensions.txt`** - Comprehensive file extensions collection

## Examples

### Security Testing

```bash
# Scan for sensitive files
./goscan -u https://target.com -w wordlists/sensitive-files.txt

# Look for backup files
./goscan -u https://target.com -w wordlists/backup-temp-files.txt

# Security-focused scanning
./goscan -u https://target.com -w wordlists/security-testing.txt
```

### CMS Detection

```bash
# Scan for CMS and framework files
./goscan -u https://target.com -w wordlists/cms-frameworks.txt

# Web technology detection
./goscan -u https://target.com -w wordlists/web-technologies.txt
```

### API Discovery

```bash
# Discover API endpoints
./goscan -u https://api.target.com -w wordlists/api-endpoints.txt

# Database path discovery
./goscan -u https://target.com -w wordlists/database-paths.txt
```

## Output Formats

### Text Output (Default)
```
[200] https://example.com/admin/
[301] https://example.com/login -> https://example.com/login/
[403] https://example.com/config/
```

### JSON Output
```json
{
  "url": "https://example.com/admin/",
  "status_code": 200,
  "content_length": 1234,
  "redirect_url": ""
}
```

### CSV Output
```csv
URL,Status Code,Content Length,Redirect URL
https://example.com/admin/,200,1234,
https://example.com/login,301,0,https://example.com/login/
```

## Performance Tips

1. **Adjust Thread Count**: Start with 20 threads and increase based on target capacity
2. **Use Rate Limiting**: Prevent overwhelming the target server with `-r` option
3. **Filter Responses**: Use `-mc` or `-fc` to focus on relevant status codes
4. **Choose Appropriate Wordlists**: Use specific wordlists for targeted scanning
5. **Monitor Resource Usage**: Watch CPU and memory usage during large scans

## Security Considerations

- **Authorization**: Ensure you have permission to scan the target
- **Rate Limiting**: Use appropriate delays to avoid DoS conditions
- **Proxy Usage**: Consider using proxies for anonymity when appropriate
- **SSL Verification**: Only use `-k` flag when necessary for testing

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Authors

Made by [isa-programmer](https://github.com/isa-programmer) & ibrahimsql[ibrahimsql](https://github.com/ibrahimsql)

## Disclaimer

This tool is for educational and authorized testing purposes only. Users are responsible for complying with applicable laws and regulations. The authors are not responsible for any misuse of this tool.
