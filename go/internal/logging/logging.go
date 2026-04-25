// Package logging provides structured, colourised logging for the tfenv Go
// edition. It wraps charmbracelet/log to give beautiful TTY output with
// automatic colour detection, structured JSON mode, and environment-driven
// configuration.
//
// Environment variables:
//
//   - TFENV_LOG_LEVEL  — error | warn | info (default) | debug
//   - TFENV_LOG_FORMAT — text (default) | json
//   - NO_COLOR         — if set, disables colour output (https://no-color.org/)
//
// All log output goes to stderr. stdout is reserved for command output.
package logging

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
	charlog "github.com/charmbracelet/log"
	"github.com/muesli/termenv"
)

// Level represents a logging verbosity level.
type Level = charlog.Level

// Exported level constants for callers that need to reference them directly.
const (
	DebugLevel Level = charlog.DebugLevel
	InfoLevel  Level = charlog.InfoLevel
	WarnLevel  Level = charlog.WarnLevel
	ErrorLevel Level = charlog.ErrorLevel
)

// Logger wraps a charmbracelet/log.Logger and adds tfenv-specific behaviour.
type Logger struct {
	inner *charlog.Logger
}

// defaultLogger is the package-level logger used by the convenience functions.
var defaultLogger *Logger

func init() {
	// Provide a usable default so callers that forget Init() still get output.
	defaultLogger = newFromEnvFunc(os.Getenv)
}

// Init reads environment variables and (re)configures the package-level
// default logger. Call once at program startup.
func Init() {
	defaultLogger = newFromEnvFunc(os.Getenv)
}

// envFunc abstracts os.Getenv for testability.
type envFunc func(string) string

// newFromEnvFunc builds a Logger from the given environment lookup function.
func newFromEnvFunc(getenv envFunc) *Logger {
	level := resolveLevel(getenv("TFENV_LOG_LEVEL"))
	format := resolveFormat(getenv("TFENV_LOG_FORMAT"))
	noColor := getenv("NO_COLOR") != ""

	return newLogger(level, format, noColor)
}

// New creates a new Logger by reading the current environment variables.
func New() *Logger {
	return newFromEnvFunc(os.Getenv)
}

// newLogger constructs a Logger with the given configuration.
func newLogger(level Level, format charlog.Formatter, noColor bool) *Logger {
	// Debug always reports caller, other levels do not.
	reportCaller := level <= DebugLevel

	inner := charlog.NewWithOptions(os.Stderr, charlog.Options{
		Level:           level,
		Formatter:       format,
		ReportTimestamp: true,
		ReportCaller:    reportCaller,
		CallerFormatter: charlog.ShortCallerFormatter,
		// CallerOffset 1 accounts for our wrapper layer so the reported
		// call site is the caller of Debug/Info/Warn/Error, not logging.go.
		CallerOffset: 1,
	})

	if noColor {
		inner.SetColorProfile(termenv.Ascii)
	}

	if format == charlog.TextFormatter && !noColor {
		inner.SetStyles(tfenvStyles())
	}

	return &Logger{inner: inner}
}

// resolveLevel parses TFENV_LOG_LEVEL into a charlog.Level. Unrecognised
// values produce a warning on stderr and fall back to info.
func resolveLevel(raw string) Level {
	if raw == "" {
		return InfoLevel
	}
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "debug":
		return DebugLevel
	case "info":
		return InfoLevel
	case "warn":
		return WarnLevel
	case "error":
		return ErrorLevel
	default:
		fmt.Fprintf(os.Stderr,
			"tfenv: invalid TFENV_LOG_LEVEL %q (valid: debug, info, warn, error); defaulting to info\n",
			raw)
		return InfoLevel
	}
}

// resolveFormat parses TFENV_LOG_FORMAT into a charlog.Formatter. Defaults to
// TextFormatter.
func resolveFormat(raw string) charlog.Formatter {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "json":
		return charlog.JSONFormatter
	default:
		return charlog.TextFormatter
	}
}

// tfenvStyles returns custom Lipgloss styles for the text formatter.
// ERROR: red bold, WARN: yellow, INFO: blue, DEBUG: grey/dim.
func tfenvStyles() *charlog.Styles {
	s := charlog.DefaultStyles()
	s.Levels[ErrorLevel] = lipgloss.NewStyle().
		SetString("ERROR").
		Bold(true).
		Foreground(lipgloss.Color("9")) // bright red
	s.Levels[WarnLevel] = lipgloss.NewStyle().
		SetString("WARN").
		Foreground(lipgloss.Color("11")) // bright yellow
	s.Levels[InfoLevel] = lipgloss.NewStyle().
		SetString("INFO").
		Foreground(lipgloss.Color("12")) // bright blue
	s.Levels[DebugLevel] = lipgloss.NewStyle().
		SetString("DEBUG").
		Faint(true) // grey/dim
	return s
}

// ---------------------------------------------------------------------------
// Logger instance methods
// ---------------------------------------------------------------------------

// Debug logs a debug-level message with optional key-value pairs.
func (l *Logger) Debug(msg string, keyvals ...any) {
	l.inner.Debug(msg, keyvals...)
}

// Info logs an info-level message with optional key-value pairs.
func (l *Logger) Info(msg string, keyvals ...any) {
	l.inner.Info(msg, keyvals...)
}

// Warn logs a warn-level message with optional key-value pairs.
func (l *Logger) Warn(msg string, keyvals ...any) {
	l.inner.Warn(msg, keyvals...)
}

// Error logs an error-level message with optional key-value pairs.
func (l *Logger) Error(msg string, keyvals ...any) {
	l.inner.Error(msg, keyvals...)
}

// With returns a new Logger carrying the given key-value pairs on every
// subsequent log line.
func (l *Logger) With(keyvals ...any) *Logger {
	return &Logger{inner: l.inner.With(keyvals...)}
}

// GetLevel returns the currently configured level.
func (l *Logger) GetLevel() Level {
	return l.inner.GetLevel()
}

// SetLevel dynamically changes the log level.
func (l *Logger) SetLevel(level Level) {
	l.inner.SetLevel(level)
}

// Inner exposes the underlying charmbracelet/log.Logger for advanced use
// cases such as using it as an slog.Handler.
func (l *Logger) Inner() *charlog.Logger {
	return l.inner
}

// ---------------------------------------------------------------------------
// Package-level convenience functions (use defaultLogger)
// ---------------------------------------------------------------------------

// Debug logs a debug-level message on the default logger.
func Debug(msg string, keyvals ...any) {
	defaultLogger.inner.Debug(msg, keyvals...)
}

// Info logs an info-level message on the default logger.
func Info(msg string, keyvals ...any) {
	defaultLogger.inner.Info(msg, keyvals...)
}

// Warn logs a warn-level message on the default logger.
func Warn(msg string, keyvals ...any) {
	defaultLogger.inner.Warn(msg, keyvals...)
}

// Error logs an error-level message on the default logger.
func Error(msg string, keyvals ...any) {
	defaultLogger.inner.Error(msg, keyvals...)
}

// With returns a new Logger derived from the default logger, carrying the
// given key-value pairs.
func With(keyvals ...any) *Logger {
	return &Logger{inner: defaultLogger.inner.With(keyvals...)}
}

// Default returns the package-level default Logger.
func Default() *Logger {
	return defaultLogger
}

// SetDefault replaces the package-level default Logger.
func SetDefault(l *Logger) {
	defaultLogger = l
}
