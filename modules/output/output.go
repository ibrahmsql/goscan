package output

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/isa-programmer/goscan/modules/scanner"
)

// OutputFormat represents different output formats
type OutputFormat string

const (
	FormatJSON OutputFormat = "json"
	FormatTXT  OutputFormat = "txt"
	FormatCSV  OutputFormat = "csv"
	FormatHTML OutputFormat = "html"
)

// OutputManager handles different output formats
type OutputManager struct {
	Format   OutputFormat
	FilePath string
	Results  []scanner.Result
}

// New returns a new OutputManager configured with the specified output format and file path.
func New(format OutputFormat, filePath string) *OutputManager {
	return &OutputManager{
		Format:   format,
		FilePath: filePath,
		Results:  make([]scanner.Result, 0),
	}
}

// AddResult adds a result to the output manager
func (om *OutputManager) AddResult(result scanner.Result) {
	om.Results = append(om.Results, result)
}

// Save saves results to file in the specified format
func (om *OutputManager) Save() error {
	switch om.Format {
	case FormatJSON:
		return om.saveJSON()
	case FormatTXT:
		return om.saveTXT()
	case FormatCSV:
		return om.saveCSV()
	case FormatHTML:
		return om.saveHTML()
	default:
		return fmt.Errorf("unsupported format: %s", om.Format)
	}
}

// saveJSON saves results in JSON format
func (om *OutputManager) saveJSON() error {
	data, err := json.MarshalIndent(om.Results, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(om.FilePath, data, 0644)
}

// saveTXT saves results in plain text format
func (om *OutputManager) saveTXT() error {
	var content strings.Builder
	content.WriteString(fmt.Sprintf("Goscan Results - %s\n", time.Now().Format("2006-01-02 15:04:05")))
	content.WriteString(strings.Repeat("=", 50) + "\n")
	
	for _, result := range om.Results {
		if result.StatusCode != 0 {
			content.WriteString(fmt.Sprintf("[%d] %s\n", result.StatusCode, result.URL))
		}
	}
	
	return os.WriteFile(om.FilePath, []byte(content.String()), 0644)
}

// saveCSV saves results in CSV format
func (om *OutputManager) saveCSV() error {
	var content strings.Builder
	content.WriteString("URL,Status Code,Response Time\n")
	
	for _, result := range om.Results {
		if result.StatusCode != 0 {
			content.WriteString(fmt.Sprintf("%s,%d,%s\n", result.URL, result.StatusCode, result.ResponseTime.String()))
		}
	}
	
	return os.WriteFile(om.FilePath, []byte(content.String()), 0644)
}

// saveHTML saves results in HTML format
func (om *OutputManager) saveHTML() error {
	var content strings.Builder
	content.WriteString(`<!DOCTYPE html>
<html>
<head>
	<title>Goscan Results</title>
	<style>
		body { font-family: Arial, sans-serif; margin: 20px; }
		table { border-collapse: collapse; width: 100%; }
		th, td { border: 1px solid #ddd; padding: 8px; text-align: left; }
		th { background-color: #f2f2f2; }
		.success { color: green; }
		.error { color: red; }
	</style>
</head>
<body>
	<h1>Goscan Results</h1>
	<p>Generated: ` + time.Now().Format("2006-01-02 15:04:05") + `</p>
	<table>
		<tr><th>URL</th><th>Status Code</th><th>Response Time</th></tr>
`)
	
	for _, result := range om.Results {
		if result.StatusCode != 0 {
			class := "success"
			if result.StatusCode >= 400 {
				class = "error"
			}
			content.WriteString(fmt.Sprintf(
				"\t\t<tr><td>%s</td><td class=\"%s\">%d</td><td>%s</td></tr>\n",
				result.URL, class, result.StatusCode, result.ResponseTime.String(),
			))
		}
	}
	
	content.WriteString(`	</table>
</body>
</html>`)
	
	return os.WriteFile(om.FilePath, []byte(content.String()), 0644)
}

// GetSuccessCount returns the number of successful results
func (om *OutputManager) GetSuccessCount() int {
	count := 0
	for _, result := range om.Results {
		if result.StatusCode != 0 && result.StatusCode < 400 {
			count++
		}
	}
	return count
}

// GetFailedCount returns the number of failed results
func (om *OutputManager) GetFailedCount() int {
	count := 0
	for _, result := range om.Results {
		if result.StatusCode == 0 {
			count++
		}
	}
	return count
}

// GetErrorCount returns the number of error results (4xx, 5xx)
func (om *OutputManager) GetErrorCount() int {
	count := 0
	for _, result := range om.Results {
		if result.StatusCode >= 400 {
			count++
		}
	}
	return count
}