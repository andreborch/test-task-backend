package analyzer

import (
	"go/ast"

	"github.com/andreborch/log-linter/internal/rules"
	"github.com/andreborch/log-linter/internal/utils"
	"github.com/andreborch/log-linter/pkg"
	"github.com/golangci/plugin-module-register/register"
	"golang.org/x/tools/go/analysis"
)

func init() {
	register.Plugin("loglint", New)
}

type LogPlugin struct {
	settings pkg.LinterSettings
}

func New(settings any) (register.LinterPlugin, error) {
	s, err := register.DecodeSettings[pkg.LinterSettings](settings)
	if err != nil {
		s = pkg.DefaultSettings()
	}

	return &LogPlugin{settings: s}, nil
}

// BuildAnalyzers returns the analyzers exposed by this plugin.
// The host (golangci-lint) uses this list to execute custom checks.
func (plug *LogPlugin) BuildAnalyzers() ([]*analysis.Analyzer, error) {
	return []*analysis.Analyzer{
		{
			// Name is the analyzer identifier shown by tooling.
			Name: "LogLint",
			// Doc is a short analyzer description.
			Doc: "Checks log calls for language, sensitive data, special chars, and casing.",
			// Run points to the analyzer execution function.
			Run: plug.run,
		},
	}, nil
}

func (plug *LogPlugin) GetLoadMode() string {
	return register.LoadModeTypesInfo
}

// run executes the log analyzer plugin for all files in the current analysis pass.
//
// It traverses each AST node, detects logger function calls configured in plugin settings,
// and applies validation rules to logger arguments, including language compliance,
// sensitive data checks, special character restrictions, and lowercase formatting.
//
// Any violations found are collected as reports and emitted through the analysis pass.
// The function returns no analysis result and no error under normal execution.
func (plug *LogPlugin) run(pass *analysis.Pass) (any, error) {
	settings := plug.settings
	for _, file := range pass.Files {

		ast.Inspect(file, func(n ast.Node) bool {
			if callExpr, ok := n.(*ast.CallExpr); ok {

				is_logger, args := utils.IsLogger(callExpr, pass.TypesInfo, settings.LoggerPackages, settings.LoggerFunctions)
				if !is_logger {
					return true
				}

				reports := []pkg.Report{}

				rules.LangIsCorrect(args, &reports, settings.Language)
				rules.HasSensitiveData(args, &reports, settings.BlockedSensitive, settings.SensitiveExceptions)
				rules.HasSpecialChar(args, &reports, settings.SpecCharsExceptions)
				rules.CheckLowerCase(args, &reports)
				utils.SendReports(&reports, pass, "log")
			}
			return true
		})
	}

	return nil, nil
}
