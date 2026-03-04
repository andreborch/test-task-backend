package lowercase

import (
	"log/slog"

	"go.uber.org/zap"
)

func slogNonPassed() {
	slog.Info("This Is Not Lowercase Message")
	slog.Debug("Another Bad Message With Uppercase")
	slog.Warn("Warning Message Not Lowercase")
	slog.Error("Error Message Not Lowercase")

	logger := slog.Default()
	logger.Info("Bad Log Message Here")
	logger.Debug("Another Bad One")
}

func zapNonPassed() {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	logger.Info("This Is Not Lowercase Message")
	logger.Debug("Another Bad Message With Uppercase")
	logger.Warn("Warning Message Not Lowercase")
	logger.Error("Error Message Not Lowercase")

	sugar := logger.Sugar()
	sugar.Infow("Bad Sugar Message Here")
	sugar.Debugw("Another Bad Sugar Message")
}
