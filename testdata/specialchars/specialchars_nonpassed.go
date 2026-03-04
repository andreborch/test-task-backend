package specialchars

import (
	"log/slog"
	"os"

	"go.uber.org/zap"
)

func slogSpecialCharsExamples() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	// Should NOT pass: special characters in message
	slog.Info("user login@ failed!")
	slog.Warn("request #123 processed?")
	slog.Error("file not found: /etc/config.yaml!")
	slog.Debug("value = 100%")

	// Should NOT pass: special characters in key
	logger.Info("some message", "user@name", "john")
	logger.Info("some message", "request#id", "456")
	logger.Warn("some message", "file/path", "/etc/config")
	logger.Error("some message", "error!", "something went wrong")

	// Should NOT pass: special characters in message with format
	slog.Info("processing item [1] of {10}")
	slog.Warn("status: 200 OK!")
	slog.Error("unexpected char: & found")
}

func zapSpecialCharsExamples() {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	sugar := logger.Sugar()

	// Should NOT pass: special characters in message
	logger.Info("user login@ failed!")
	logger.Warn("request #123 processed?")
	logger.Error("file not found: /etc/config.yaml!")

	// Should NOT pass: special characters in key
	logger.Info("some message", zap.String("user@name", "john"))
	logger.Info("some message", zap.String("request#id", "456"))
	logger.Warn("some message", zap.String("file/path", "/etc/config"))

	// Should NOT pass: special characters in sugar message
	sugar.Infow("processing item [1] of {10}")
	sugar.Warnw("status: 200 OK!")
	sugar.Errorw("unexpected char: & found")

	// Should NOT pass: special characters in sugar key
	sugar.Infow("some message", "user@name", "john")
	sugar.Warnw("some message", "request#id", "456")
}
