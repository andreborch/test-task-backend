package language

import (
	"context"
	"log/slog"

	"go.uber.org/zap"
)

func slogLanguagePassed() {
	slog.Info("user created successfully")
	slog.Error("failed to connect to database")
	slog.Warn("retry attempt exceeded")
	slog.Debug("processing request")

	slog.InfoContext(context.Background(), "server started")
	slog.ErrorContext(context.Background(), "unexpected error occurred")
}

func zapLanguagePassed() {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	logger.Info("user created successfully")
	logger.Error("failed to connect to database")
	logger.Warn("retry attempt exceeded")
	logger.Debug("processing request")

	sugar := logger.Sugar()
	sugar.Info("server started")
	sugar.Errorf("unexpected error occurred: %v", "some error")
	sugar.Warnw("connection lost", "retries", 3)
}
