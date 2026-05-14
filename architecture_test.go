package main

import (
	"go/types"
	"strings"
	"testing"

	"github.com/mstrYoda/go-arctest/pkg/arctest"
	"golang.org/x/tools/go/packages"
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

	// all *Screen structs in ui/screen implement FocusableModel.
	//
	// arctest's own StructsImplementInterfaces is unusable here: it only sees
	// methods with an explicit receiver and ignores method sets promoted from embedded structs
	t.Run("screens implement FocusableModel", func(t *testing.T) {
		cfg := &packages.Config{Mode: packages.NeedTypes | packages.NeedName}
		pkgs, err := packages.Load(cfg, "hostettler.dev/dnc/ui/screen")
		if err != nil {
			t.Fatalf("failed to load ui/screen: %v", err)
		}
		if packages.PrintErrors(pkgs) > 0 {
			t.Fatal("ui/screen has type errors")
		}

		scope := pkgs[0].Types.Scope()
		ifaceObj := scope.Lookup("FocusableModel")
		if ifaceObj == nil {
			t.Fatal("FocusableModel not found in ui/screen")
		}
		iface, ok := ifaceObj.Type().Underlying().(*types.Interface)
		if !ok {
			t.Fatalf("FocusableModel is not an interface, got %T", ifaceObj.Type().Underlying())
		}

		for _, name := range scope.Names() {
			if !strings.HasSuffix(name, "Screen") {
				continue
			}
			obj, ok := scope.Lookup(name).(*types.TypeName)
			if !ok {
				continue
			}
			if _, ok := obj.Type().Underlying().(*types.Struct); !ok {
				continue
			}
			// Screens are used as pointers (dncapp.go holds *Screen values),
			// so check the pointer method set.
			if ptr := types.NewPointer(obj.Type()); !types.Implements(ptr, iface) {
				t.Errorf("*%s does not implement FocusableModel", name)
			}
		}
	})
}
