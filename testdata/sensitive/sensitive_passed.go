package testdata

import (
	"io"
	"log/slog"
	"time"

	"go.uber.org/zap"
)

// SensitivePassed logs only non-sensitive operational data.
func SensitivePassed() {
	// slog
	slogger := slog.New(slog.NewTextHandler(io.Discard, nil))
	slogger.Info("request handled",
		"request_id", "req-12345",
		"route", "/health",
		"status", 200,
		"duration_ms", 12,
	)

	// zap
	zlogger := zap.NewNop()
	defer func() { _ = zlogger.Sync() }()

	zlogger.Info("request handled",
		zap.String("request_id", "req-12345"),
		zap.String("route", "/health"),
		zap.Int("status", 200),
		zap.Duration("duration", 12*time.Millisecond),
	)
}
