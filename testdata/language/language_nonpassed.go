package language

import (
	"log/slog"

	"go.uber.org/zap"
)

func LanguageFailed() {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	// zap logs with non-English messages - should fail language check
	logger.Info("привет мир")
	logger.Error("ошибка подключения к базе данных")
	logger.Warn("предупреждение: низкий уровень памяти")
	logger.Debug("отладочное сообщение")

	// slog logs with non-English messages - should fail language check
	slog.Info("сервер запущен")
	slog.Error("не удалось открыть файл")
	slog.Warn("истекает срок действия токена")
	slog.Debug("обработка запроса")

	// mixed language - should also fail
	logger.Info("server started успешно")
	slog.Info("connection failed неожиданно")
}
