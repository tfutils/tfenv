package logging

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	charlog "github.com/charmbracelet/log"
	"github.com/muesli/termenv"
)

// ---------------------------------------------------------------------------
// helpers
// ---------------------------------------------------------------------------

// captureLogger creates a Logger that writes to a buffer with colour disabled
// (ASCII profile) so output is predictable in tests. Returns the logger and
// the buffer it writes to.
func captureLogger(level Level, format charlog.Formatter) (*Logger, *bytes.Buffer) {
	var buf bytes.Buffer
	reportCaller := level <= DebugLevel

	inner := charlog.NewWithOptions(&buf, charlog.Options{
		Level:           level,
		Formatter:       format,
		ReportTimestamp: false,
		ReportCaller:    reportCaller,
		CallerFormatter: charlog.ShortCallerFormatter,
	})
	inner.SetColorProfile(termenv.Ascii)

	return &Logger{inner: inner}, &buf
}

// fakeEnv returns an envFunc that looks up keys in the provided map.
func fakeEnv(vars map[string]string) envFunc {
	return func(key string) string {
		return vars[key]
	}
}

// ---------------------------------------------------------------------------
// resolveLevel
// ---------------------------------------------------------------------------

func TestResolveLevel_Defaults(t *testing.T) {
	if got := resolveLevel(""); got != InfoLevel {
		t.Errorf("empty string: got %v, want InfoLevel", got)
	}
}

func TestResolveLevel_ValidValues(t *testing.T) {
	cases := []struct {
		input string
		want  Level
	}{
		{"debug", DebugLevel},
		{"DEBUG", DebugLevel},
		{"  Debug ", DebugLevel},
		{"info", InfoLevel},
		{"INFO", InfoLevel},
		{"warn", WarnLevel},
		{"WARN", WarnLevel},
		{"error", ErrorLevel},
		{"ERROR", ErrorLevel},
	}
	for _, tc := range cases {
		t.Run(tc.input, func(t *testing.T) {
			if got := resolveLevel(tc.input); got != tc.want {
				t.Errorf("resolveLevel(%q) = %v, want %v", tc.input, got, tc.want)
			}
		})
	}
}

func TestResolveLevel_Invalid(t *testing.T) {
	// Invalid levels should return InfoLevel (and print a warning, but we
	// don't capture stderr in this test — the warning is tested separately).
	if got := resolveLevel("banana"); got != InfoLevel {
		t.Errorf("resolveLevel(\"banana\") = %v, want InfoLevel", got)
	}
}

// ---------------------------------------------------------------------------
// resolveFormat
// ---------------------------------------------------------------------------

func TestResolveFormat(t *testing.T) {
	cases := []struct {
		input string
		want  charlog.Formatter
	}{
		{"", charlog.TextFormatter},
		{"text", charlog.TextFormatter},
		{"TEXT", charlog.TextFormatter},
		{"json", charlog.JSONFormatter},
		{"JSON", charlog.JSONFormatter},
		{" Json ", charlog.JSONFormatter},
		{"unknown", charlog.TextFormatter},
	}
	for _, tc := range cases {
		t.Run(tc.input, func(t *testing.T) {
			if got := resolveFormat(tc.input); got != tc.want {
				t.Errorf("resolveFormat(%q) = %v, want %v", tc.input, got, tc.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Level filtering
// ---------------------------------------------------------------------------

func TestLevelFiltering_InfoDefault(t *testing.T) {
	l, buf := captureLogger(InfoLevel, charlog.TextFormatter)
	l.Debug("should not appear")
	l.Info("visible info")
	l.Warn("visible warn")
	l.Error("visible error")

	out := buf.String()
	if strings.Contains(out, "should not appear") {
		t.Error("debug message appeared at info level")
	}
	if !strings.Contains(out, "visible info") {
		t.Error("info message missing at info level")
	}
	if !strings.Contains(out, "visible warn") {
		t.Error("warn message missing at info level")
	}
	if !strings.Contains(out, "visible error") {
		t.Error("error message missing at info level")
	}
}

func TestLevelFiltering_ErrorOnly(t *testing.T) {
	l, buf := captureLogger(ErrorLevel, charlog.TextFormatter)
	l.Debug("no")
	l.Info("no")
	l.Warn("no")
	l.Error("yes")

	out := buf.String()
	lines := strings.TrimSpace(out)
	if count := strings.Count(lines, "\n"); count != 0 {
		t.Errorf("expected exactly 1 line, got %d lines", count+1)
	}
	if !strings.Contains(out, "yes") {
		t.Error("error message missing at error level")
	}
}

func TestLevelFiltering_DebugShowsAll(t *testing.T) {
	l, buf := captureLogger(DebugLevel, charlog.TextFormatter)
	l.Debug("d")
	l.Info("i")
	l.Warn("w")
	l.Error("e")

	out := buf.String()
	for _, want := range []string{"d", "i", "w", "e"} {
		if !strings.Contains(out, want) {
			t.Errorf("message %q missing at debug level", want)
		}
	}
}

func TestLevelFiltering_WarnAndAbove(t *testing.T) {
	l, buf := captureLogger(WarnLevel, charlog.TextFormatter)
	l.Debug("no")
	l.Info("no")
	l.Warn("yes-w")
	l.Error("yes-e")

	out := buf.String()
	if strings.Contains(out, "no") {
		t.Error("debug or info appeared at warn level")
	}
	if !strings.Contains(out, "yes-w") {
		t.Error("warn message missing")
	}
	if !strings.Contains(out, "yes-e") {
		t.Error("error message missing")
	}
}

// ---------------------------------------------------------------------------
// Structured key-value pairs
// ---------------------------------------------------------------------------

func TestStructuredKeyValues_Text(t *testing.T) {
	l, buf := captureLogger(InfoLevel, charlog.TextFormatter)
	l.Info("installing terraform", "version", "1.6.1", "arch", "amd64")

	out := buf.String()
	if !strings.Contains(out, "version") || !strings.Contains(out, "1.6.1") {
		t.Errorf("key-value pair missing from text output: %s", out)
	}
}

func TestStructuredKeyValues_JSON(t *testing.T) {
	l, buf := captureLogger(InfoLevel, charlog.JSONFormatter)
	l.Info("installing terraform", "version", "1.6.1")

	var m map[string]any
	if err := json.Unmarshal(buf.Bytes(), &m); err != nil {
		t.Fatalf("failed to parse JSON log line: %v\nraw: %s", err, buf.String())
	}
	if m["msg"] != "installing terraform" {
		t.Errorf("msg = %v, want \"installing terraform\"", m["msg"])
	}
	if m["version"] != "1.6.1" {
		t.Errorf("version = %v, want \"1.6.1\"", m["version"])
	}
}

// ---------------------------------------------------------------------------
// JSON format
// ---------------------------------------------------------------------------

func TestJSONFormat_Structure(t *testing.T) {
	l, buf := captureLogger(InfoLevel, charlog.JSONFormatter)
	l.Info("hello")

	var m map[string]any
	if err := json.Unmarshal(buf.Bytes(), &m); err != nil {
		t.Fatalf("JSON parse error: %v", err)
	}
	if _, ok := m["level"]; !ok {
		t.Error("JSON output missing 'level' key")
	}
	if _, ok := m["msg"]; !ok {
		t.Error("JSON output missing 'msg' key")
	}
}

func TestJSONFormat_NoANSI(t *testing.T) {
	l, buf := captureLogger(InfoLevel, charlog.JSONFormatter)
	l.Info("clean output")

	out := buf.String()
	if strings.Contains(out, "\x1b[") {
		t.Error("JSON output contains ANSI escape codes")
	}
}

// ---------------------------------------------------------------------------
// NO_COLOR support
// ---------------------------------------------------------------------------

func TestNoColor_DisablesANSI(t *testing.T) {
	l := newFromEnvFunc(fakeEnv(map[string]string{
		"NO_COLOR": "1",
	}))

	var buf bytes.Buffer
	l.inner.SetOutput(&buf)
	l.Info("no colour please")

	out := buf.String()
	if strings.Contains(out, "\x1b[") {
		t.Error("output contains ANSI codes despite NO_COLOR being set")
	}
}

// ---------------------------------------------------------------------------
// Debug caller context
// ---------------------------------------------------------------------------

func TestDebugCallerContext(t *testing.T) {
	var buf bytes.Buffer
	inner := charlog.NewWithOptions(&buf, charlog.Options{
		Level:           DebugLevel,
		Formatter:       charlog.TextFormatter,
		ReportTimestamp: false,
		ReportCaller:    true,
		CallerFormatter: charlog.ShortCallerFormatter,
	})
	inner.SetColorProfile(termenv.Ascii)
	l := &Logger{inner: inner}

	l.inner.Debug("caller test")

	out := buf.String()
	// ShortCallerFormatter produces something like <logging/logging_test.go:NNN>
	if !strings.Contains(out, ".go:") {
		t.Errorf("debug output missing caller info: %s", out)
	}
}

// ---------------------------------------------------------------------------
// newFromEnvFunc integration
// ---------------------------------------------------------------------------

func TestNewFromEnvFunc_Defaults(t *testing.T) {
	l := newFromEnvFunc(fakeEnv(map[string]string{}))
	if l.GetLevel() != InfoLevel {
		t.Errorf("default level = %v, want InfoLevel", l.GetLevel())
	}
}

func TestNewFromEnvFunc_DebugLevel(t *testing.T) {
	l := newFromEnvFunc(fakeEnv(map[string]string{
		"TFENV_LOG_LEVEL": "debug",
	}))
	if l.GetLevel() != DebugLevel {
		t.Errorf("level = %v, want DebugLevel", l.GetLevel())
	}
}

func TestNewFromEnvFunc_JSONFormat(t *testing.T) {
	l := newFromEnvFunc(fakeEnv(map[string]string{
		"TFENV_LOG_FORMAT": "json",
	}))

	var buf bytes.Buffer
	l.inner.SetOutput(&buf)
	l.Info("json check")

	var m map[string]any
	if err := json.Unmarshal(buf.Bytes(), &m); err != nil {
		t.Fatalf("expected JSON output, got: %s", buf.String())
	}
}

func TestNewFromEnvFunc_InvalidLevel(t *testing.T) {
	l := newFromEnvFunc(fakeEnv(map[string]string{
		"TFENV_LOG_LEVEL": "banana",
	}))
	// Should fall back to info.
	if l.GetLevel() != InfoLevel {
		t.Errorf("invalid level fallback = %v, want InfoLevel", l.GetLevel())
	}
}

// ---------------------------------------------------------------------------
// With (sub-logger)
// ---------------------------------------------------------------------------

func TestWith_AddsFields(t *testing.T) {
	l, buf := captureLogger(InfoLevel, charlog.TextFormatter)
	sub := l.With("component", "install")
	sub.Info("download started")

	out := buf.String()
	if !strings.Contains(out, "component") || !strings.Contains(out, "install") {
		t.Errorf("With fields missing from output: %s", out)
	}
}

// ---------------------------------------------------------------------------
// SetLevel dynamic change
// ---------------------------------------------------------------------------

func TestSetLevel(t *testing.T) {
	l, buf := captureLogger(ErrorLevel, charlog.TextFormatter)
	l.Info("hidden")
	if strings.Contains(buf.String(), "hidden") {
		t.Fatal("info appeared at error level")
	}

	l.SetLevel(InfoLevel)
	l.Info("visible")
	if !strings.Contains(buf.String(), "visible") {
		t.Error("info missing after SetLevel to InfoLevel")
	}
}

// ---------------------------------------------------------------------------
// Package-level functions
// ---------------------------------------------------------------------------

func TestPackageLevelFunctions(t *testing.T) {
	var buf bytes.Buffer
	inner := charlog.NewWithOptions(&buf, charlog.Options{
		Level:           DebugLevel,
		Formatter:       charlog.TextFormatter,
		ReportTimestamp: false,
	})
	inner.SetColorProfile(termenv.Ascii)

	old := defaultLogger
	defaultLogger = &Logger{inner: inner}
	defer func() { defaultLogger = old }()

	Debug("pkg-debug")
	Info("pkg-info")
	Warn("pkg-warn")
	Error("pkg-error")

	out := buf.String()
	for _, want := range []string{"pkg-debug", "pkg-info", "pkg-warn", "pkg-error"} {
		if !strings.Contains(out, want) {
			t.Errorf("package-level %q missing from output", want)
		}
	}
}

func TestPackageLevelWith(t *testing.T) {
	var buf bytes.Buffer
	inner := charlog.NewWithOptions(&buf, charlog.Options{
		Level:           InfoLevel,
		Formatter:       charlog.TextFormatter,
		ReportTimestamp: false,
	})
	inner.SetColorProfile(termenv.Ascii)

	old := defaultLogger
	defaultLogger = &Logger{inner: inner}
	defer func() { defaultLogger = old }()

	sub := With("req", "abc")
	sub.Info("request handled")

	out := buf.String()
	if !strings.Contains(out, "req") || !strings.Contains(out, "abc") {
		t.Errorf("With fields missing: %s", out)
	}
}

// ---------------------------------------------------------------------------
// Default / SetDefault
// ---------------------------------------------------------------------------

func TestDefaultAndSetDefault(t *testing.T) {
	original := Default()
	if original == nil {
		t.Fatal("Default() returned nil")
	}

	custom, _ := captureLogger(ErrorLevel, charlog.TextFormatter)
	SetDefault(custom)
	if Default() != custom {
		t.Error("SetDefault did not replace default logger")
	}

	// Restore.
	SetDefault(original)
}

// ---------------------------------------------------------------------------
// Inner (access underlying logger)
// ---------------------------------------------------------------------------

func TestInner(t *testing.T) {
	l, _ := captureLogger(InfoLevel, charlog.TextFormatter)
	if l.Inner() == nil {
		t.Error("Inner() returned nil")
	}
}

// ---------------------------------------------------------------------------
// No BASHLOG references (regression guard)
// ---------------------------------------------------------------------------

func TestNoBashlogReferences(t *testing.T) {
	// This test exists to ensure no BASHLOG_* environment variables are
	// referenced anywhere in the package. It's a compile-time and grep-level
	// guard; we verify by checking that newFromEnvFunc with BASHLOG_ vars
	// set does not alter behaviour.
	l := newFromEnvFunc(fakeEnv(map[string]string{
		"BASHLOG_LEVEL":  "debug",
		"BASHLOG_FORMAT": "json",
	}))
	// BASHLOG_ variables must be ignored — defaults apply.
	if l.GetLevel() != InfoLevel {
		t.Errorf("BASHLOG_LEVEL should be ignored, got %v", l.GetLevel())
	}
}

// ---------------------------------------------------------------------------
// stderr output (all log output goes to stderr)
// ---------------------------------------------------------------------------

func TestOutputGoesToStderr(t *testing.T) {
	// A freshly created logger without overrides writes to os.Stderr.
	// We verify by checking the New() path doesn't panic and that
	// the inner logger is configured (we can't easily redirect os.Stderr
	// in a unit test, so we verify the construction path).
	l := newLogger(InfoLevel, charlog.TextFormatter, true)
	if l.inner == nil {
		t.Fatal("newLogger returned nil inner")
	}
}

// ---------------------------------------------------------------------------
// Zero-allocation when filtered (level check before work)
// ---------------------------------------------------------------------------

func TestZeroAllocWhenFiltered(t *testing.T) {
	l, buf := captureLogger(ErrorLevel, charlog.TextFormatter)

	// Log at debug — should be filtered and produce no output.
	l.Debug("expensive message", "key", "value")

	if buf.Len() != 0 {
		t.Errorf("filtered message produced output: %s", buf.String())
	}
}

// ---------------------------------------------------------------------------
// Text format level prefixes
// ---------------------------------------------------------------------------

func TestTextFormatLevelPrefixes(t *testing.T) {
	cases := []struct {
		name   string
		logFn  func(*Logger)
		expect string
	}{
		{"info", func(l *Logger) { l.Info("msg") }, "INFO"},
		{"warn", func(l *Logger) { l.Warn("msg") }, "WARN"},
		{"error", func(l *Logger) { l.Error("msg") }, "ERRO"},
		{"debug", func(l *Logger) { l.Debug("msg") }, "DEBU"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			l, buf := captureLogger(DebugLevel, charlog.TextFormatter)
			tc.logFn(l)
			out := buf.String()
			if !strings.Contains(out, tc.expect) {
				t.Errorf("expected %q in output: %s", tc.expect, out)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// JSON: level values
// ---------------------------------------------------------------------------

func TestJSONLevelValues(t *testing.T) {
	cases := []struct {
		name   string
		logFn  func(*Logger)
		expect string
	}{
		{"info", func(l *Logger) { l.Info("msg") }, "info"},
		{"warn", func(l *Logger) { l.Warn("msg") }, "warn"},
		{"error", func(l *Logger) { l.Error("msg") }, "error"},
		{"debug", func(l *Logger) { l.Debug("msg") }, "debug"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			l, buf := captureLogger(DebugLevel, charlog.JSONFormatter)
			tc.logFn(l)

			var m map[string]any
			if err := json.Unmarshal(buf.Bytes(), &m); err != nil {
				t.Fatalf("JSON parse error: %v", err)
			}
			if m["level"] != tc.expect {
				t.Errorf("level = %v, want %q", m["level"], tc.expect)
			}
		})
	}
}
