package pkg

type LinterSettings struct {
	EnabledRules        []string `json:"enabled_rules"`
	BlockedSensitive    []string `json:"sensitive_bans"`
	SpecCharsExceptions string   `json:"spec_char_exceptions"`
	SensitiveExceptions []string `json:"sens_exceptions"`
	LoggerPackages      []string `json:"logger_packages"`
	LoggerFunctions     []string `json:"logger_funcs"`
	Language            string   `json:"lang"`
}

func DefaultSettings() LinterSettings {
	return LinterSettings{
		EnabledRules: []string{
			"language",
			"specialchars",
			"lowercase",
			"sensitive",
		},
		BlockedSensitive:    []string{},
		SpecCharsExceptions: ":_=-/%@",
		SensitiveExceptions: []string{},
		LoggerPackages: []string{
			"log/slog",
			"go.uber.org/zap",
		},
		LoggerFunctions: []string{
			"Debug",
			"Info",
			"Warn",
			"Error",
			"Log",
		},
		Language: "en",
	}
}
