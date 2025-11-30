package logging

import (
	"log/slog"
	"os"
	"strings"

	"github.com/lmittmann/tint"
	"golang.org/x/term"
)

// NewLogger creates a new slog logger with JSON formatting
func NewLogger(level slog.Level) *slog.Logger {
	handler := slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
		Level: level,
	})
	return slog.New(handler)
}

// NewTextLogger creates a text-formatted logger (for CLI tools like migration)
func NewTextLogger(level slog.Level) *slog.Logger {
	handler := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: level,
	})
	return slog.New(handler)
}

// NewLoggerWithFormat creates a logger with specified format (json or text)
func NewLoggerWithFormat(level slog.Level, format string) *slog.Logger {
	format = strings.ToLower(strings.TrimSpace(format))

	var handler slog.Handler
	switch format {
	case "text":
		// Use colorized tint handler for text format
		// Auto-detect TTY for color support (disables colors when piped)
		handler = tint.NewHandler(os.Stderr, &tint.Options{
			Level:      level,
			TimeFormat: "[15:04:05]", // Bracketed 24-hour format with seconds
			NoColor:    !isTerminal(os.Stderr),
		})
	case "json", "": // default to JSON if empty or unrecognized
		handler = slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
			Level: level,
		})
	default:
		// Unknown format, default to JSON
		handler = slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
			Level: level,
		})
	}

	return slog.New(handler)
}

// isTerminal checks if the file descriptor is a terminal (for color detection)
func isTerminal(f *os.File) bool {
	return term.IsTerminal(int(f.Fd()))
}

// ParseLevel converts string log level to slog.Level
func ParseLevel(levelStr string) slog.Level {
	switch strings.ToLower(strings.TrimSpace(levelStr)) {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
