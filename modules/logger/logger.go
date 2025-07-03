package logger

import (
	"fmt"
	"os"
	"time"
)

// LogLevel represents different log levels
type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	WARN
	ERROR
	FATAL
)

// Logger handles all logging operations
type Logger struct {
	verbose   bool
	logFile   *os.File
	logLevel  LogLevel
	showTime  bool
	colorized bool
}

// Colors for terminal output
const (
	ColorReset  = "\033[0m"
	ColorRed    = "\033[31m"
	ColorGreen  = "\033[32m"
	ColorYellow = "\033[33m"
	ColorBlue   = "\033[34m"
	ColorPurple = "\033[35m"
	ColorCyan   = "\033[36m"
	ColorWhite  = "\033[37m"
	ColorGray   = "\033[90m"
)

// New creates a new Logger with the specified verbosity, INFO log level, timestamps, and colorized output enabled.
func New(verbose bool) *Logger {
	return &Logger{
		verbose:   verbose,
		logLevel:  INFO,
		showTime:  true,
		colorized: true,
	}
}

// NewWithFile creates a new Logger with the specified verbosity and configures it to write logs to the given file path in addition to standard output.
// Returns an error if the log file cannot be opened.
func NewWithFile(verbose bool, logFilePath string) (*Logger, error) {
	logger := New(verbose)
	
	if logFilePath != "" {
		file, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			return nil, fmt.Errorf("failed to open log file: %v", err)
		}
		logger.logFile = file
	}
	
	return logger, nil
}

// SetLevel sets the minimum log level
func (l *Logger) SetLevel(level LogLevel) {
	l.logLevel = level
}

// SetColorized enables or disables colored output
func (l *Logger) SetColorized(enabled bool) {
	l.colorized = enabled
}

// SetShowTime enables or disables timestamp in logs
func (l *Logger) SetShowTime(enabled bool) {
	l.showTime = enabled
}

// Debug logs debug messages (only in verbose mode)
func (l *Logger) Debug(message string) {
	if l.verbose && l.logLevel <= DEBUG {
		l.log("DEBUG", ColorGray, message)
	}
}

// Info logs informational messages
func (l *Logger) Info(message string) {
	if l.logLevel <= INFO {
		l.log("INFO", ColorBlue, message)
	}
}

// Warn logs warning messages
func (l *Logger) Warn(message string) {
	if l.logLevel <= WARN {
		l.log("WARN", ColorYellow, message)
	}
}

// Error logs error messages
func (l *Logger) Error(message string) {
	if l.logLevel <= ERROR {
		l.log("ERROR", ColorRed, message)
	}
}

// Fatal logs fatal messages and exits
func (l *Logger) Fatal(message string) {
	l.log("FATAL", ColorRed, message)
	os.Exit(1)
}

// Success logs success messages
func (l *Logger) Success(message string) {
	if l.logLevel <= INFO {
		l.log("SUCCESS", ColorGreen, message)
	}
}

// Progress logs progress messages (only in verbose mode)
func (l *Logger) Progress(message string) {
	if l.verbose {
		l.log("PROGRESS", ColorCyan, message)
	}
}

// Request logs HTTP request details (only in verbose mode)
func (l *Logger) Request(method, url string, statusCode int) {
	if l.verbose {
		color := ColorGreen
		if statusCode >= 400 {
			color = ColorRed
		} else if statusCode >= 300 {
			color = ColorYellow
		}
		message := fmt.Sprintf("%s %s -> %d", method, url, statusCode)
		l.log("REQUEST", color, message)
	}
}

// Statistics logs scanning statistics
func (l *Logger) Statistics(found, total int, elapsed time.Duration) {
	message := fmt.Sprintf("Found: %d/%d | Elapsed: %v | Rate: %.2f req/s", 
		found, total, elapsed, float64(total)/elapsed.Seconds())
	l.log("STATS", ColorPurple, message)
}

// log is the internal logging function
func (l *Logger) log(levelStr, color, message string) {
	timestamp := ""
	if l.showTime {
		timestamp = time.Now().Format("15:04:05") + " "
	}
	
	// Format message
	var formattedMessage string
	if l.colorized {
		formattedMessage = fmt.Sprintf("%s[%s%s%s] %s%s", 
			timestamp, color, levelStr, ColorReset, message, ColorReset)
	} else {
		formattedMessage = fmt.Sprintf("%s[%s] %s", timestamp, levelStr, message)
	}
	
	// Output to console
	fmt.Println(formattedMessage)
	
	// Output to file if configured
	if l.logFile != nil {
		fileMessage := fmt.Sprintf("%s[%s] %s\n", timestamp, levelStr, message)
		l.logFile.WriteString(fileMessage)
	}
}

// Close closes the log file if open
func (l *Logger) Close() error {
	if l.logFile != nil {
		return l.logFile.Close()
	}
	return nil
}

// Printf provides printf-style logging
func (l *Logger) Printf(format string, args ...interface{}) {
	l.Info(fmt.Sprintf(format, args...))
}

// Debugf provides printf-style debug logging
func (l *Logger) Debugf(format string, args ...interface{}) {
	l.Debug(fmt.Sprintf(format, args...))
}

// Errorf provides printf-style error logging
func (l *Logger) Errorf(format string, args ...interface{}) {
	l.Error(fmt.Sprintf(format, args...))
}