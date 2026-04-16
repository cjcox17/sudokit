package logging

import (
	"io"
	"log/slog"
	"os"

	"github.com/lmittmann/tint"
)

// InitLogger initializes the slog logger with a colored, formatted handler
// This should be called early in the application startup, before any other logging
func InitLogger() {
	// Check if NO_COLOR environment variable is set
	noColorEnv := os.Getenv("NO_COLOR")

	// Determine if we should use colors
	// Use colors unless NO_COLOR is explicitly set to a non-empty value
	useColors := noColorEnv == ""

	// Create a tint handler with colored output
	// The tint handler will use colors if NoColor is false AND the output is a TTY
	// To force colors in non-TTY environments, we need to use a custom approach
	handler := tint.NewHandler(os.Stderr, &tint.Options{
		Level:      slog.LevelDebug,
		TimeFormat: "15:04:05",
		NoColor:    !useColors, // Set to true if NO_COLOR is set, false otherwise
		AddSource:  true,       // Add source file information
	})

	// Set the default logger
	slog.SetDefault(slog.New(handler))
}

// GetLogger returns the default slog logger
func GetLogger() *slog.Logger {
	return slog.Default()
}

// NewLoggerWithWriter creates a new logger with a custom writer
// Useful for testing or redirecting logs to different outputs
func NewLoggerWithWriter(w io.Writer) *slog.Logger {
	// Check if NO_COLOR environment variable is set
	noColorEnv := os.Getenv("NO_COLOR")
	useColors := noColorEnv == ""

	handler := tint.NewHandler(w, &tint.Options{
		Level:      slog.LevelDebug,
		TimeFormat: "15:04:05",
		NoColor:    !useColors,
		AddSource:  true,
	})
	return slog.New(handler)
}
