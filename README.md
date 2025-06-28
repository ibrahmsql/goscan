# goscan

A cross-platform directory scanner written in Go.

## Overview

**goscan** is a powerful and efficient tool for scanning directories across multiple platforms. Built with Go, it offers fast performance and easy deployment as a single binary. Use it to enumerate files, analyze directory structures, or integrate it into your automation workflows.

## Features

- 🚀 **High performance** directory scanning
- 💻 **Cross-platform** support (Windows, macOS, Linux)
- 🔎 **Recursive** file and folder enumeration
- 📦 **Single binary** deployment
- 🛠️ Easily extensible for custom use-cases

## Installation

<!-- ### Download Pre-built Binary

Pre-built binaries will be available in the [Releases](https://github.com/isa-programmer/goscan/releases) section.
-->
### Build from Source

1. **Clone the repository:**
   ```bash
   git clone https://github.com/isa-programmer/goscan.git
   cd goscan
   ```

2. **Build the binary:**
   ```bash
   go build -o goscan
   ```

## Usage

```bash
goscan [flags] <directory>
```

### Example

```bash
./goscan /path/to/scan
```

### Flags

- `-depth int` : Maximum depth to scan (default: unlimited)
- `-hidden`    : Include hidden files and directories
- `-exclude string` : Comma-separated list of patterns to exclude

> **Tip:** Run `goscan -h` to see all available options.

## Contributing

Contributions are welcome! Please open issues or pull requests to improve **goscan**.

1. Fork the repository.
2. Create your feature branch (`git checkout -b feature/your-feature`).
3. Commit your changes (`git commit -am 'Add new feature'`).
4. Push to the branch (`git push origin feature/your-feature`).
5. Create a new Pull Request.

## License

This project is licensed under the [MIT License](LICENSE).

## Acknowledgements

- Built with [Go](https://golang.org/)

---

> **goscan** - Making directory scanning simple and efficient 🚀
```
```
