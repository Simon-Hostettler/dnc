package main

import (
	"testing"

	"github.com/mstrYoda/go-arctest/pkg/arctest"
)

func TestArchitectureDependencies(t *testing.T) {
	arch, err := arctest.New("./")
	if err != nil {
		t.Fatalf("failed to create architecture: %v", err)
	}

	if err := arch.ParsePackages(); err != nil {
		t.Fatalf("failed to parse packages: %v", err)
	}

	module := `hostettler\.dev/dnc/`

	rules := []struct {
		name   string
		source string
		target string
	}{
		// command, util, db, models have no internal imports
		{"command has no internal imports", `^command$`, module},
		{"util has no internal imports", `^util$`, module},
		{"db has no internal imports", `^db$`, module},
		{"models has no internal imports", `^models$`, module},

		// repository must not import command or ui
		{"repository does not import command", `^repository$`, module + `command`},
		{"repository does not import ui", `^repository$`, module + `ui`},

		// no internal package imports the root package
		{"no package imports root", `.*`, `^hostettler\.dev/dnc$`},
	}

	for _, tt := range rules {
		t.Run(tt.name, func(t *testing.T) {
			rule, err := arch.DoesNotDependOn(tt.source, tt.target)
			if err != nil {
				t.Fatalf("failed to create rule: %v", err)
			}

			valid, violations := arch.ValidateDependenciesWithRules([]*arctest.DependencyRule{rule})
			if !valid {
				for _, v := range violations {
					t.Error(v)
				}
			}
		})
	}

	// all *Screen structs in ui/screen implement FocusableModel
	t.Run("screens implement FocusableModel", func(t *testing.T) {
		rule, err := arch.StructsImplementInterfaces(`Screen$`, `^FocusableModel$`)
		if err != nil {
			t.Fatalf("failed to create rule: %v", err)
		}

		valid, violations := arch.ValidateInterfaceImplementations([]*arctest.InterfaceImplementationRule{rule})
		if !valid {
			for _, v := range violations {
				t.Error(v)
			}
		}
	})
}
