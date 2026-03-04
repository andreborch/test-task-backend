package analyzer

import (
	"go/ast"
	"go/token"
	"reflect"
	"strconv"
	"testing"

	"github.com/andreborch/log-linter/internal/rules"
	"github.com/andreborch/log-linter/pkg"
	"golang.org/x/tools/go/analysis"
)

func TestNew_WithValidSettings_Decodes(t *testing.T) {
	input := map[string]any{
		"enabled_rules":        []string{"language", "lowercase"},
		"sensitive_bans":       []string{"secret"},
		"spec_char_exceptions": ":",
		"sens_exceptions":      []string{"password:"},
		"logger_packages":      []string{"log/slog"},
		"logger_funcs":         []string{"Info"},
		"lang":                 "en",
	}

	lp, err := New(input)
	if err != nil {
		t.Fatalf("New() returned unexpected error: %v", err)
	}

	plug, ok := lp.(*LogPlugin)
	if !ok {
		t.Fatalf("New() returned unexpected type: %T", lp)
	}

	want := pkg.LinterSettings{
		EnabledRules:        []string{"language", "lowercase"},
		BlockedSensitive:    []string{"secret"},
		SpecCharsExceptions: ":",
		SensitiveExceptions: []string{"password:"},
		LoggerPackages:      []string{"log/slog"},
		LoggerFunctions:     []string{"Info"},
		Language:            "en",
	}

	if !reflect.DeepEqual(plug.settings, want) {
		t.Fatalf("decoded settings mismatch:\nwant: %#v\ngot:  %#v", want, plug.settings)
	}
}

func TestNew_WithInvalidSettings_FallsBackToDefaults(t *testing.T) {
	// Intentionally invalid shape for DecodeSettings.
	lp, err := New(12345)
	if err != nil {
		t.Fatalf("New() returned unexpected error: %v", err)
	}

	plug, ok := lp.(*LogPlugin)
	if !ok {
		t.Fatalf("New() returned unexpected type: %T", lp)
	}

	want := pkg.DefaultSettings()
	if !reflect.DeepEqual(plug.settings, want) {
		t.Fatalf("default settings mismatch:\nwant: %#v\ngot:  %#v", want, plug.settings)
	}
}

func TestLogPlugin_BuildAnalyzers(t *testing.T) {
	plug := &LogPlugin{settings: pkg.DefaultSettings()}

	analyzers, err := plug.BuildAnalyzers()
	if err != nil {
		t.Fatalf("BuildAnalyzers() returned unexpected error: %v", err)
	}
	if len(analyzers) != 1 {
		t.Fatalf("BuildAnalyzers() analyzers count = %d, want 1", len(analyzers))
	}

	a := analyzers[0]
	if a == nil {
		t.Fatal("BuildAnalyzers() returned nil analyzer")
	}
	if a.Name != "LogLint" {
		t.Fatalf("analyzer Name = %q, want %q", a.Name, "LogLint")
	}
	if a.Doc == "" {
		t.Fatal("analyzer Doc is empty")
	}
	if a.Run == nil {
		t.Fatal("analyzer Run is nil")
	}
}

func TestRules_LangIsCorrect(t *testing.T) {
	t.Run("reports non-english text when lang=en", func(t *testing.T) {
		args := []ast.Expr{
			&ast.BasicLit{Kind: token.STRING, Value: strconv.Quote("Привет мир")},
		}
		reports := []pkg.Report{}

		rules.LangIsCorrect(args, &reports, "en")

		if len(reports) == 0 {
			t.Fatal("LangIsCorrect() expected at least one report for non-english text")
		}
	})

	t.Run("does not report english text when lang=en", func(t *testing.T) {
		args := []ast.Expr{
			&ast.BasicLit{Kind: token.STRING, Value: strconv.Quote("hello world")},
		}
		reports := []pkg.Report{}

		rules.LangIsCorrect(args, &reports, "en")

		if len(reports) != 0 {
			t.Fatalf("LangIsCorrect() expected no reports, got %d", len(reports))
		}
	})
}

func TestRules_HasSensitiveData(t *testing.T) {
	t.Run("reports banned sensitive token", func(t *testing.T) {
		args := []ast.Expr{
			&ast.BasicLit{Kind: token.STRING, Value: strconv.Quote("user secret leaked")},
		}
		reports := []pkg.Report{}

		rules.HasSensitiveData(args, &reports, []string{"secret"}, nil)

		if len(reports) == 0 {
			t.Fatal("HasSensitiveData() expected at least one report")
		}
	})

	t.Run("does not report when token is in exceptions", func(t *testing.T) {
		args := []ast.Expr{
			&ast.BasicLit{Kind: token.STRING, Value: strconv.Quote("password: *****")},
		}
		reports := []pkg.Report{}

		rules.HasSensitiveData(args, &reports, []string{"password"}, []string{"password:"})

		if len(reports) != 0 {
			t.Fatalf("HasSensitiveData() expected no reports, got %d", len(reports))
		}
	})
}

func TestRules_HasSpecialChar(t *testing.T) {
	t.Run("reports forbidden special character", func(t *testing.T) {
		args := []ast.Expr{
			&ast.BasicLit{Kind: token.STRING, Value: strconv.Quote("email@test")},
		}
		reports := []pkg.Report{}

		rules.HasSpecialChar(args, &reports, ":")

		if len(reports) == 0 {
			t.Fatal("HasSpecialChar() expected at least one report")
		}
	})

	t.Run("does not report allowed exception character", func(t *testing.T) {
		args := []ast.Expr{
			&ast.BasicLit{Kind: token.STRING, Value: strconv.Quote("status: ok")},
		}
		reports := []pkg.Report{}

		rules.HasSpecialChar(args, &reports, ":")

		if len(reports) != 0 {
			t.Fatalf("HasSpecialChar() expected no reports, got %d", len(reports))
		}
	})
}

func TestRules_CheckLowerCase(t *testing.T) {
	t.Run("reports non-lowercase message", func(t *testing.T) {
		args := []ast.Expr{
			&ast.BasicLit{Kind: token.STRING, Value: strconv.Quote("Hello world")},
		}
		reports := []pkg.Report{}

		rules.CheckLowerCase(args, &reports)

		if len(reports) == 0 {
			t.Fatal("CheckLowerCase() expected at least one report")
		}
	})

	t.Run("does not report lowercase message", func(t *testing.T) {
		args := []ast.Expr{
			&ast.BasicLit{Kind: token.STRING, Value: strconv.Quote("hello world")},
		}
		reports := []pkg.Report{}

		rules.CheckLowerCase(args, &reports)

		if len(reports) != 0 {
			t.Fatalf("CheckLowerCase() expected no reports, got %d", len(reports))
		}
	})
}

func TestLogPlugin_run_EmptyPass_NoError(t *testing.T) {
	plug := &LogPlugin{settings: pkg.DefaultSettings()}

	got, err := plug.run(&analysis.Pass{})
	if err != nil {
		t.Fatalf("run() returned unexpected error: %v", err)
	}
	if got != nil {
		t.Fatalf("run() = %#v, want nil", got)
	}
}
