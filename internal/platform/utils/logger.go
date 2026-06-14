package utils

import (
	"log/slog"
	"os"
)

// logger is the process-wide structured logger. It defaults to human-readable
// text on stderr so output is sane before Init is called (e.g. in tests).
var logger = slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo}))

// InitLogger configures the global logger for the given environment. Production
// emits JSON (machine-parseable for log aggregation); everything else emits
// readable text. Call once at startup.
func InitLogger(prod bool, level slog.Level) {
	opts := &slog.HandlerOptions{Level: level}
	if prod {
		logger = slog.New(slog.NewJSONHandler(os.Stdout, opts))
	} else {
		logger = slog.New(slog.NewTextHandler(os.Stderr, opts))
	}
	slog.SetDefault(logger)
}

// Logger returns the configured structured logger for callers that want to add
// their own structured attributes.
func Logger() *slog.Logger { return logger }

// The Log* helpers below preserve the original call sites across the codebase
// while routing through structured logging. Prefer Logger() with explicit
// attributes in new code.

func LogInfo(msg string) {
	logger.Info(msg)
}

func LogInfoF(msg string, value any) {
	logger.Info(msg, "value", value)
}

func LogError(msg string, err error) {
	logger.Error(msg, "error", err)
}

func LogErrorf(msg string) {
	logger.Error(msg)
}
