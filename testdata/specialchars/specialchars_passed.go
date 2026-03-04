package specialchars

import (
	"log/slog"
	"os"

	"go.uber.org/zap"
)

func slogSpecialCharsPassed() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	// Correct log messages without special characters at start/end
	logger.Info("user logged in")
	logger.Error("failed to connect to database")
	logger.Warn("request timeout occurred")
	logger.Debug("processing request")

	// Messages with special characters in the middle are fine
	logger.Info("user email: test@example.com")
	logger.Error("path /api/v1/users not found")
	logger.Info("value is 100% correct")
}

func zapSpecialCharsPassed() {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	// Correct log messages without special characters at start/end
	logger.Info("user logged in")
	logger.Error("failed to connect to database")
	logger.Warn("request timeout occurred")
	logger.Debug("processing request")

	// Messages with special characters in the middle are fine
	logger.Info("user email: test@example.com")
	logger.Error("path /api/v1/users not found")
	logger.Info("value is 100% correct")
}
