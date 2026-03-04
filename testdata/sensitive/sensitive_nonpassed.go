package testdata

import (
	"log/slog"

	"go.uber.org/zap"
)

func sensitiveNonPassed() {
	// slog examples - should NOT pass sensitive log check
	slog.Info("user logged in", "password", "secret123")
	slog.Debug("auth token", "token", "eyJhbGciOiJIUzI1NiJ9")
	slog.Warn("user data", "credit_card", "4111111111111111")
	slog.Error("failed login", "secret", "mysecretvalue")
	slog.Info("request", "api_key", "sk-1234567890abcdef")
	slog.Info("user info", "ssn", "123-45-6789")
	slog.Debug("credentials", "private_key", "-----BEGIN RSA PRIVATE KEY-----")

	// zap examples - should NOT pass sensitive log check
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	logger.Info("user logged in", zap.String("password", "secret123"))
	logger.Debug("auth token", zap.String("token", "eyJhbGciOiJIUzI1NiJ9"))
	logger.Warn("user data", zap.String("credit_card", "4111111111111111"))
	logger.Error("failed login", zap.String("secret", "mysecretvalue"))
	logger.Info("request", zap.String("api_key", "sk-1234567890abcdef"))
	logger.Info("user info", zap.String("ssn", "123-45-6789"))
	logger.Debug("credentials", zap.String("private_key", "-----BEGIN RSA PRIVATE KEY-----"))
}
