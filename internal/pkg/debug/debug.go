package debug

import (
	"log/slog"
	"os"
	"strings"
)

var enabled bool

func Init() {
	level := strings.ToLower(os.Getenv("IROOM_DEBUG"))
	enabled = level == "1" || level == "true" || level == "yes"
	if enabled {
		slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		})))
		slog.Info("debug logging enabled")
	}
}

func IsEnabled() bool {
	return enabled
}

func Log(msg string, args ...any) {
	if enabled {
		slog.Debug(msg, args...)
	}
}
