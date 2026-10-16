package rofex

import (
	"runtime/debug"
	"testing"
)

// The User-Agent must report the version the user actually depends on, so
// Primary/broker support can tell which SDK release sent a request.
func TestVersionFrom(t *testing.T) {
	dep := func(v string, replace *debug.Module) *debug.Module {
		return &debug.Module{Path: modulePath, Version: v, Replace: replace}
	}
	tests := []struct {
		name string
		bi   *debug.BuildInfo
		want string
	}{
		{"no build info", nil, "devel"},
		{"imported as dependency", &debug.BuildInfo{
			Main: debug.Module{Path: "example.com/bot"},
			Deps: []*debug.Module{{Path: "golang.org/x/text", Version: "v0.43.0"}, dep("v0.3.4", nil)},
		}, "v0.3.4"},
		{"dependency replaced by fork", &debug.BuildInfo{
			Main: debug.Module{Path: "example.com/bot"},
			Deps: []*debug.Module{dep("v0.3.4", &debug.Module{Path: "github.com/me/rofex-go", Version: "v0.3.5-fix"})},
		}, "v0.3.5-fix"},
		{"replaced by local directory", &debug.BuildInfo{
			Main: debug.Module{Path: "example.com/bot"},
			Deps: []*debug.Module{dep("v0.3.4", &debug.Module{Path: "../rofex-go"})},
		}, "devel"},
		{"local checkout", &debug.BuildInfo{Main: debug.Module{Path: modulePath, Version: "(devel)"}}, "devel"},
		{"not linked", &debug.BuildInfo{Main: debug.Module{Path: "example.com/bot"}}, "devel"},
	}
	for _, tt := range tests {
		if got := versionFrom(tt.bi); got != tt.want {
			t.Errorf("%s: got %q, want %q", tt.name, got, tt.want)
		}
	}
}
