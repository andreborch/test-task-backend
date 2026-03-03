# log-linter (loglint)

Статический анализатор для Go, который проверяет корректность вызовов логгеров.
Построен на `golang.org/x/tools/go/analysis` и интегрируется в `golangci-lint` через систему плагинов [`plugin-module-register`](https://github.com/golangci/plugin-module-register).

---

## Что проверяет

| Правило | Описание |
|---|---|
| `language` | Язык сообщений логгера должен соответствовать настройке `lang` (по умолчанию `en`) |
| `specialchars` | Запрещает спецсимволы в сообщениях (кроме `spec_char_exceptions`) |
| `lowercase` | Сообщения логгера должны начинаться со строчной буквы |
| `sensitive` | Запрещает ключевые слова из `sensitive_bans` (пароли, токены и т.д.) |

Поддерживаемые логгеры по умолчанию:
- `log/slog`
- `go.uber.org/zap`

Поддерживаемые функции по умолчанию: `Debug`, `Info`, `Warn`, `Error`, `Log`

---

## Требования

- **Go** 1.21+
- **golangci-lint** v2.10.1+ (для работы как плагин)

---

## Структура проекта

```
log-linter/
├── cmd/
│   └── loglint/
│       └── main.go          # standalone-бинарь (singlechecker)
├── internal/
│   ├── analyzers/           # AST-анализ
│   ├── rules/               # Правила: language, sensitive, specialchars, lowercase
│   └── utils/
│       ├── send_reports.go  # Отправка диагностик в analysis.Pass
│       └── is_logger.go     # Определение вызовов логгера
├── pkg/
│   ├── settings.go          # Структура LinterSettings и DefaultSettings()
│   ├── report.go            # Структура Report (Pos, Length, Message)
│   └── analyzer/
│       └── analyzer.go      # Плагин: New(), BuildAnalyzers(), run()
├── .custom-gcl.yml          # Конфиг сборки кастомного golangci-lint
├── .golangci.yml            # Конфиг запуска линтера
└── go.mod
```

---

## Установка и сборка

### Зависимости

```bash
go mod tidy
```

### Сборка standalone-бинаря

Standalone-режим позволяет запускать линтер напрямую через `go vet`.

**Windows:**
```powershell
go build -o loglint.exe ./cmd/loglint
```

**Linux / macOS:**
```bash
go build -o loglint ./cmd/loglint
```

### Запуск standalone-бинаря

```bash
# Проверить все пакеты
go vet -vettool=./loglint.exe ./...

# Проверить конкретный пакет
go vet -vettool=./loglint.exe ./internal/service/...
```

---

## Работа как плагин для golangci-lint

Линтер использует официальную систему плагинов golangci-lint v2 через
`github.com/golangci/plugin-module-register`. Это требует сборки **кастомного бинаря** golangci-lint с вшитым плагином.

### Шаг 1: Установите golangci-lint

```bash
go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.10.1
```

Проверьте:
```bash
golangci-lint --version
```

### Шаг 2: Конфиг сборки `.custom-gcl.yml`

Файл уже находится в корне проекта:

```yaml
# .custom-gcl.yml
version: v2.10.1
plugins:
  - module: 'github.com/andreborch/log-linter'
    import: 'github.com/andreborch/log-linter/pkg/analyzer'
    path: ./   # путь к локальному исходнику плагина
```

- `module` — Go-модуль плагина (из `go.mod` вашего линтера)
- `import` — пакет, в котором вызывается `register.Plugin(...)` (файл `pkg/analyzer/analyzer.go`)
- `path: ./` — собирается из локальных исходников; замените на версию тега (`version: v1.0.0`) при публикации

### Шаг 3: Соберите кастомный golangci-lint

```bash
golangci-lint custom -v
```

Флаг `-v` показывает подробный лог сборки. Бинарь появится в:
```
./bin/golangci-lint      # Linux/macOS
./bin/golangci-lint.exe  # Windows
```

> ⚠️ Пересобирайте бинарь каждый раз после изменений в коде линтера.

### Шаг 4: Конфиг запуска `.golangci.yml`

```yaml
# .golangci.yml
version: "2"

linters:
  default: none
  enable:
    - loglint   # имя из register.Plugin("loglint", New)

  settings:
    custom:
      loglint:
        type: module
        description: "Logging linter"
        original-url: github.com/andreborch/log-linter
        settings:
          # Список активных правил
          # Доступные: language, specialchars, lowercase, sensitive
          enabled_rules:
            - "language"
            - "specialchars"
            - "lowercase"
            - "sensitive"

          # Заблокированные слова (sensitive-правило)
          # Пример: ["password", "token", "secret"]
          sensitive_bans: []

          # Исключения для спецсимволов (specialchars-правило)
          # Символы, которые разрешены в сообщениях логгера
          spec_char_exceptions: ":_=-/"

          # Исключения для sensitive-правила
          # Слова, которые не считаются sensitive несмотря на sensitive_bans
          sens_exceptions: []

          # Пакеты, вызовы из которых считаются логгером
          logger_packages:
            - "log/slog"
            - "go.uber.org/zap"

          # Функции, вызовы которых считаются логгером
          logger_funcs:
            - "Debug"
            - "Info"
            - "Warn"
            - "Error"
            - "Log"

          # Ожидаемый язык сообщений (language-правило)
          lang: "en"
```

### Шаг 5: Запуск

```bash
# Запуск кастомного бинаря
./bin/golangci-lint run ./...

# Только loglint
./bin/golangci-lint run --enable-only loglint ./...

# Подробный вывод
./bin/golangci-lint run -v ./...

# Вывод в JSON (для CI/CD)
./bin/golangci-lint run --output.formats.text.path stdout --output.json.path report.json ./...
```

---

## Настройка правил

### Отключить отдельные правила

Уберите ненужные правила из `enabled_rules`:

```yaml
settings:
  enabled_rules:
    - "lowercase"    # только проверка регистра
    # - "language"   # отключено
    # - "sensitive"  # отключено
    # - "specialchars" # отключено
```

### Настройка sensitive-правила

```yaml
settings:
  enabled_rules:
    - "sensitive"
  sensitive_bans:
    - "password"
    - "passwd"
    - "token"
    - "secret"
    - "api_key"
    - "private_key"
  sens_exceptions:
    - "token_type"   # это слово не будет считаться sensitive
```

### Настройка specialchars-правила

```yaml
settings:
  enabled_rules:
    - "specialchars"
  spec_char_exceptions: ":_=-/"  # эти символы разрешены
```

### Добавить свой логгер

```yaml
settings:
  logger_packages:
    - "log/slog"
    - "go.uber.org/zap"
    - "github.com/sirupsen/logrus"   # добавить logrus
  logger_funcs:
    - "Debug"
    - "Info"
    - "Warn"
    - "Error"
    - "Log"
    - "WithField"                    # добавить метод logrus
```

---

## Интеграция с VS Code

В `.vscode/settings.json` вашего проекта:

```json
{
  "go.lintTool": "golangci-lint",
  "go.lintFlags": ["--config=${workspaceFolder}/.golangci.yml"],
  "go.lintOnSave": "workspace",
  "go.alternateTools": {
    "golangci-lint": "${workspaceFolder}/bin/golangci-lint"
  }
}
```

> `go.alternateTools` указывает VS Code использовать кастомный бинарь вместо системного golangci-lint.

---

## Интеграция с CI/CD

### GitHub Actions

```yaml
# .github/workflows/lint.yml
name: Lint

on:
  push:
    branches: [main]
  pull_request:
    branches: [main]

jobs:
  lint:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - uses: actions/setup-go@v5
        with:
          go-version: '1.21'

      - name: Install golangci-lint
        run: go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.10.1

      - name: Build custom golangci-lint with loglint plugin
        run: golangci-lint custom -v

      - name: Run loglint
        run: ./bin/golangci-lint run ./...
```

### Makefile

```makefile
LINT_BIN := ./bin/golangci-lint

.PHONY: build-lint
build-lint:
    golangci-lint custom -v

.PHONY: lint
lint: build-lint
    $(LINT_BIN) run ./...

.PHONY: lint-only
lint-only:
    $(LINT_BIN) run --enable-only loglint ./...

.PHONY: test
test:
    go test -v ./...

.PHONY: check
check: test lint

.PHONY: clean
clean:
    rm -f $(LINT_BIN) loglint.exe loglint
```

---

## Тесты

```bash
go test ./...
go test -v ./...
go test -cover ./...
```

---

## Как устроено

```
вызов логгера в коде
        │
        ▼
  utils.IsLogger()          # определяет, является ли вызов логгером
        │                   # по logger_packages и logger_funcs из настроек
        ▼
   rules.LangIsCorrect()    # правило language
   rules.HasSensitiveData() # правило sensitive
   rules.HasSpecialChar()   # правило specialchars
   rules.CheckLowerCase()   # правило lowercase
        │
        ▼
  []pkg.Report{Pos, Length, Message}
        │
        ▼
  utils.SendReports()       # Report → analysis.Diagnostic{Pos, End}
        │                   # End = Pos + Length (подчёркивание в IDE)
        ▼
  analysis.Pass.Report()    # диагностика уходит в golangci-lint / go vet / IDE
```

---

## Типичные проблемы

| Проблема | Решение |
|---|---|
| `loglint: not found` | Убедитесь, что в `.golangci.yml` имя линтера `loglint` (из `register.Plugin("loglint", ...)`) |
| Кастомный бинарь не работает | Пересоберите: `golangci-lint custom -v` |
| IDE не подсвечивает | Проверьте `go.alternateTools` в `settings.json` |
| `plugin: not registered` | Убедитесь, что `import` в `.custom-gcl.yml` указывает на пакет с `func init()` |
| Правила не применяются | Проверьте, что правило добавлено в `enabled_rules` в `.golangci.yml` |
| `sensitive_bans: []` не работает | Список пуст — добавьте слова: `["password", "token"]` |

---

## Лицензия

MIT
