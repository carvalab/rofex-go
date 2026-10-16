package rofex

import (
	"runtime/debug"
	"sync"
)

const modulePath = "github.com/carvalab/rofex-go"

// Version returns the version of this module as recorded by the Go toolchain
// in the binary that imports it (e.g. "v0.3.4"), or "devel" when built from a
// local checkout. It never needs to be updated by hand.
func Version() string { return buildVersion() }

var buildVersion = sync.OnceValue(func() string {
	bi, _ := debug.ReadBuildInfo()
	return versionFrom(bi)
})

func versionFrom(bi *debug.BuildInfo) string {
	if bi == nil {
		return "devel"
	}
	mod := &bi.Main
	if mod.Path != modulePath {
		mod = nil
		for _, d := range bi.Deps {
			if d.Path == modulePath {
				mod = d
				break
			}
		}
	}
	if mod == nil {
		return "devel"
	}
	if mod.Replace != nil {
		mod = mod.Replace
	}
	if mod.Version == "" || mod.Version == "(devel)" {
		return "devel"
	}
	return mod.Version
}

func defaultUserAgent() string {
	return "rofex-go/" + Version() + " (+https://" + modulePath + ")"
}
