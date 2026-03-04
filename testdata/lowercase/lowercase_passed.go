package lowercase

import (
	"log/slog"
	"os"

	"go.uber.org/zap"
)

func slogExamples() {
	// slog examples - all messages are lowercase, should pass
	slog.Info("server started successfully")
	slog.Debug("processing request")
	slog.Warn("cache miss detected")
	slog.Error("failed to connect to database")

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	logger.Info("user logged in")
	logger.Debug("fetching data from cache")
	logger.Warn("rate limit approaching")
	logger.Error("timeout occurred")

	logger.Info("request completed",
		slog.String("method", "GET"),
		slog.Int("status", 200),
	)
}

func zapExamples() {
	// zap examples - all messages are lowercase, should pass
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	logger.Info("application started")
	logger.Debug("processing item")
	logger.Warn("disk space low")
	logger.Error("failed to write file")

	logger.Info("request handled",
		zap.String("path", "/api/v1"),
		zap.Int("status", 200),
	)

	sugar := logger.Sugar()
	sugar.Info("sugar logger initialized")
	sugar.Debugf("processing record %d", 42)
	sugar.Warnw("connection unstable", "retries", 3)
	sugar.Errorf("failed to parse config: %s", "invalid format")
}
