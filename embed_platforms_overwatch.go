//go:build overwatch

package main

import (
	"TcNo-Acc-Switcher/internal/platform"

	_ "embed"
)

// The Overwatch build ships only Battle.net and Steam - the two platforms its
// single account list can show - on every OS, so there is no per-OS split
// here the way the full catalog has one.
//
//go:embed Platforms.overwatch.json
var overwatchPlatformsJSON []byte

func init() {
	platform.SetEmbeddedPlatformsJSON(overwatchPlatformsJSON)
}
