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
| `sensitive` | Запрещает ключевые слова и паттерны токенов в сообщениях логгера |

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
│       └── main.go              # standalone-бинарь (singlechecker)
├── internal/
│   ├── rules/                   # Правила: language, sensitive, specialchars, lowercase
│   └── utils/
│       ├── send_reports.go      # Отправка диагностик в analysis.Pass
│       └── is_logger.go         # Определение вызовов логгера
├── pkg/
│   ├── settings.go              # Структура LinterSettings и DefaultSettings()
│   ├── report.go                # Структура Report (Pos, Length, Message)
│   ├── sensitive_def.go         # Дефолтные бан-листы слов и regex-паттернов токенов
│   └── analyzer/
│       └── analyzer.go          # Плагин: New(), BuildAnalyzers(), run()
├── .custom-gcl.yml              # Конфиг сборки кастомного golangci-lint
├── .golangci.yml                # Конфиг запуска линтера
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

| Поле | Описание |
|---|---|
| `module` | Go-модуль плагина (из `go.mod`) |
| `import` | Пакет с `func init()` где вызывается `register.Plugin(...)` |
| `path: ./` | Сборка из локальных исходников. При публикации замените на `version: v1.0.0` |

### Шаг 3: Соберите кастомный golangci-lint

```bash
make build-lint
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

          # Дополнительные заблокированные слова (sensitive-правило)
          # Дефолтный бан-лист встроен и всегда активен (см. раздел ниже)
          sensitive_bans: []

          # Исключения для спецсимволов (specialchars-правило)
          spec_char_exceptions: ":_=-/@%"

          # Исключения для sensitive-правила
          # Слова из этого списка не будут считаться sensitive
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
./custom-gcl run ./...

# Только loglint
./custom-gcl run --enable-only loglint ./...

# Подробный вывод
./custom-gcl run -v ./...

# Вывод в JSON (для CI/CD)
./custom-gcl run --output.formats.text.path stdout --output.json.path report.json ./...
```

---

## Настройка правил

### Отключить отдельные правила

```yaml
settings:
  enabled_rules:
    - "lowercase"      # только проверка регистра
    # - "language"     # отключено
    # - "sensitive"    # отключено
    # - "specialchars" # отключено
```

### Настройка sensitive-правила

Правило `sensitive` работает в два этапа:

1. **Дефолтный бан-лист** — всегда активен, не требует настройки (см. раздел ниже)
2. **Кастомный бан-лист** — задаётся через `sensitive_bans` в `.golangci.yml`

```yaml
settings:
  enabled_rules:
    - "sensitive"
  # Дополнительные слова поверх дефолтного бан-листа
  sensitive_bans:
    - "internal_key"
    - "db_password"
  # Исключения — слова, которые НЕ считаются sensitive
  sens_exceptions:
    - "token_type"
    - "session_name"
```

### Настройка specialchars-правила

```yaml
settings:
  enabled_rules:
    - "specialchars"
  # Символы, разрешённые в сообщениях логгера
  spec_char_exceptions: ":_=-/"
```

### Добавить свой логгер

```yaml
settings:
  logger_packages:
    - "log/slog"
    - "go.uber.org/zap"
    - "github.com/sirupsen/logrus"
  logger_funcs:
    - "Debug"
    - "Info"
    - "Warn"
    - "Error"
    - "Log"
    - "WithField"
```

---

## Дефолтные значения sensitive-правила

Правило `sensitive` содержит два встроенных списка, которые **всегда активны**, за исключением добавления в исключения.

### Дефолтный бан-лист слов (`DefaultSensBans`)

Следующие слова запрещены в сообщениях логгера, проверяются по наличию ключевого слова в конце строки или ключевое слово + ":", "=", "is", иначе проблем не возникнет:

| Категория | Слова |
|---|---|
| **Пароли** | `password`, `passwd`, `pwd`, `pass` |
| **PIN / OTP** | `pin`, `pincode`, `pin_code`, `otp`, `one_time_password`, `mfa`, `totp`, `2fa` |
| **Секреты** | `secret`, `token` |
| **API ключи** | `api_key`, `apikey`, `api-key`, `api_token`, `apitoken`, `api-token` |
| **Токены доступа** | `access_token`, `access_key`, `auth_token`, `auth_key` |
| **Учётные данные** | `credentials` |
| **Банковские данные** | `credit_card`, `creditcard`, `card_number`, `cardnumber`, `cvv`, `cvc` |
| **Персональные данные** | `ssn`, `social_security` |
| **Криптография** | `private_key`, `privatekey`, `secret_key`, `secretkey`, `encryption_key` |
| **Авторизация** | `bearer`, `authorization` |
| **Сессии** | `session_id`, `sessionid`, `cookie` |
| **JWT / OAuth** | `jwt`, `oauth_token`, `refresh_token` |
| **Клиентские данные** | `client_secret`, `client_id` |
| **AWS** | `aws_access_key`, `aws_secret_key`, `aws_session_token` |

### Дефолтные паттерны токенов (`DefaultTokensPatterns`)

Помимо слов, линтер также проверяет сообщения на наличие реальных секретов по regex-паттернам:

| Сервис / Тип | Паттерн |
|---|---|
| **GitHub токены** | `ghp_`, `gho_`, `ghu_`, `ghs_`, `ghr_` + 36 символов |
| **GitLab токены** | `glpat-` + 20+ символов |
| **AWS Access Key** | `AKIA`, `ABIA`, `ACCA`, `ASIA` + 16 символов |
| **OpenAI API key** | `sk-` + 48 символов |
| **Stripe keys** | `sk_live_`, `pk_live_`, `rk_live_` + 24+ символов |
| **Square tokens** | `sq0csp-`, `sq0atp-` |
| **Slack tokens** | `xoxb-`, `xoxp-`, `xoxa-`, `xoxr-` |
| **SendGrid API key** | `SG.` + формат |
| **JWT токены** | `eyJhbGciOi...` |
| **Bearer токены** | `Bearer <token>` |
| **Hex-секреты** | `secret=<32-64 hex символа>` |
| **Base64-секреты** | `token=<40+ base64 символа>` |
| **Heroku API key** | UUID формат |
| **Google API key** | `AIza` + 35 символов |
| **Google OAuth** | `ya29.` |
| **Twilio API key** | `SK` + 32 hex символа |
| **Mailgun API key** | `key-` + 32 символа |
| **NPM токен** | `npm_` + 36 символов |
| **PyPI токен** | `pypi-` + 100+ символов |
| **Telegram Bot** | `<id>:<35 символов>` |
| **Discord Bot** | Специфичный формат |
| **Azure Storage Key** | `DefaultEndpointsProtocol=https;...` |
| **RSA приватный ключ** | `-----BEGIN RSA PRIVATE KEY-----` |
| **SSH приватный ключ** | `-----BEGIN OPENSSH PRIVATE KEY-----` |
| **PGP приватный ключ** | `-----BEGIN PGP PRIVATE KEY BLOCK-----` |

> Если в сообщении логгера обнаружен любой из этих паттернов — это нарушение правила `sensitive`.

### Пример срабатывания

```go
// ❌ Нарушение — слово "password" в сообщении
slog.Info("user password updated")

// ❌ Нарушение — реальный GitHub токен в сообщении
slog.Debug("token: ghp_A1b2C3d4E5f6G7h8I9j0K1l2M3n4O5p6Q7r8")

// ✅ Корректно
slog.Info("user credentials updated successfully")
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
  utils.IsLogger()            # определяет вызов логгера
        │                     # по logger_packages и logger_funcs
        ▼
   rules.LangIsCorrect()      # правило language
   rules.HasSensitiveData()   # правило sensitive
        │                     # ├─ DefaultSensBans() — встроенный бан-лист слов
        │                     # ├─ DefaultTokensPatterns() — regex паттерны токенов
        │                     # └─ sensitive_bans из настроек — кастомные слова
   rules.HasSpecialChar()     # правило specialchars
   rules.CheckLowerCase()     # правило lowercase
        │
        ▼
  []pkg.Report{Pos, Length, Message}
        │
        ▼
  utils.SendReports()         # Report → analysis.Diagnostic{Pos, End}
        │                     # End = Pos + Length (подчёркивание в IDE)
        ▼
  analysis.Pass.Report()      # диагностика → golangci-lint / go vet / IDE
```

---

## Типичные проблемы

| Проблема | Решение |
|---|---|
| `loglint: not found` | Убедитесь, что имя линтера в `.golangci.yml` — `loglint` (из `register.Plugin("loglint", ...)`) |
| Кастомный бинарь не работает | Пересоберите: `golangci-lint custom -v` |
| IDE не подсвечивает | Проверьте `go.alternateTools` в `settings.json` |
| `plugin: not registered` | Убедитесь, что `import` в `.custom-gcl.yml` указывает на пакет с `func init()` |
| Правила не применяются | Проверьте, что правило добавлено в `enabled_rules` |
| Ложные срабатывания sensitive | Добавьте слово в `sens_exceptions` |
| Хочу добавить своё sensitive-слово | Добавьте в `sensitive_bans` в `.golangci.yml` |

---

## Лицензия

MIT
